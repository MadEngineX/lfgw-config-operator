/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ACLSpec defines the desired state of ACL
type ACLSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Rules []Rule `json:"rules"`
}

type Rule struct {
	RoleName string `json:"roleName"`

	// +kubebuilder:validation:Pattern="^([a-z0-9.*-]*)$"
	NamespaceFilter string `json:"namespaceFilter"`
}

// ACLStatus defines the observed state of ACL
type ACLStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum=SYNCED;ERROR
	Code string `json:"code"`

	Message string `json:"message"`
}

// +kubebuilder:printcolumn:name="Status",description="Current resource status",type=string,JSONPath=`.status.code`
// +kubebuilder:printcolumn:name="Message",description="Event message from operator",type=string,JSONPath=`.status.message`
// +kubebuilder:printcolumn:name="Timestamp",description="Creation date",type=string,JSONPath=`.metadata.creationTimestamp`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ACL is the Schema for the acls API
type ACL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ACLSpec   `json:"spec,omitempty"`
	Status ACLStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ACLList contains a list of ACL
type ACLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ACL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ACL{}, &ACLList{})
}
