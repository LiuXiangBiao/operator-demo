package controllers

import (
	"github.com/alibabacloud-go/tea/tea"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	v1beta13 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"mytool/api/v1beta1"
)

func NewIngress(app *v1beta1.Automon) *v1beta12.Ingress {
	return &v1beta12.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Annotations: map[string]string{
				v1beta13.AnnotationIsDefaultIngressClass: "true",
				v1beta13.AnnotationIngressClass:          "nginx",
			},

			OwnerReferences: makeOwnerReference(app),
		},
		Spec: v1beta12.IngressSpec{
			IngressClassName: tea.String(app.Name),
			Rules:            app.Spec.Rule,
		}}

}
