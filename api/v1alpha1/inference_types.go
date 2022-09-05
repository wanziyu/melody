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
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InferenceSpec defines the desired state of Inference
type InferenceSpec struct {
	// Domain specify the domain of inference, such as nlp
	Domain DomainType `json:"domain,omitempty"`

	OptionalNodes []string `json:"optionalNodes,omitempty"`

	// The request template in json format, used for testing against the REST API of target service.
	RequestTemplate string `json:"requestTemplate,omitempty"`

	// The client template to trigger a prometheus monitoring client.
	ClientTemplate v1.JobTemplateSpec `json:"clientTemplate,omitempty"`

	// The target service pod/deployment whose parameters to be tuned
	ServingTemplate corev1.PodTemplate `json:"servingTemplate,omitempty"`

	// The maximum time in seconds for a deployment to make progress before it is considered to be failed.
	ServiceProgressDeadline *int32 `json:"serviceProgressDeadline,omitempty"`
}

// InferenceStatus defines the observed state of Inference
type InferenceStatus struct {
	// Output of the Monitoring Client, including the inference running status and node status (e.g., CPU)
	MonitorResult *MonitoringResult `json:"monitorResult,omitempty"`

	// Observed runtime condition for this Inference.
	Conditions []InferenceCondition `json:"servingStatuses,omitempty"`

	// The time this inference job was started.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// The time this inference job was completed.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

// MonitoringResult describes pod and node metrics.
type MonitoringResult struct {
	// Monitoring of pod metrics
	PodMetrics PodMetricSpec `json:"podMetrics,omitempty"`

	// Monitoring of node metrics
	NodeMetrics []NodeMetricSpec `json:"nodeMetrics,omitempty"`
}

//PodMetricSpec describes pod  metrics.
type PodMetricSpec struct {
	//Name
	PodName string `json:"podName"`
	//Metrics
	Metrics []PodMetrics `json:"metrics"`
}

// PodMetrics describes pod  metrics.
type PodMetrics struct {
	Category MetricType `json:"category"`
	Value    string     `json:"value"`
}

// NodeMetricSpec describes node metrics.
type NodeMetricSpec struct {
	NodeName string      `json:"nodeName"`
	Metrics  NodeMetrics `json:"metrics"`
}

// NodeMetrics describes node metrics.
type NodeMetrics struct {
	Category MetricType `json:"category"`
	Value    string     `json:"value"`
}

// MetricType of the status to be monitored,
type MetricType string

const (
	CPUResource MetricType = "cpu"

	MemResource MetricType = "memory"

	CPUUsage MetricType = "cpuUsage"

	MemUsage MetricType = "memUsage"

	JobCompletionTime MetricType = "jct"
)

type InferenceCondition struct {

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

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[-1:].type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Monitor-PodName",type=string,JSONPath=`.status.monitoringResult.monitoringInferences[-1:].podName`
// +kubebuilder:printcolumn:name="Monitor-NodeName",type=string,JSONPath=`.status.monitoringResult.monitoringNodes[-1:].nodeName`
// +kubebuilder:subresource:status

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
