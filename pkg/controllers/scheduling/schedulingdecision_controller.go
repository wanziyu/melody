/*
Copyright 2022. The Melody Authors

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

package scheduling

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"melody/pkg/controllers/scheduling/scheduling_client"
	"melody/pkg/controllers/util"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	melodyiov1alpha1 "melody/api/v1alpha1"
)

const (
	ControllerName = "scheduling-controller"
)

var (
	log = logf.Log.WithName(ControllerName)
)

// NewReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) *SchedulingDecisionReconciler {
	r := &SchedulingDecisionReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(ControllerName),
	}
	r.Sampling = scheduling_client.New(mgr.GetScheme(), mgr.GetClient())
	r.updateStatusHandler = r.updateStatus
	return r
}

func (r *SchedulingDecisionReconciler) updateStatus(instance *melodyiov1alpha1.SchedulingDecision) error {
	err := r.Status().Update(context.TODO(), instance)
	if err != nil {
		if !errors.IsConflict(err) {
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SchedulingDecisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New(ControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		log.Error(err, "Failed to create scheduling controller")
		return err
	}
	// Add watch
	if err = addWatch(c); err != nil {
		log.Error(err, "Inference watch failed")
		return err
	}
	log.Info("Scheduling controller created")
	return nil
}

// Add Watch of resources
func addWatch(c controller.Controller) error {
	// Watch for changes to Experiment
	err := c.Watch(&source.Kind{Type: &melodyiov1alpha1.SchedulingDecision{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Scheduling watch failed")
		return err
	}

	// Watch for trials for the experiments
	err = c.Watch(
		&source.Kind{Type: &melodyiov1alpha1.Inference{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &melodyiov1alpha1.SchedulingDecision{},
		})
	if err != nil {
		log.Error(err, "Inference watch failed")
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &SchedulingDecisionReconciler{}

type updateStatusFunc func(instance *melodyiov1alpha1.SchedulingDecision) error

// SchedulingDecisionReconciler reconciles a SchedulingDecision object
type SchedulingDecisionReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	sampling_client.Sampling
	updateStatusHandler updateStatusFunc
}

//+kubebuilder:rbac:groups=melody.io,resources=schedulingdecisions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=melody.io,resources=schedulingdecisions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=melody.io,resources=schedulingdecisions/finalizers,verbs=update
// +kubebuilder:rbac:groups=melody.io,resources=inferences,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=melody.io,resources=inferences/status,verbs=get;update;patch

// Reconcile reads that state of the cluster for a trial object and makes changes based on the state read

func (r *SchedulingDecisionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.WithValues("Scheduling", req.NamespacedName)

	// Fetch the profiling experiment instance
	original := &melodyiov1alpha1.SchedulingDecision{}
	err := r.Get(context.TODO(), req.NamespacedName, original)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		logger.Error(err, "Scheduling decision get error")
		return reconcile.Result{}, err
	}
	instance := original.DeepCopy()

	// Cleanup upon completion
	if util.IsCompletedScheduling(instance) {
		if !util.HasRunningInferences(instance) {
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
