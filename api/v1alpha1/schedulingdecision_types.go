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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SchedulingDecisionSpec defines the desired state of SchedulingDecision
type SchedulingDecisionSpec struct {
	// Scheduling algorithm, e.g., A2C.
	Algorithm AlgorithmSpec `json:"algorithm,omitempty"`

	//Scheduling decisions obtained from algorithm server to update inference
	Decision SchedulingActionSpec `json:"decision,omitempty"`

	// Maximum number of inferences
	MaxNumInferences *int32 `json:"maxNumInferences,omitempty"`

	// Parallelism is the number of concurrent inferences.
	Parallelism *int32 `json:"parallelism,omitempty"`

	// The request template in json format, used for testing against the REST API of target service.
	RequestTemplate string `json:"requestTemplate,omitempty"`

	// The target service inference needed to be better scheduled
	ServiceInferenceTemplate []InferenceSpec `json:"servicePodTemplate,omitempty"`

	// The maximum time in seconds for a deployment to make progress before it is considered to be failed.
	ServiceProgressDeadline *int32 `json:"serviceProgressDeadline,omitempty"`
}

type SchedulingActionSpec struct {
	// Obtained time of the scheduling action
	ObtainedTime *metav1.Time `json:"obtainedTime,omitempty"`

	// Obtained scheduling action from algorithm server
	Actions []ActionSpec `json:"actions,omitempty"`
}

type ActionSpec struct {
	Type       ActionType `json:"type,omitempty"`
	TargetNode string     `json:"targetNode,omitempty"`
	Value      string     `json:"value,omitempty"`
}

// SchedulingDecisionStatus defines the observed state of SchedulingDecision
type SchedulingDecisionStatus struct {
	// List of observed runtime conditions for this SchedulingDecision.
	Conditions []SchedulingCondition `json:"conditions,omitempty"`

	// Current monitoring results
	CurrentMonitoring []MonitoringResult `json:"currentMonitoring,omitempty"`

	// Completion time of the scheduling
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Start time of the scheduling
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// List of inference names which are running.
	RunningInferenceList []string `json:"runningInferenceList,omitempty"`

	// List of inference names which are pending.
	PendingInferenceList []string `json:"pendingInferenceList,omitempty"`

	// List of inference names which have already failed.
	FailedInferenceList []string `json:"failedInferenceList,omitempty"`

	// List of inference names which have already succeeded.
	SucceededInferenceList []string `json:"succeededInferenceList,omitempty"`

	// List of inference names which have been killed.
	KilledInferenceList []string `json:"killedInferenceList,omitempty"`

	// InferencesTotal is the total number of inference owned by the experiment.
	InferencesTotal int32 `json:"inferencesTotal,omitempty"`

	// How many inferences have succeeded.
	InferencesSucceeded int32 `json:"inferencesSucceeded,omitempty"`

	// How many inferences have been killed.
	InferencesKilled int32 `json:"inferencesKilled,omitempty"`

	// How many inferences are pending.
	InferencesPending int32 `json:"inferencesPending,omitempty"`

	// How many inferences are running.
	InferencesRunning int32 `json:"inferencesRunning,omitempty"`

	// How many inferences have failed.
	InferencesFailed int32 `json:"inferencesFailed,omitempty"`
}

// SchedulingCondition describes the state of the experiment at a certain point.
type SchedulingCondition struct {
	// Type of experiment condition.
	Type SchedulingConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// A description message indicating details about the transition.
	Message string `json:"message,omitempty"`

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

type MonitoringResult struct {
	MonitoringInferences []InferencePodStatus `json:"monitoringPods,omitempty"`
	MonitoringNodes      []NodeStatus         `json:"monitoringNodes,omitempty"`
}

type InferencePodStatus struct {
	PodName string          `json:"podName,omitempty"`
	Metrics []PodMetricSpec `json:"metrics,omitempty"`
}

type PodMetricSpec struct {
	Name     string   `json:"name,omitempty"`
	Value    string   `json:"value,omitempty"`
	Category Category `json:"category,omitempty"`
}

type NodeStatus struct {
	NodeName string           `json:"nodeName,omitempty"`
	Metrics  []NodeMetricSpec `json:"metrics,omitempty"`
}

type NodeMetricSpec struct {
	Category Category `json:"category,omitempty"`
	Value    string   `json:"value,omitempty"`
}

type SchedulingResult struct {
	// The scheduling target inference instance
	TargetPod string `json:"targetPod,omitempty"`
	// The target scheduling edge node for target pod
	NodeName string `json:"nodeName,omitempty"`
}

// SchedulingConditionType defines the status of the SchedulingDecision
type SchedulingConditionType string

const (
	SchedulingCreated    SchedulingConditionType = "Created"
	SchedulingRunning    SchedulingConditionType = "Running"
	SchedulingRestarting SchedulingConditionType = "Restarting"
	SchedulingSucceeded  SchedulingConditionType = "Succeeded"
	SchedulingFailed     SchedulingConditionType = "Failed"
	SchedulingCompleted  SchedulingConditionType = "Completed"
)

type ActionType string

const (
	NodeTransferAction  ActionType = "NodeTransfer"
	AddResourceAction   ActionType = "AddResource"
	MinusResourceAction ActionType = "MinusResource"
)

// AlgorithmName is the supported searching algorithms
type AlgorithmName string

const (
	RLScheduling  AlgorithmName = "RLScheduling"
	DQNScheduling AlgorithmName = "DQNScheduling"
	GridSearch    AlgorithmName = "grid"
)

// Category of the status to be monitored,
type Category string

const (
	// Computing resources, including cpu, memory, gpumem.
	CPUResource Category = "cpu"

	MemResource Category = "memory"

	FinishedCount Category = "count"

	// Environment variables, set for service pods/deployments.
	CategoryEnv Category = "env"

	// Args for codes running in service pods/deployments.
	CategoryArgs Category = "args"
)

// AlgorithmSpec is the specification of Opt. algorithm
type AlgorithmSpec struct {
	// The name of algorithm for sampling_client: random, grid, bayesian optimization.
	AlgorithmName AlgorithmName `json:"algorithmName,omitempty"`

	// The key-value pairs representing settings for sampling_client algorithms.
	AlgorithmSettings []AlgorithmSetting `json:"algorithmSettings,omitempty"`
}

// AlgorithmSetting defines the parameters key-value pair of the scheduling algorithm
type AlgorithmSetting struct {
	// The name of the key-value pair.
	Name string `json:"name,omitempty"`

	// The value of the key.
	Value string `json:"value,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[-1:].type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Optimal-Scheduling-Pod",type=string,JSONPath=`.status.currentOptimalScheduling.targetPod`
// +kubebuilder:printcolumn:name="Optimal-Scheduling-Node",type=string,JSONPath=`.status.currentOptimalScheduling.nodeName`
// +kubebuilder:resource:shortName="sd"
// +kubebuilder:subresource:status

// SchedulingDecision is the Schema for the schedulingdecisions API
type SchedulingDecision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SchedulingDecisionSpec   `json:"spec,omitempty"`
	Status SchedulingDecisionStatus `json:"status,omitempty"`
}

// SchedulingDecisionList contains a list of SchedulingDecision
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SchedulingDecisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SchedulingDecision `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SchedulingDecision{}, &SchedulingDecisionList{})
}
