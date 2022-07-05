package scheduling_client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	grpcapi "melody/api/v1alpha1/grpc_proto/grpc_algorithm/go"
	"melody/pkg/controllers/consts"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	melodyv1alpha1 "melody/api/v1alpha1"
)

type Scheduling interface {
	GetScheduling(numRequests int32, instance *melodyv1alpha1.SchedulingDecision, currentCount int32, inferences []melodyv1alpha1.Inference) ([]melodyv1alpha1.SchedulingResult, error)
}

var (
	log     = logf.Log.WithName("scheduling_client-client")
	timeout = 60 * time.Second
)

type General struct {
	scheme *runtime.Scheme
	client.Client
}

func New(scheme *runtime.Scheme, client client.Client) Scheduling {
	return &General{scheme: scheme, Client: client}
}

func (g *General) GetScheduling(requestNum int32, instance *melodyv1alpha1.SchedulingDecision, currentCount int32, inferences []melodyv1alpha1.Inference) ([]melodyv1alpha1.SchedulingResult, error) {
	logger := log.WithValues("Scheduling", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	if requestNum <= 0 {
		err := fmt.Errorf("request samplings should be lager than zero")
		return nil, err
	}

	if (instance.Spec.MaxNumInferences != nil) && (requestNum+currentCount > *instance.Spec.MaxNumInferences) {
		err := fmt.Errorf("request samplings should smaller than MaxNumInferences")
		return nil, err
	}

	if (instance.Spec.Parallelism != nil) && (requestNum > *instance.Spec.Parallelism) {
		err := fmt.Errorf("request samplings should smaller than Parallelism")
		return nil, err
	}

	endpoint := getAlgorithmServerEndpoint() //"localhost:9996"
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	clientGRPC := grpcapi.NewSuggestionClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	request, err := newSchedulingRequest(requestNum, instance, currentCount, inferences)
	if err != nil {
		return nil, err
	}

	response, err := clientGRPC.GetSuggestions(ctx, request, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}

	if len(response.SchedulingDecisions) != int(requestNum) {
		err := fmt.Errorf("the response contains unexpected inferences")
		logger.Error(err, "The response contains unexpected inferences", "requestNum", requestNum, "response", response)
		return nil, err
	}

	// Succeeded
	logger.V(0).Info("Getting scheduling decisions", "endpoint", endpoint, "response", response.String(), "request", request)
	results := make([]melodyv1alpha1.SchedulingResult, 0)
	for _, t := range response.SchedulingDecisions {
		results = append(results,
			melodyv1alpha1.SchedulingResult{
				InferenceName: t.InferenceName, // inference name
				TargetNode:    t.TargetNode,
				//ResourceAssignments:  t.Resource
				ResourceAssignments: composeResourceAssignments(t.Resource),
			})
	}
	return results, nil
}

func composeResourceAssignments(pas []*grpcapi.KeyValue) []melodyv1alpha1.ResourceAssignment {

	res := make([]melodyv1alpha1.ResourceAssignment, 0)

	for _, pa := range pas {
		categoryThis := melodyv1alpha1.MemResource
		if pa.Key != string(categoryThis) {
			categoryThis = melodyv1alpha1.CPUResource
		}

		res = append(res, melodyv1alpha1.ResourceAssignment{
			Name:     pa.Key,
			Value:    pa.Value,
			Category: categoryThis,
			//todo: Category
		})
	}
	return res

}

func newSchedulingRequest(requestNum int32, instance *melodyv1alpha1.SchedulingDecision, currentCount int32, inferences []melodyv1alpha1.Inference) (*grpcapi.SchedulingRequest, error) {
	request := &grpcapi.SchedulingRequest{
		AlgorithmName:    string(instance.Spec.Algorithm.AlgorithmName),
		RequiredSampling: requestNum,
	}
	if instance.Spec.MaxNumTrials != nil {
		request.SamplingNumberSpecified = *instance.Spec.MaxNumTrials
	}
	pars, err := convertPars(instance)
	if err != nil {
		return nil, err
	}
	request.Parameters = pars

	existingTrials, err := convertTrials(trials)
	if err != nil {
		return nil, err
	}
	request.ExistingResults = existingTrials

	request.IsFirstRequest = currentCount < 1
	request.AlgorithmExtraSettings = convertSettings(instance)
	request.IsMaximize = instance.Spec.Objective.Type == morphlingv1alpha1.ObjectiveTypeMaximize
	return request, nil
}

func getAlgorithmServerEndpoint() string {

	serviceName := consts.DefaultSamplingService
	return fmt.Sprintf("%s:%d",
		serviceName,
		//s.Namespace,
		consts.DefaultSamplingPort)
}
