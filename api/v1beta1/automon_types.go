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

package v1beta1

import (
	"k8s.io/api/apps/v1beta1"
	v1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AutomonSpec defines the desired state of Automon
type AutomonSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Quantity *int32                 `json:"quantity"`
	Image    string                 `json:"image"`
	Port     []v1.ServicePort       `json:"port"`
	Rule     []v1beta12.IngressRule `json:"rule"`
	Env      []v1.EnvVar            `json:"env,omitempty"`
}

// AutomonStatus defines the observed state of Automon
type AutomonStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	v1beta1.DeploymentStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Automon is the Schema for the automons API
type Automon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutomonSpec   `json:"spec,omitempty"`
	Status AutomonStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AutomonList contains a list of Automon
type AutomonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Automon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Automon{}, &AutomonList{})
}
