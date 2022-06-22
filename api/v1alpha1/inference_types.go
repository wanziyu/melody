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

	//Metrics MonitoringMetric `json:"metrics,omitempty"`
	//Node indicates the node of serving instance
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// PredictorStatuses exposes current observed status for each predictor.
	ServingTemplate corev1.PodTemplate `json:"servingTemplate"`

	// The request template in json format, used for testing against the REST API of target service.
	RequestTemplate string `json:"requestTemplate,omitempty"`

	// The client template to trigger a prometheus monitoring client.
	ClientTemplate v1beta1.JobTemplateSpec `json:"clientTemplate,omitempty"`

	// The maximum time in seconds for a deployment to make progress before it is considered to be failed.
	ServiceProgressDeadline *int32 `json:"serviceProgressDeadline,omitempty"`
}

type MonitoringMetric struct {
	MetricName string `json:"metricName,omitempty"`
}

type ServingSpec struct {
	//Name indicates the serving name.
	Name string `json:"name,omitempty"`

	//Image indicates the serving instance image
	// +required
	ImageRepo string `json:"imageRepo,omitempty"`

	//Node indicates the node of serving instance
	// +optional
	Node string `json:"node,omitempty"`

	// Storage is the location where this ModelVersion is stored.
	// +optional
	Storage *Storage `json:"storage,omitempty"`

	//ModelPath is the loaded madel filepath in model storage.
	// +optional
	ModelPath *string `json:"modelPath,omitempty"`

	//ModelVersion specifies the name of target model version to be loaded.
	// +optional
	ModelVersion string `json:"modelVersion,omitempty"`
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

type Storage struct {
	// NFS represents the alibaba cloud nas storage
	// +optional
	NFS *NFS `json:"nfs,omitempty"`

	// LocalStorage represents the local host storage
	// +optional
	LocalStorage *LocalStorage `json:"localStorage,omitempty"`
}

type NFS struct {
	// NFS server address, e.g. "***.cn-beijing.nas.aliyuncs.com"
	Server string `json:"server,omitempty"`

	// The path under which the model is stored, e.g. /models/my_model1
	Path string `json:"path,omitempty"`

	// The mounted path inside the container.
	// The training code is expected to export the model artifacts under this path, such as storing the tensorflow saved_model.
	MountPath string `json:"mountPath,omitempty"`
}

type LocalStorage struct {
	// The local host path to export the model.
	// +required
	Path string `json:"path,omitempty"`

	// The mounted path inside the container.
	// The training code is expected to export the model artifacts under this path, such as storing the tensorflow saved_model.
	MountPath string `json:"mountPath,omitempty"`

	// The name of the node for storing the model. This node will be where the chief worker run to export the model.
	// +required
	NodeName string `json:"nodeName,omitempty"`
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
