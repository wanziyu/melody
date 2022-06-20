/*
Copyright 2022. The Melody Authors

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
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InferenceSpec defines the desired state of Inference
type InferenceSpec struct {

	//Domain specify the domain of inference, such as nlp
	Domain DomainType `json:"domain,omitempty"`

	// Replicas specify the expected model serving replicas.
	Replicas *int32 `json:"replicas,omitempty"`

	// PredictorStatuses exposes current observed status for each predictor.
	Servings []ServingSpec `json:"servings"`

	// The request template in json format, used for testing against the REST API of target service.
	RequestTemplate string `json:"requestTemplate,omitempty"`

	// The client template to trigger the test against target service.
	ClientTemplate v1beta1.JobTemplateSpec `json:"clientTemplate,omitempty"`
}

type ServingSpec struct {
	//Name indicates the serving name.
	Name string `json:"name,omitempty"`

	//Image indicates the serving instance image
	Image string `json:"image,omitempty"`

	//Node indicates the node of serving instance
	Node string `json:"node,omitempty"`

	//ModelPath is the loaded madel filepath in model storage.
	ModelPath *string `json:"modelPath,omitempty"`

	//ModelVersion specifies the name of target model version to be loaded.
	ModelVersion string `json:"modelVersion,omitempty"`
}

// InferenceStatus defines the observed state of Inference
type InferenceStatus struct {
	// Output of the trial, including the TrialAssignment and the Objective value (e.g., QPS)
	MonitorResult *MonitoringResult `json:"trialResult,omitempty"`

	// Observed runtime condition for this Inference.
	Conditions []ServingConditions `json:"servingStatuses,omitempty"`

	// The time this inference job was started.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// The time this inference job was completed.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

type ServingConditions struct {

	// Name is the name of current predictor.
	Name string `json:"name"`

	// Replicas is the expected replicas of current predictor.
	Replicas int32 `json:"replicas"`

	//ReadyReplicas is the ready replicas of current predictor.
	ReadyReplicas int32 `json:"readyReplicas"`

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Standard Kubernetes object's LastTransitionTime
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	//Status Type of Serving,
	Type ServingStatusType `json:"type"`

	// A description message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

type DomainType string

const (
	ImageProcessing      DomainType = "image-processing"
	TimeSeriesPrediction DomainType = "time-series-prediction"
	SignalProcessing     DomainType = "signal-processing"
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

func init() {
	SchemeBuilder.Register(&Inference{}, &InferenceList{})
}
