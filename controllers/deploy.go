package controllers

import (
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"mytool/api/v1beta1"
)

func NewDeploy(app *v1beta1.Automon) *v1.Deployment {
	labels := map[string]string{"app": app.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	return &v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,

			OwnerReferences: makeOwnerReference(app),
		},
		Spec: v1.DeploymentSpec{
			Replicas: app.Spec.Quantity,
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v12.PodSpec{
					Containers: newContainers(app),
				},
			},
			Selector: selector,
		},
	}
}

func newContainers(app *v1beta1.Automon) []v12.Container {
	containerPorts := []v12.ContainerPort{}
	for _, svcPort := range app.Spec.Port {
		cport := v12.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal
		containerPorts = append(containerPorts, cport)
	}
	return []v12.Container{
		{
			Name:            app.Name,
			Image:           app.Spec.Image,
			Ports:           containerPorts,
			ImagePullPolicy: v12.PullIfNotPresent,
			Env:             app.Spec.Env,
		},
	}
}
