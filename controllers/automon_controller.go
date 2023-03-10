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

	//????????? mytool ??????
	var automon toolv1beta1.Automon
	err := r.Client.Get(ctb, req.NamespacedName, &automon)
	if err != nil {
		if err := client.IgnoreNotFound(err); err != nil {
			return ctrl.Result{}, err
		}
		// ???????????????????????????????????????????????????not-found?????????????????????????????????????????????????????????
		return ctrl.Result{}, nil
	}

	// ???????????????????????????
	if automon.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	// ???????????????????????????????????????????????????
	// ????????????????????????????????????
	deploy := &v1.Deployment{}
	if err := r.Client.Get(ctb, req.NamespacedName, deploy); err != nil && errors.IsNotFound(err) {
		// ??????annotations
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

		// deployment?????????????????????
		deploy := NewDeploy(&automon)
		if err := r.Client.Create(ctb, deploy); err != nil {
			return ctrl.Result{}, err
		}
		// 3. ?????? Service
		service := NewService(&automon)
		if err := r.Create(ctb, service); err != nil {
			return ctrl.Result{}, err
		}
		// 4??? ?????? Ingress
		ingress := NewIngress(&automon)
		if err := r.Client.Create(ctb, ingress); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// todo ??????  ??????????????????????????? ???yaml??????????????????  old yaml ?????????annnotations?????????
	oldSpec := toolv1beta1.AutomonSpec{}
	if err := json.Unmarshal([]byte(automon.Annotations[oldSpecAnnotation]), &oldSpec); err != nil {
		return ctrl.Result{}, err
	}

	// ??????deployment ?????????????????????????????????????????????
	if !reflect.DeepEqual(automon.Spec, oldSpec) {
		// ?????????????????????
		newDeploy := NewDeploy(&automon)
		oldDeploy := &v1.Deployment{}
		if err := r.Client.Get(ctb, req.NamespacedName, oldDeploy); err != nil {
			return ctrl.Result{}, err
		}
		oldDeploy.Spec = newDeploy.Spec
		// ??????????????????oldDeploy,??????????????????????????? update ??????
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

		// ?????? Service
		newService := NewService(&automon)
		oldService := &v12.Service{}
		if err := r.Client.Get(ctb, req.NamespacedName, oldService); err != nil {
			return ctrl.Result{}, err
		}
		newService.Spec.ClusterIP = oldService.Spec.ClusterIP
		oldService.Spec = newService.Spec
		// ??????????????????oldDeploy,??????????????????????????? update ??????
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

		// ?????? Ingress
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
