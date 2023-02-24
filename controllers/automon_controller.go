/*
Copyright 2023 liuxiangbiao.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/util/retry"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	toolv1beta1 "mytool/api/v1beta1"
)

var (
	oldSpecAnnotation = "old/spec"
)

// AutomonReconciler reconciles a Automon object
type AutomonReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tool.liuxiangbiao.com,resources=automons,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tool.liuxiangbiao.com,resources=automons/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tool.liuxiangbiao.com,resources=automons/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Automon object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AutomonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	ctb := context.Background()

	//先获取 mytool 实例
	var automon toolv1beta1.Automon
	err := r.Client.Get(ctb, req.NamespacedName, &automon)
	if err != nil {
		if err := client.IgnoreNotFound(err); err != nil {
			return ctrl.Result{}, err
		}
		// 删除一个不存在对象的时候，可能会报not-found错误，这种情况不需要重新入队列排队修复
		return ctrl.Result{}, nil
	}

	// 当前对象标记删除时
	if automon.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	// 如果关联的资源不存在，那么就去创建
	// 存在的话判断是否需要更新
	deploy := &v1.Deployment{}
	if err := r.Client.Get(ctb, req.NamespacedName, deploy); err != nil && errors.IsNotFound(err) {
		// 关联annotations
		annoData, _ := json.Marshal(automon.Spec)
		if automon.Annotations != nil {
			automon.Annotations[oldSpecAnnotation] = string(annoData)
		} else {
			automon.Annotations = map[string]string{oldSpecAnnotation: string(annoData)}
		}
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Client.Update(ctb, &automon); err != nil {
				return err
			}
			if err := Notice(&automon.Kind, &automon.Name); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return ctrl.Result{}, err
		}

		// deployment不存在创建资源
		deploy := NewDeploy(&automon)
		if err := r.Client.Create(ctb, deploy); err != nil {
			return ctrl.Result{}, err
		}
		// 3. 创建 Service
		service := NewService(&automon)
		if err := r.Create(ctb, service); err != nil {
			return ctrl.Result{}, err
		}
		// 4、 创建 Ingress
		ingress := NewIngress(&automon)
		if err := r.Client.Create(ctb, ingress); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// todo 更新  先判断是否需要更新 （yaml文件是否变化  old yaml 可以从annnotations获取）
	oldSpec := toolv1beta1.AutomonSpec{}
	if err := json.Unmarshal([]byte(automon.Annotations[oldSpecAnnotation]), &oldSpec); err != nil {
		return ctrl.Result{}, err
	}

	// 更新deployment 新旧对象进行比较，不一样就更新
	if !reflect.DeepEqual(automon.Spec, oldSpec) {
		// 应该更新资源了
		newDeploy := NewDeploy(&automon)
		oldDeploy := &v1.Deployment{}
		if err := r.Client.Get(ctb, req.NamespacedName, oldDeploy); err != nil {
			return ctrl.Result{}, err
		}
		oldDeploy.Spec = newDeploy.Spec
		// 正常直接更新oldDeploy,但一般不会直接调用 update 更新
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Client.Update(ctb, oldDeploy); err != nil {
				return err
			}
			if err := Notice(&oldDeploy.Kind, &oldDeploy.Name); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return ctrl.Result{}, err
		}

		// 更新 Service
		newService := NewService(&automon)
		oldService := &v12.Service{}
		if err := r.Client.Get(ctb, req.NamespacedName, oldService); err != nil {
			return ctrl.Result{}, err
		}
		newService.Spec.ClusterIP = oldService.Spec.ClusterIP
		oldService.Spec = newService.Spec
		// 正常直接更新oldDeploy,但一般不会直接调用 update 更新
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Client.Update(ctb, oldService); err != nil {
				return err
			}
			if err := Notice(&oldService.Kind, &oldService.Name); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return ctrl.Result{}, err
		}

		// 更新 Ingress
		newIngress := NewIngress(&automon)
		oldIngress := &v1beta1.Ingress{}
		if err := r.Client.Get(ctb, req.NamespacedName, oldIngress); err != nil {
			return ctrl.Result{}, err
		}

		oldIngress.Spec = newIngress.Spec

		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Client.Update(ctb, oldIngress); err != nil {
				return err
			}
			if err := Notice(&oldIngress.Kind, &oldIngress.Name); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutomonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toolv1beta1.Automon{}).
		Complete(r)
}
