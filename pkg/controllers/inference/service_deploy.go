package inference

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	melodyiov1alpha1 "melody/api/v1alpha1"
	"melody/pkg/controllers/consts"
	"melody/pkg/controllers/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// getDesiredService returns a new k8s service for ML service test
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
	// ToDo: SetControllerReference here is useless, as the controller delete svc upon trial completion
	// Add owner reference to the service so that it could be GC
	if err := controllerutil.SetControllerReference(t, service, r.Scheme); err != nil {
		return nil, err
	}
	return service, nil
}

// reconcileService reconciles a k8s service for ML service
func (r *InferenceReconciler) reconcileService(instance *melodyiov1alpha1.Inference, service *corev1.Service) error {
	logger := log.WithValues("Inference", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	foundService := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	// Create svc
	if err != nil && errors.IsNotFound(err) && !util.IsCompletedInference(instance) {
		logger.Info("Creating ML service", "namespace", service.Namespace, "name", service.Name)
		err = r.Create(context.TODO(), service)
		return err
	}
	// Delete svc
	if util.IsCompletedInference(instance) {
		// Delete svc upon trial completions
		if foundService.ObjectMeta.DeletionTimestamp != nil || errors.IsNotFound(err) {
			logger.Info("Deleting ML service")
			return nil
		}
		if err = r.Delete(context.TODO(), foundService, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Delete ML service operation is redundant")
				return nil
			}
			return err
		}
	}
	return nil
}

// getDesiredPodSpec returns a new deployment containing the inference service under test
func (r *InferenceReconciler) getDesiredDeploymentSpec(instance *melodyiov1alpha1.Inference) (*appsv1.Deployment, error) {
	// Prepare podTemplate and embed tunable parameters
	podTemplate := &corev1.PodTemplateSpec{}
	if &instance.Spec.ServingTemplate != nil {
		instance.Spec.ServingTemplate.Template.Spec.DeepCopyInto(&podTemplate.Spec)
	}

	podTemplate.Labels = util.ServicePodLabels(instance)

	//algorithm sever chose a node
	if instance.Spec.NodeSelector != nil {
		podTemplate.Spec.NodeSelector = instance.Spec.NodeSelector
	}
	/*	for i := range podTemplate.Spec.Containers {
		c := &podTemplate.Spec.Containers[i]
		c.Env, c.Args, c.Resources = appendServiceEnv(instance, c.Env, c.Args, c.Resources)
	}*/
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
		},
	}
	if instance.Spec.ServiceProgressDeadline != nil {
		deploy.Spec.ProgressDeadlineSeconds = instance.Spec.ServiceProgressDeadline
	}
	// ToDo: SetControllerReference here is useless, as the controller delete svc upon trial completion
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
		// If not created, create the service deployment
		if errors.IsNotFound(err) {
			if util.IsCompletedInference(instance) {
				return nil, nil
			}

			logger.Info("Creating ML service deployment", "name", deploy.GetName())
			err = r.Create(context.TODO(), deploy)
			if err != nil {
				logger.Error(err, "Create service deployment error", "name", deploy.GetName())
				return nil, err
			}
		} else {
			logger.Error(err, "Get service deployment error", "name", deploy.GetName())
			return nil, err
		}
	} else {
		if util.IsCompletedInference(instance) {
			if deploy.ObjectMeta.DeletionTimestamp != nil || errors.IsNotFound(err) {
				logger.Info("Deleting ML deployment", "name", deploy.GetName())
				return nil, nil
			}
			// // Delete ML deployments upon trial completions
			if err = r.Delete(context.TODO(), deploy, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Delete ML deployment operation is redundant", "name", deploy.GetName())
					return nil, nil
				}
				logger.Error(err, "Delete ML deployment error", "name", deploy.GetName())
				return nil, err
			} else {
				logger.Info("Delete ML deployment succeeded", "name", deploy.GetName())
				return nil, nil
			}
		}
	}
	return deploy, nil
}
