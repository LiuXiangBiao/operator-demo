package controllers

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"mytool/api/v1beta1"
)

func NewService(app *v1beta1.Automon) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,

			OwnerReferences: makeOwnerReference(app),
		},
		Spec: v1.ServiceSpec{
			ClusterIP: v1.ClusterIPNone,
			Type:      v1.ServiceTypeClusterIP,
			Ports:     app.Spec.Port,
			Selector: map[string]string{
				"app": app.Name,
			},
		},
	}
}

func makeOwnerReference(app *v1beta1.Automon) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(app, schema.GroupVersionKind{
			Group:   v1beta1.GroupVersion.Group,
			Version: v1beta1.GroupVersion.Version,
			Kind:    v1beta1.Kind,
		}),
	}
}
