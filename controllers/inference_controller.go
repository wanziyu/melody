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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	melodyiov1alpha1 "melody/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	util "melody/controllers/utils"
)

const (
	ControllerName = "inference-controller"
)

var (
	log = logf.Log.WithName(ControllerName)
)

//NewInferenceReconciler returns a new reconciler
func NewInferenceReconciler(mgr manager.Manager) *InferenceReconciler {
	r := &InferenceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(ControllerName),
		Log:      logf.Log.WithName(ControllerName),
	}
	//r.updateStatusHandler = r.updateStatus
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
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	//updateStatusHandler updateStatusFunc
}

//type updateStatusFunc func(instance *melodyiov1alpha1.Inference) error

func addWatch(c controller.Controller) error {
	// Watch for changes to inference
	err := c.Watch(&source.Kind{Type: &melodyiov1alpha1.Inference{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Inference watch error")
		return err
	}

	// Watch for changes to service deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &melodyiov1alpha1.Inference{},
	})
	if err != nil {
		log.Error(err, "Inference Deployment watch error")
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

//+kubebuilder:rbac:groups=melody.io.melody.io,resources=inferences,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=melody.io.melody.io,resources=inferences/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=melody.io.melody.io,resources=inferences/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=services,verbs=get;list;create;update;patch;delete

func (r *InferenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.WithValues("Inference", req.NamespacedName)

	// 如果监听到事件的变化
	// Fetch the inference instance
	original := &melodyiov1alpha1.Inference{}
	err := r.Get(ctx, req.NamespacedName, original)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			log.Info("try to get inference, but it has been deleted", "key", req.String())
			return reconcile.Result{}, nil
		}
		logger.Error(err, "Inference instance get error")
		return reconcile.Result{}, err
	}

	//Fetch the Inference fields
	instance := original.DeepCopy()
	// If not created, create the inference
	if !util.IsCreatedInference(instance) {
		if instance.Status.StartTime == nil {
			now := metav1.Now()
			instance.Status.StartTime = &now
		}
		msg := "Inference is created"
		util.MarkInferenceStatusPendingTrial(instance, msg)
	} else {
		// Reconcile Inference
		err := r.reconcileInference(instance)
		if err != nil {
			logger.Error(err, "Reconcile inference error")
			return reconcile.Result{}, err
		}
	}

	//4) Compare status before-and-after reconciling and update changes to cluster.
	if !reflect.DeepEqual(original.Status, instance.Status) {
		if err = r.Client.Status().Update(context.Background(), instance); err != nil {
			if errors.IsConflict(err) {
				// retry later when update operation violates with etcd concurrency control.
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

//reconcileInference reconcile the inference with core functions
func (r *InferenceReconciler) reconcileInference(instance *melodyiov1alpha1.Inference) error {
	logger := log.WithValues("Inference", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// 获得期望的Service 然后Reconcile
	service, err := r.getDesiredService(instance)
	if err != nil {
		logger.Error(err, "ML service get error")
		return err
	}

	// 获得期望的deployment, 然后Reconcile
	desiredDeploy, err := r.getDesiredDeploymentSpec(instance)
	if err != nil {
		logger.Error(err, "Service deployment construction error")
		return err
	}

	// Reconcile创建的service实例
	err = r.reconcileService(instance, service)
	if err != nil {
		logger.Error(err, "Reconcile ML inference service error")
		return err
	}

	// Reconcile创建的deployment实例
	deployedDeployment, err := r.reconcileServiceDeployment(instance, desiredDeploy)
	if err != nil {
		logger.Error(err, "Reconcile ML inference deployment error")
		return err
	}

	// 更新inference的状态
	if util.IsServiceDeplomentReady(deployedDeployment.Status.Conditions) {
		logger.Info("Service Pod is ready", "name", deployedDeployment.GetName())
		err = r.updateInferenceStatus(instance, deployedDeployment)

		if err != nil {
			logger.Error(err, "Update ML inference status error")
			return err
		}
	}
	return nil

}

func (r *InferenceReconciler) updateInferenceStatus(instance *melodyiov1alpha1.Inference, deploy *appsv1.Deployment) error {

	//logger := r.Log.WithValues("Inference", "updateStatus")
	// 2) Sync each predictor to deploy containers mounted with specific model.
	for pi := range instance.Spec.Servings {
		predictor := &instance.Spec.Servings[pi]
		endpoint := util.SvcHostForPredictor(instance, predictor)
		//result, err := r.syncPredictor(&instance, pi, predictor)
		if psLen := len(instance.Status.ServingStatuses); psLen == 0 || psLen < pi {
			instance.Status.ServingStatuses = append(instance.Status.ServingStatuses, melodyiov1alpha1.ServingStatus{
				Name:              predictor.Name,
				Replicas:          deploy.Status.Replicas,
				ReadyReplicas:     deploy.Status.ReadyReplicas,
				InferenceEndpoint: endpoint,
			})
		} else {
			ps := &instance.Status.ServingStatuses[pi]
			ps.Name = predictor.Name
			ps.Replicas = deploy.Status.Replicas
			ps.ReadyReplicas = deploy.Status.ReadyReplicas
			ps.InferenceEndpoint = endpoint
		}
		if len(instance.Status.ServingStatuses) > pi+1 {
			instance.Status.ServingStatuses = instance.Status.ServingStatuses[:pi+1]
		}

	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InferenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New(ControllerName, mgr, controller.Options{Reconciler: r})

	if err != nil {
		log.Error(err, "Failed to create inference controller")
		return err
	}
	//watch for changes Inference
	if err = addWatch(c); err != nil {
		log.Error(err, "Inference watch failed")
		return err
	}
	log.Info("Inference controller created")
	return nil
}
