package inference

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	melodyiov1alpha1 "melody/api/v1alpha1"
	"melody/pkg/controllers/consts"
	"melody/pkg/controllers/util"
)

type updateStatusFunc func(instance *melodyiov1alpha1.Inference) error

func (r *InferenceReconciler) UpdateInferenceStatusByClientJob(instance *melodyiov1alpha1.Inference, deployedJob *batchv1.Job) error {
	logger := log.WithValues("Inference", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// Update inference result
	if err := r.updateInferenceResult(instance, deployedJob); err != nil {
		logger.Error(err, "Update inference result error")
		return err
	}
	// Update inference condition
	jobCondition := deployedJob.Status.Conditions
	r.updateInferenceStatusCondition(instance, deployedJob, jobCondition)
	return nil
}

func (r *InferenceReconciler) UpdateInferenceStatusByServiceDeployment(instance *melodyiov1alpha1.Inference, deployedDeployment *appsv1.Deployment) {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	ServiceDeploymentCondition := deployedDeployment.Status.Conditions
	if util.IsServiceDeplomentFail(ServiceDeploymentCondition) {
		message := "Trial service pod failed"
		objectiveMetricName := instance.Spec.Objective.ObjectiveMetricName
		metric := morphlingv1alpha1.Metric{Name: objectiveMetricName, Value: "0.0"}
		instance.Status.TrialResult = &morphlingv1alpha1.TrialResult{}
		instance.Status.TrialResult.ObjectiveMetricsObserved = []morphlingv1alpha1.Metric{metric}
		util.MarkInferenceStatusFailed(instance, message)
		logger.Info("Service deployment is failed", "name", deployedDeployment.GetName())
	} else {
		message := "Trial service pod pending"
		util.MarkInferenceStatusPendingTrial(instance, message)
		logger.Info("Service deployment is pending", "name", deployedDeployment.GetName())
	}
}

//根据job来更新inference的condition
func (r *InferenceReconciler) updateInferenceStatusCondition(instance *melodyiov1alpha1.Inference, deployedJob *batchv1.Job, jobCondition []batchv1.JobCondition) {

	if jobCondition == nil || instance == nil || deployedJob == nil {
		msg := "Trial is running"
		util.MarkInferenceStatusRunning(instance, msg)
		return
	}

	now := metav1.Now()
	if util.IsJobSucceeded(jobCondition) {
		// Client-side stress test job is completed
		if isInferenceResultAvailable(instance) {
			msg := "Client-side stress test job has completed"
			util.MarkInferenceStatusSucceeded(instance, corev1.ConditionTrue, msg)
			instance.Status.CompletionTime = &now
			eventMsg := fmt.Sprintf("Client-side stress test job %s has succeeded", deployedJob.GetName())
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "JobSucceeded", eventMsg)
		} else {
			// Client job has NOT recorded the trial result
			msg := "Trial results are not available"
			util.MarkInferenceStatusSucceeded(instance, corev1.ConditionFalse, msg)
		}
	} else if util.IsJobFailed(jobCondition) {
		// Client-side stress test job is failed
		msg := "Client-side stress test job has failed"
		util.MarkInferenceStatusFailed(instance, msg)
		instance.Status.CompletionTime = &now
	} else {
		// Client-side stress test job is still running
		msg := "Client-side stress test job is running"
		util.MarkInferenceStatusRunning(instance, msg)
	}
}

func (r *InferenceReconciler) updateInferenceResult(instance *melodyiov1alpha1.Inference, deployedJob *batchv1.Job) error {
	logger := log.WithValues("Inference", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	jobCondition := deployedJob.Status.Conditions
	if util.IsJobSucceeded(jobCondition) {
		logger.Info("Client Job is Completed", "name", deployedJob.GetName())
		// Update inference observation
		if err := r.updateInferenceResultForSucceededInference(instance); err != nil {
			logger.Error(err, "Update inference result error")
			return err
		}
	} else if util.IsJobFailed(jobCondition) {
		logger.Info("Client Job is Failed", "name", deployedJob.GetName())
		r.updateInferenceResultForFailedInference(instance)
	}
	return nil
}

func (r *InferenceReconciler) updateInferenceResultForSucceededInference(instance *melodyiov1alpha1.Inference) error {
	if &instance.Spec.ClientTemplate == nil || &instance.Spec.Domain == nil || r.DBClient == nil {
		return nil
	}
	reply, err := r.GetMonitorResult(instance)
	if err != nil {
		return err
	}
	if reply != nil {
		instance.Status.MonitorResult = reply
	}
	return nil
}

func (r *InferenceReconciler) updateInferenceResultForFailedInference(instance *melodyiov1alpha1.Inference) {
	instance.Status.MonitorResult = &melodyiov1alpha1.MonitoringResult{
		MonitoringInferences: nil,
		MonitoringNodes:      nil,
	}

	instance.Status.TrialResult.TunableParameters = make([]morphlingv1alpha1.ParameterAssignment, 0)
	for _, assignment := range instance.Spec.SamplingResult {
		instance.Status.TrialResult.TunableParameters = append(instance.Status.TrialResult.TunableParameters, morphlingv1alpha1.ParameterAssignment{
			Name:     assignment.Name,
			Value:    assignment.Value,
			Category: assignment.Category,
		})
	}
	instance.Status.TrialResult.ObjectiveMetricsObserved = append(instance.Status.TrialResult.ObjectiveMetricsObserved, morphlingv1alpha1.Metric{
		Name:  instance.Spec.Objective.ObjectiveMetricName,
		Value: consts.DefaultMetricValue,
	})
}

func isInferenceResultAvailable(instance *melodyiov1alpha1.Inference) bool {
	if instance == nil || &instance.Spec.Objective == nil || &instance.Spec.Objective.ObjectiveMetricName == nil {
		return false
	}
	// Get the name of the objective metric
	objectiveMetricName := instance.Spec.Objective.ObjectiveMetricName
	if instance.Status.TrialResult != nil {
		if instance.Status.TrialResult.ObjectiveMetricsObserved != nil {
			for _, metric := range instance.Status.TrialResult.ObjectiveMetricsObserved {
				// Find the objective metric record from trail status
				if metric.Name == objectiveMetricName {
					return true
				}
			}
		}
	}
	// Objective metric record Not found
	return false
}

func (r *InferenceReconciler) ControllerName() string {
	return ControllerName
}
