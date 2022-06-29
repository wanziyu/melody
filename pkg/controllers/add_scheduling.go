package controllers

import (
	"melody/api/v1alpha1"
	"melody/pkg/controllers/scheduling"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	SetupWithManagerMap[&v1alpha1.SchedulingDecision{}] = func(mgr controllerruntime.Manager) error {
		return scheduling.NewReconciler(mgr).SetupWithManager(mgr)
	}
}
