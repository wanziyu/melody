package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	melodyiov1alpha1 "melody/api/v1alpha1"
	consts "melody/controllers/const"
	util "melody/controllers/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// getDesiredService returns a new k8s service for ML service
func (r *InferenceReconciler) getDesiredService(t *melodyiov1alpha1.Inference) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GetServiceName(t),
			Namespace: t.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: util.ServicePodLabels(t),
			Ports: []corev1.ServicePort{
				{
					Name: consts.InferenceServicePortName,
					Port: consts.InferenceServicePort,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// ToDo: SetControllerReference here is useless, as the controller delete svc upon inference completion
	// Add owner reference to the service so that it could be GC
	if err := controllerutil.SetControllerReference(t, service, r.Scheme); err != nil {
		return nil, err
	}
	return service, nil

}

// reconcileService reconciles a k8s service for ML inference instance
func (r *InferenceReconciler) reconcileService(instance *melodyiov1alpha1.Inference, service *corev1.Service) error {
	logger := log.WithValues("Inference", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	foundService := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)

	// 如果不存在service, 就创建
	if err != nil && errors.IsNotFound(err) && !util.IsCompletedInference(instance) {
		logger.Info("Creating ML Inference service", "namespace", service.Namespace, "name", service.Name)
		err = r.Create(context.TODO(), service)
		return err
	}
	// Delete svc
	if util.IsCompletedInference(instance) {
		// Delete svc upon trial completions
		if foundService.ObjectMeta.DeletionTimestamp != nil || errors.IsNotFound(err) {
			logger.Info("Deleting ML inference service")
			return nil
		}
		if err = r.Delete(context.TODO(), foundService, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Delete ML inference service operation is redundant")
				return nil
			}
			return err
		}
	}
	return nil
}

// getDesiredPodSpec returns a new deployment containing the ML service under test
func (r *InferenceReconciler) getDesiredDeploymentSpec(instance *melodyiov1alpha1.Inference) (*appsv1.Deployment, error) {
	// Prepare podTemplate
	podTemplate := &corev1.PodTemplateSpec{}

	podTemplate.Labels = util.ServicePodLabels(instance)

	for pi := range instance.Spec.Servings {
		predictor := &instance.Spec.Servings[pi]
		Containers := []corev1.Container{
			{
				Name: predictor.Name,
				// 用指定的镜像
				Image:           predictor.Image,
				ImagePullPolicy: "IfNotPresent",
				//指定端口
				Ports: []corev1.ContainerPort{
					{
						Name:          "http",
						Protocol:      corev1.ProtocolTCP,
						ContainerPort: consts.InferenceContainerPort,
					},
				},
			},
		}
		podTemplate.Spec.Containers = append(podTemplate.Spec.Containers, Containers...)
	}

	// Prepare k8s deployment
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        util.GetServiceDeploymentName(instance),
			Namespace:   instance.GetNamespace(),
			Labels:      util.ServiceDeploymentLabels(instance),
			Annotations: instance.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: util.ServicePodLabels(instance)},
			Template: *podTemplate,
			Replicas: instance.Spec.Replicas,
		},
	}
	// Add owner reference to the service so that it could be GC
	if err := controllerutil.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		return nil, err
	}
	return deploy, nil
}

// reconcileServiceDeployment reconciles the ML deployment containing the ML service under test
func (r *InferenceReconciler) reconcileServiceDeployment(instance *melodyiov1alpha1.Inference, deploy *appsv1.Deployment) (*appsv1.Deployment, error) {
	logger := log.WithValues("Inference", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.GetName(), Namespace: deploy.GetNamespace()}, deploy)
	if err != nil && !util.IsCompletedInference(instance) {
		// 如果没有创建deployment，就创建
		if errors.IsNotFound(err) {
			if util.IsCompletedInference(instance) {
				return nil, nil
			}

			logger.Info("Creating ML inference deployment", "name", deploy.GetName())
			//创建deployment
			err = r.Create(context.TODO(), deploy)
			//创建成功的log
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "PredictorDeploymentCreated",
				"Deployment %s for predictor successfully created, replicas: %d", deploy.Name, *deploy.Spec.Replicas)

			if err != nil {
				logger.Error(err, "Create inference deployment error", "name", deploy.GetName())
				return nil, err
			}
		} else {
			logger.Error(err, "Get inference deployment error", "name", deploy.GetName())
			return nil, err
		}
	} else {
		//如果完成了，就删除就好了。。。
		if util.IsCompletedInference(instance) {
			//如果已经删除了，或者已经找不到啦
			if deploy.ObjectMeta.DeletionTimestamp != nil || errors.IsNotFound(err) {
				logger.Info("Deleting ML inference deployment", "name", deploy.GetName())
				return nil, nil
			}
			//删除deployment, 如果inference完成
			//Delete ML deployments upon inference completions
			if err = r.Delete(context.TODO(), deploy, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Delete ML inference deployment operation is redundant", "name", deploy.GetName())
					return nil, nil
				}
				logger.Error(err, "Delete ML inference deployment error", "name", deploy.GetName())
				return nil, err
			} else {
				logger.Info("Delete ML inference deployment succeeded", "name", deploy.GetName())
				return nil, nil
			}
		}
	}
	return deploy, nil
}

// getDesiredJobSpec returns a new inference run job from the template on the inference
func (r *InferenceReconciler) getDesiredJobSpec(instance *melodyiov1alpha1.Inference) (*batchv1.Job, error) {
	logger := log.WithValues("Inference", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:        util.GetStressTestJobName(instance),
			Namespace:   instance.GetNamespace(),
			Labels:      util.ServiceDeploymentLabels(instance),
			Annotations: instance.Annotations,
		},
	}
	/*	if &instance.Spec.ClientTemplate != nil {
		instance.Spec.ClientTemplate.Spec.DeepCopyInto(&job.Spec)
	}*/
	// The default restart policy for a pod is not acceptable in the context of a job
	if job.Spec.Template.Spec.RestartPolicy == "" {
		job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	}
	// The default backoff limit will restart the trial job which is unlikely to produce desirable results
	if job.Spec.BackoffLimit == nil {
		job.Spec.BackoffLimit = new(int32)
	}
	/*	// Expose the current assignments as environment variables to every container
		for i := range job.Spec.Template.Spec.Containers {
			c := &job.Spec.Template.Spec.Containers[i]
			c.Env = appendJobEnv(instance, c.Env)
		}*/

	if err := controllerutil.SetControllerReference(instance, job, r.Scheme); err != nil {
		logger.Error(err, "Set inference job controller reference error", "name", job.GetName())
		return nil, err
	}
	return job, nil
}
