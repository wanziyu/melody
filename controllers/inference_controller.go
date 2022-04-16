/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	melodyiov1alpha1 "melody/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	ControllerName = "inference-controller"
)

var (
	log = logf.Log.WithName(ControllerName)
)

//NewReconciler returns a new reconciler
func NewReconciler(mgr manager.Manager) *InferenceReconciler {
	r := &InferenceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(ControllerName),
		Log:      logf.Log.WithName(ControllerName),
	}
	r.updateStatusHandler = r.updateStatus
	return r
}

func (r *InferenceReconciler) updateStatus(instance *melodyiov1alpha1.Inference) error {
	err := r.Status().Update(context.TODO(), instance)
	if err != nil {
		if !errors.IsConflict(err) {
			return err
		}
	}
	return nil
}

// InferenceReconciler reconciles a Inference object
type InferenceReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	recorder            record.EventRecorder
	updateStatusHandler updateStatusFunc
}

type updateStatusFunc func(instance *melodyiov1alpha1.Inference) error

var _ reconcile.Reconciler = &InferenceReconciler{}

//+kubebuilder:rbac:groups=melody.io.melody.io,resources=inferences,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=melody.io.melody.io,resources=inferences/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=melody.io.melody.io,resources=inferences/finalizers,verbs=update
//+kubebuilder:rbac:groups=melody.io.melody.io,resources=pod,verbs=get;update;patch
func (r *InferenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.WithValues("Inference", req.NamespacedName)

	// Fetch the inference instance
	original := &melodyiov1alpha1.Inference{}
	err := r.Get(context.TODO(), req.NamespacedName, original)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		logger.Error(err, "Inference instance get error")
		return reconcile.Result{}, err
	}
	instance := original.DeepCopy()

	// Cleanup upon completion
	if util.IsCompletedExperiment(instance) {
		if !util.HasRunningTrials(instance) {
			return reconcile.Result{}, nil
		}
	}
	if !util.IsCreatedExperiment(instance) {
		// Create the experiment
		if instance.Status.StartTime == nil {
			now := metav1.Now()
			instance.Status.StartTime = &now
		}
		message := "Experiment is created"
		util.MarkExperimentStatusCreated(instance, message)
	} else {
		// Reconcile experiment
		err := r.ReconcileExperiment(instance)
		if err != nil {
			logger.Error(err, "Reconcile experiment error")
			r.recorder.Eventf(instance, corev1.EventTypeWarning, "ReconcileFailed", "Failed to reconcile: %v", err)
			return reconcile.Result{}, err
		}
	}

	// Update experiment status
	if !equality.Semantic.DeepEqual(original.Status, instance.Status) {
		err = r.updateStatusHandler(instance)
		if err != nil {
			logger.Error(err, "Update experiment status error")
			return reconcile.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InferenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&melodyiov1alpha1.Inference{}).
		Complete(r)
}
