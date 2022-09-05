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

package inference

import (
	"context"
	"github.com/go-logr/logr"
	logger "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"melody/controllers/inference/dbclient"
	"melody/controllers/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	melodyiov1alpha1 "melody/api/v1alpha1"
)

const (
	ControllerName = "InferenceController"
)

var (
	log = logf.Log.WithName("inference-controller")
)

// InferenceReconciler reconciles a Inference object
type InferenceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	dbclient.DBClient
	updateStatusHandler updateStatusFunc
}

func NewReconciler(mgr manager.Manager) *InferenceReconciler {
	r := &InferenceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		DBClient: dbclient.NewInferenceDBClient(),
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

// +kubebuilder:rbac:groups=melody.io,resources=inferences,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=melody.io,resources=inferences/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=melody.io,resources=inferences/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *InferenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the inference instance
	original := &melodyiov1alpha1.Inference{}
	err := r.Get(context.TODO(), req.NamespacedName, original)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("try to get inference, but it has been deleted", "key", req.String())
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Inference Get error")
		return reconcile.Result{}, err
	}

	instance := original.DeepCopy()
	// If not created, create the inference
	if !util.IsCreatedInference(instance) {
		if instance.Status.StartTime == nil {
			now := metav1.Now()
			instance.Status.StartTime = &now
		}
		msg := "Inference is created"
		util.MarkInferenceStatusCreatedInference(instance, msg)
	} else {
		// Reconcile inference
		err := r.reconcileInference(instance)
		if err != nil {
			logger.Error(err, "Reconcile inference error")
			return reconcile.Result{}, err
		}
	}
	// Update trial status
	if !equality.Semantic.DeepEqual(original.Status, instance.Status) {
		err = r.updateStatusHandler(instance)
		if err != nil {
			logger.Error(err, "Update inference instance status error")
			return reconcile.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InferenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New(ControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		log.Error(err, "Failed to create inference controller")
		return err
	}
	if err = addWatch(c); err != nil {
		log.Error(err, "Inference watch failed")
		return err
	}
	logger.Info("Inference controller created")
	return nil
}

func addWatch(c controller.Controller) error {
	// Watch for changes to trial
	err := c.Watch(&source.Kind{Type: &melodyiov1alpha1.Inference{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Inference watch error")
		return err
	}
	// Watch for changes to client job
	err = c.Watch(
		&source.Kind{Type: &batchv1.Job{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &melodyiov1alpha1.Inference{},
		})
	if err != nil {
		log.Error(err, "Client Job watch error")
		return err
	}
	// Watch for changes to service deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &melodyiov1alpha1.Inference{},
	})
	if err != nil {
		log.Error(err, "Service Deployment watch error")
		return err
	}
	// Watch for changes to service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &melodyiov1alpha1.Inference{},
	})
	if err != nil {
		log.Error(err, "Inference Service watch error")
		return err
	}
	return nil
}

//reconcileInference reconciles the inference with core functions
func (r *InferenceReconciler) reconcileInference(instance *melodyiov1alpha1.Inference) error {
	// Get desired service, and reconcile it
	service, err := r.getDesiredService(instance)
	if err != nil {
		logger.Error(err, "Inference service get error")
		return err
	}
	// Get desired deployment
	desiredDeploy, err := r.getDesiredDeploymentSpec(instance)
	if err != nil {
		logger.Error(err, "Service deployment construction error")
		return err
	}

	// Get desired client job
	desiredJob, err := r.getDesiredJobSpec(instance)
	if err != nil {
		logger.Error(err, "Client-side job construction error")
		return err
	}

	// Reconcile the service
	err = r.reconcileService(instance, service)
	if err != nil {
		logger.Error(err, "Reconcile ML service error")
		return err
	}

	// Reconcile the deployment
	deployedDeployment, err := r.reconcileServiceDeployment(instance, desiredDeploy)
	if err != nil {
		logger.Error(err, "Reconcile ML deployment error")
		return err
	}
	// Check if the job need to be deleted
	if deployedDeployment == nil {
		_, err := r.reconcileJob(instance, desiredJob)
		if err != nil {
			logger.Error(err, "Reconcile client-side job error")
			return err
		}
		return nil
	}

	deployedJob := &batchv1.Job{}
	// Create client job
	if util.IsServiceDeplomentReady(deployedDeployment.Status.Conditions) {
		logger.Info("Service Pod is ready", "name", deployedDeployment.GetName())
		deployedJob, err = r.reconcileJob(instance, desiredJob)
		if err != nil || deployedJob == nil {
			logger.Error(err, "Reconcile client-side job error")
			return err
		}
	}
	// Update inference status (conditions and results) from DB connection
	if util.IsServiceDeplomentReady(deployedDeployment.Status.Conditions) {
		if err = r.UpdateInferenceStatusByClientJob(instance, deployedJob); err != nil {
			logger.Error(err, "Update inference status by client-side job condition error")
			return err
		}
	} else {
		r.UpdateInferenceStatusByServiceDeployment(instance, deployedDeployment)
	}
	return nil
}
