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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SchedulingDecesionSpec defines the desired state of SchedulingDecesion
type SchedulingDecesionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Algorithm specifies the scheduling algorithm (i.e. DQN) for serving tasks.
	Algorithm *SchedulingAlgorithm `json:"algorithm,omitempty"`

	//SchedulingResult specifies
	Objective SchedulingObjective `json:"schedulingResult,omitempty"`

	ResultTime metav1.Time `json:"resultTime"`
}

type SchedulingAlgorithm string

const (
	DQNScheduling     SchedulingAlgorithm = "DQN"
	DefaultScheduling SchedulingAlgorithm = "default"
)

type SchedulingObjective struct {
	Type           SchedulingType `json:"type"`
	TargetPod      corev1.Pod     `json:"targetPod"`
	TargetNode     corev1.Node    `json:"targetNode"`
	ScalingReplica int32          `json:"scalingReplica"`
}

// SchedulingDecesionStatus defines the observed state of SchedulingDecesion
type SchedulingDecesionStatus struct {

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	//Is this sd is used to scheduling.
	Used bool `json:"used"`

	//The time SchedulingDecesion has been completed.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SchedulingDecesion is the Schema for the schedulingdecesions API
type SchedulingDecesion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SchedulingDecesionSpec   `json:"spec,omitempty"`
	Status SchedulingDecesionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SchedulingDecesionList contains a list of SchedulingDecesion
type SchedulingDecesionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SchedulingDecesion `json:"items"`
}

type SchedulingType string

const (
	Transition        SchedulingType = "Transition"
	Scaling           SchedulingType = "Scaling"
	TransitionScaling SchedulingType = "TransitionScaling"
)

func init() {
	SchemeBuilder.Register(&SchedulingDecesion{}, &SchedulingDecesionList{})
}
