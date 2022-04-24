package utils

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"errors"
	melodyv1alpha1 "melody/api/v1alpha1"
)

// Inference related

func IsCreatedInference(inference *melodyv1alpha1.Inference) bool {
	return hasConditionInference(inference, melodyv1alpha1.ServingCreated)
}

func hasConditionInference(inference *melodyv1alpha1.Inference, condType melodyv1alpha1.ServingStatusType) bool {
	cond := getConditionInference(inference, condType)
	if cond != nil && cond.Status == v1.ConditionTrue {
		return true
	}
	return false
}

func getConditionInference(inference *melodyv1alpha1.Inference, condType melodyv1alpha1.ServingStatusType) *melodyv1alpha1.ServingStatus {
	for _, condition := range inference.Status.ServingStatuses {
		if condition.Type == condType {
			return &condition
		}
	}
	return nil
}

func newConditionInference(conditionType melodyv1alpha1.ServingStatusType, status v1.ConditionStatus, name string) melodyv1alpha1.ServingStatus {
	return melodyv1alpha1.ServingStatus{
		Name:           name,
		Type:           conditionType,
		Status:         status,
		LastUpdateTime: metav1.Now(),
	}
}

func SetConditionInference(inference *melodyv1alpha1.Inference, conditionType melodyv1alpha1.ServingStatusType, status v1.ConditionStatus, name string) {

	newCond := newConditionInference(conditionType, status, name)
	currentCond := getConditionInference(inference, conditionType)
	if currentCond != nil && currentCond.Status == newCond.Status {
		newCond.LastTransitionTime = currentCond.LastTransitionTime
	}
	removeConditionInference(inference, conditionType)
	inference.Status.ServingStatuses = append(inference.Status.ServingStatuses, newCond)
}

func removeConditionInference(inference *melodyv1alpha1.Inference, condType melodyv1alpha1.ServingStatusType) {
	var newConditions []melodyv1alpha1.ServingStatus
	for _, c := range inference.Status.ServingStatuses {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	inference.Status.ServingStatuses = newConditions
}

func MarkInferenceStatusCreatedInference(inference *melodyv1alpha1.Inference, message string) {
	SetConditionInference(inference, melodyv1alpha1.ServingCreated, v1.ConditionTrue, message)
}

func MarkInferenceStatusPendingTrial(inference *melodyv1alpha1.Inference, message string) {
	SetConditionInference(inference, melodyv1alpha1.ServingPending, v1.ConditionTrue, message)
}

func MarkInferenceStatusSucceeded(inference *melodyv1alpha1.Inference, status v1.ConditionStatus, message string) {
	currentCond := getConditionInference(inference, melodyv1alpha1.ServingRunning)
	if currentCond != nil {
		SetConditionInference(inference, melodyv1alpha1.ServingRunning, v1.ConditionFalse, currentCond.Message)
	}
	SetConditionInference(inference, melodyv1alpha1.ServingSucceeded, status, message)

}

func MarkInferenceStatusFailed(inference *melodyv1alpha1.Inference, message string) {
	currentCond := getConditionInference(inference, melodyv1alpha1.ServingRunning)
	if currentCond != nil {
		SetConditionInference(inference, melodyv1alpha1.ServingRunning, v1.ConditionFalse, currentCond.Message)
	}
	SetConditionInference(inference, melodyv1alpha1.ServingFailed, v1.ConditionTrue, message)
}

func MarkInferenceStatusRunning(inference *melodyv1alpha1.Inference, message string) {
	SetConditionInference(inference, melodyv1alpha1.ServingRunning, v1.ConditionTrue, message)
}

func GetLastConditionType(inference *melodyv1alpha1.Inference) (melodyv1alpha1.ServingStatusType, error) {
	if len(inference.Status.ServingStatuses) > 0 {
		return inference.Status.ServingStatuses[len(inference.Status.ServingStatuses)-1].Type, nil
	}
	return "", errors.New("Inference doesn't have any condition")
}

func IsJobSucceeded(jobCondition []batchv1.JobCondition) bool {
	for _, condition := range jobCondition {
		if condition.Type == batchv1.JobComplete {
			return true
		}
	}
	return false
}

func IsJobFailed(jobCondition []batchv1.JobCondition) bool {
	for _, condition := range jobCondition {
		if condition.Type == batchv1.JobFailed {
			return true
		}
	}
	return false
}

func IsServiceDeplomentReady(podConditions []appsv1.DeploymentCondition) bool {
	for _, condition := range podConditions {
		if condition.Type == appsv1.DeploymentAvailable && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsServiceDeplomentFail(podConditions []appsv1.DeploymentCondition) bool {
	for _, condition := range podConditions {
		if condition.Type == appsv1.DeploymentReplicaFailure && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsCompletedInference(inference *melodyv1alpha1.Inference) bool {
	return IsSucceededInference(inference) || IsFailedInference(inference)
}

func IsSucceededInference(inference *melodyv1alpha1.Inference) bool {
	return hasConditionInference(inference, melodyv1alpha1.ServingSucceeded)
}

func IsFailedInference(inference *melodyv1alpha1.Inference) bool {
	return hasConditionInference(inference, melodyv1alpha1.ServingFailed)
}

func IsRunningInference(inference *melodyv1alpha1.Inference) bool {
	return hasConditionInference(inference, melodyv1alpha1.ServingRunning)
}

func IsKilledInference(inference *melodyv1alpha1.Inference) bool {
	return hasConditionInference(inference, melodyv1alpha1.ServingKilled)
}

func IsPendingInference(inference *melodyv1alpha1.Inference) bool {
	return hasConditionInference(inference, melodyv1alpha1.ServingPending)
}
