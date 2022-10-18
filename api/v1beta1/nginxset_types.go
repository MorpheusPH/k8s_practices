/*
Copyright 2022.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const NginxSetFinalizer = "finalizers.nginxset.io"

type Condition struct {
	// type of condition
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// status of the condition, one of True, False, Unknown.
	// +required
	Status metav1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// reason contains a programmatic identifier indicating the reason for the condition's last transition.
	// Producers of specific condition types may define expected values and meanings for this field,
	// and whether the values are considered a guaranteed API.
	// The value should be a CamelCase string.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// message is a human readable message indicating details about the transition.
	// This may be an empty string.
	// +optional
	// +kubebuilder:validation:MaxLength=32768
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
	// Last time the condition was probed
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,3,opt,name=lastUpdateTime"`
	// lastTransitionTime is the last time the condition transitioned from one status to another.
	// This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
}

// NginxSetSpec defines the desired state of NginxSet
type NginxSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +required
	Replicas *int32 `json:"replicas"`
}

// NginxSetStatus defines the observed state of NginxSet
type NginxSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +optional
	ConditionStatus `json:",inline"`
	// +optional
	FailureStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NginxSet is the Schema for the nginxsets API
type NginxSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NginxSetSpec   `json:"spec,omitempty"`
	Status NginxSetStatus `json:"status,omitempty"`
}

func (app *NginxSet) SetConditions(conditionType string, status metav1.ConditionStatus, reason string, message string) {
	app.Status.SetConditions(conditionType, status, reason, message)
}

func (app *NginxSet) FindCondition(conditionType string) *Condition {
	return app.Status.FindCondition(conditionType)
}

func (app *NginxSet) AddFailureCounts() {
	app.Status.AddFailureCounts()
}

func (app *NginxSet) ResetFailureCounts() {
	app.Status.ResetFailureCounts()
}

//+kubebuilder:object:root=true

// NginxSetList contains a list of NginxSet
type NginxSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NginxSet{}, &NginxSetList{})
}
