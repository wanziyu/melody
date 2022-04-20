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

// InferenceSpec defines the desired state of Inference
type InferenceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Domain DomainType `json:"domain,omitempty"`
	// PredictorStatuses exposes current observed status for each predictor.
	Servings []ServingSpec `json:"servings"`
}

type ServingSpec struct {
	//Name indicates the serving name.
	Name string `json:"name,omitempty"`

	//ModelPath is the loaded madel filepath in model storage.
	ModelPath *string `json:"modelPath,omitempty"`

	//ModelVersion specifies the name of target model version to be loaded.
	ModelVersion string `json:"modelVersion,omitempty"`

	// Replicas specify the expected model serving  replicas.
	Replicas int32 `json:"replicas,omitempty"`

	BatchSize int32 `json:"batchSize,omitempty"`

	// Template describes a template of predictor pod with its properties.
	Template corev1.PodTemplateSpec `json:"template"`
}

// InferenceStatus defines the observed state of Inference
type InferenceStatus struct {

	// InferenceEndpoints exposes available serving service endpoint.
	InferenceEndpoint string `json:"inferenceEndpoint,omitempty"`

	// The time this inference job was started.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// The time this inference job was completed.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	ServingStatuses []ServingStatus `json:"servingStatuses,omitempty"`
}

type ServingStatus struct {

	// Name is the name of current predictor.
	Name string `json:"name"`
	// Replicas is the expected replicas of current predictor.
	Replicas int32 `json:"replicas"`
	// ReadyReplicas is the ready replicas of current predictor.
	ReadyReplicas int32 `json:"readyReplicas"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	Type ServingStatusType `json:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Inference is the Schema for the inferences API
type Inference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InferenceSpec   `json:"spec,omitempty"`
	Status InferenceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InferenceList contains a list of Inference
type InferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Inference `json:"items"`
}

type DomainType string

const (
	ImageProcessingDomain DomainType = "image"
	TimeSeriesDomain      DomainType = "timeseries"
)

type ServingStatusType string

const (
	ServingRunning   ServingStatusType = "Running"
	ServingSucceeded ServingStatusType = "Succeeded"
	ServingFailed    ServingStatusType = "Failed"
	ServingCreated   ServingStatusType = "Created"
	ServingPending   ServingStatusType = "Pending"
	ServingKilled    ServingStatusType = "Killed"
)

func init() {
	SchemeBuilder.Register(&Inference{}, &InferenceList{})
}
