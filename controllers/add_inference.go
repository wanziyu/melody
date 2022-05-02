package controllers

import (
	melodyiov1alpha1 "melody/api/v1alpha1"
	runtime "sigs.k8s.io/controller-runtime"
)

func init() {
	SetupWithManagerMap[&melodyiov1alpha1.Inference{}] = func(mgr runtime.Manager) error {
		return NewInferenceReconciler(mgr).SetupWithManager(mgr)
	}
}
