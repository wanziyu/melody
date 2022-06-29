package controllers

import (
	"melody/api/v1alpha1"
	"melody/pkg/controllers/inference"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func init() {
	SetupWithManagerMap[&v1alpha1.Inference{}] = func(mgr controllerruntime.Manager) error {
		return inference.NewReconciler(mgr).SetupWithManager(mgr)
	}
}
