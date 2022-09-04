package inference

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	melodyiov1alpha1 "melody/api/v1alpha1"
	"melody/controllers/const"
	"melody/controllers/util"
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
	logger := log.WithValues("Inference", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	ServiceDeploymentCondition := deployedDeployment.Status.Conditions
	if util.IsServiceDeplomentFail(ServiceDeploymentCondition) {
		message := "Inference service pod failed"
		cpu := melodyiov1alpha1.PodMetricSpec{PodName: instance.Name, Metrics: melodyiov1alpha1.PodMetrics{Category: melodyiov1alpha1.CPUResource, Value: "0.0"}}
		mem := melodyiov1alpha1.PodMetricSpec{PodName: instance.Name, Metrics: melodyiov1alpha1.PodMetrics{Category: melodyiov1alpha1.MemResource, Value: "0.0"}}
		jct := melodyiov1alpha1.PodMetricSpec{PodName: instance.Name, Metrics: melodyiov1alpha1.PodMetrics{Category: melodyiov1alpha1.JobCompletionTime, Value: "0.0"}}
		instance.Status.MonitorResult = &melodyiov1alpha1.MonitoringResult{}
		instance.Status.MonitorResult.PodMetrics = []melodyiov1alpha1.PodMetricSpec{cpu, mem, jct}

		nodes := instance.Spec.OptionalNodes
		instance.Status.MonitorResult.NodeMetrics = make([]melodyiov1alpha1.NodeMetricSpec, 0)
		for _, node := range nodes {
			cpu := melodyiov1alpha1.NodeMetricSpec{NodeName: node, Metrics: melodyiov1alpha1.NodeMetrics{Category: melodyiov1alpha1.CPUResource, Value: "0.0"}}
			mem := melodyiov1alpha1.NodeMetricSpec{NodeName: node, Metrics: melodyiov1alpha1.NodeMetrics{Category: melodyiov1alpha1.CPUResource, Value: "0.0"}}
			instance.Status.MonitorResult.NodeMetrics = append(instance.Status.MonitorResult.NodeMetrics, cpu)
			instance.Status.MonitorResult.NodeMetrics = append(instance.Status.MonitorResult.NodeMetrics, mem)
		}

		util.MarkInferenceStatusFailed(instance, message)
		logger.Info("Service deployment is failed", "name", deployedDeployment.GetName())
	} else {
		message := "Inference service pod pending"
		util.MarkInferenceStatusPendingTrial(instance, message)
		logger.Info("Service deployment is pending", "name", deployedDeployment.GetName())
	}
}

//根据job来更新inference的condition
func (r *InferenceReconciler) updateInferenceStatusCondition(instance *melodyiov1alpha1.Inference, deployedJob *batchv1.Job, jobCondition []batchv1.JobCondition) {

	if jobCondition == nil || instance == nil || deployedJob == nil {
		msg := "Inference is running"
		util.MarkInferenceStatusRunning(instance, msg)
		return
	}

	now := metav1.Now()
	if util.IsJobSucceeded(jobCondition) {
		// Client-side stress test job is completed
		if isInferenceResultAvailable(instance) {
			msg := "Client-side monitoring prom job has completed"
			util.MarkInferenceStatusSucceeded(instance, corev1.ConditionTrue, msg)
			instance.Status.CompletionTime = &now
			eventMsg := fmt.Sprintf("Client-side monitoring job %s has succeeded", deployedJob.GetName())
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "JobSucceeded", eventMsg)
		} else {
			// Client job has NOT recorded the inference result
			msg := "Monitoring results are not available"
			util.MarkInferenceStatusSucceeded(instance, corev1.ConditionFalse, msg)
		}
	} else if util.IsJobFailed(jobCondition) {
		// Client-side monitoring job is failed
		msg := "Client-side monitoring job has failed"
		util.MarkInferenceStatusFailed(instance, msg)
		instance.Status.CompletionTime = &now
	} else {
		// Client-side monitoring job is still running
		msg := "Client-side monitoring job is running"
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
		PodMetrics:  nil,
		NodeMetrics: nil,
	}

	instance.Status.MonitorResult.PodMetrics = make([]melodyiov1alpha1.PodMetricSpec, 0)
	instance.Status.MonitorResult.NodeMetrics = make([]melodyiov1alpha1.NodeMetricSpec, 0)

	instance.Status.MonitorResult.PodMetrics = append(instance.Status.MonitorResult.PodMetrics, melodyiov1alpha1.PodMetricSpec{
		PodName: instance.Name,
		Metrics: melodyiov1alpha1.PodMetrics{
			Category: melodyiov1alpha1.CPUUsage,
			Value:    consts.DefaultMetricValue,
		},
	})

	instance.Status.MonitorResult.PodMetrics = append(instance.Status.MonitorResult.PodMetrics, melodyiov1alpha1.PodMetricSpec{
		PodName: instance.Name,
		Metrics: melodyiov1alpha1.PodMetrics{
			Category: melodyiov1alpha1.MemUsage,
			Value:    consts.DefaultMetricValue,
		},
	})

	instance.Status.MonitorResult.PodMetrics = append(instance.Status.MonitorResult.PodMetrics, melodyiov1alpha1.PodMetricSpec{
		PodName: instance.Name,
		Metrics: melodyiov1alpha1.PodMetrics{
			Category: melodyiov1alpha1.JobCompletionTime,
			Value:    consts.DefaultMetricValue,
		},
	})

	for _, node := range instance.Spec.OptionalNodes {
		instance.Status.MonitorResult.NodeMetrics = append(instance.Status.MonitorResult.NodeMetrics, melodyiov1alpha1.NodeMetricSpec{
			NodeName: node,
			Metrics:  melodyiov1alpha1.NodeMetrics{Category: melodyiov1alpha1.CPUResource, Value: consts.DefaultMetricValue},
		})
		instance.Status.MonitorResult.NodeMetrics = append(instance.Status.MonitorResult.NodeMetrics, melodyiov1alpha1.NodeMetricSpec{
			NodeName: node,
			Metrics:  melodyiov1alpha1.NodeMetrics{Category: melodyiov1alpha1.MemResource, Value: consts.DefaultMetricValue},
		})
	}
}

func isInferenceResultAvailable(instance *melodyiov1alpha1.Inference) bool {
	if instance == nil || &instance.Spec.OptionalNodes == nil || &instance.Spec.ServingTemplate == nil {
		return false
	}

	// Get the name of the objective metric
	targetInferenceName := instance.Name
	if instance.Status.MonitorResult != nil {
		if instance.Status.MonitorResult.PodMetrics != nil {
			for _, metric := range instance.Status.MonitorResult.PodMetrics {
				// Find the objective metric record from trail status
				if metric.PodName == targetInferenceName {
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
