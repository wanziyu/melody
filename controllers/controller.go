package controllers

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var SetupWithManagerMap = make(map[runtime.Object]func(mgr manager.Manager) error)

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {
	for workload, f := range SetupWithManagerMap {
		if err := f(m); err != nil {
			return err
		}
		gvk, err := apiutil.GVKForObject(workload, m.GetScheme())
		if err != nil {
			klog.Warningf("unrecognized workload object %+v in scheme: %v", gvk, err)
			return err
		}
		klog.Infof("workload %v controller has started.", gvk.Kind)
	}

	return nil
}
