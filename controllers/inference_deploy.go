package controllers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	melodyiov1alpha1 "melody/api/v1alpha1"
	util "melody/controllers/utils"
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
					Name: consts.DefaultServicePortName,
					Port: consts.DefaultServicePort,
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
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

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
