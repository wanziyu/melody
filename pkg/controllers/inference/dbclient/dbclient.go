package dbclient

import (
	"google.golang.org/grpc"
	melodyiov1alpha1 "melody/api/v1alpha1"
	"melody/pkg/controllers/util"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	"context"
	api_pb "melody/api/v1alpha1/grpc_proto/grpc_storage/go"
)

var (
	log                = logf.Log.WithName("inference-db-client")
	timeout            = 60 * time.Second
	defaultMetricValue = string("0.0")
)

type DBClient interface {
	GetMonitorResult(trial *melodyiov1alpha1.Inference) (*melodyiov1alpha1.MonitoringResult, error)
}

type InferenceDBClient struct {
}

func NewInferenceDBClient() DBClient {
	return &InferenceDBClient{}
}

func (t InferenceDBClient) GetMonitorResult(inference *melodyiov1alpha1.Inference) (*melodyiov1alpha1.MonitoringResult, error) {
	// Prepare db request
	request := prepareDBRequest(inference)

	// Dial DB storage
	endpoint := util.GetDBStorageEndpoint()
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	clientGRPC := api_pb.NewDBClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Send request, receive reply
	response, err := clientGRPC.GetResult(ctx, request, grpc.WaitForReady(true))
	if err != nil {
		log.Error(err, "Failed to get monitor result from db storage")
		return nil, err
	}

	// Validate and convert response
	reply := validateDBResult(inference, response)
	return reply, nil
}

func validateDBResult(inference *melodyiov1alpha1.Inference, response *api_pb.GetResultReply) *melodyiov1alpha1.MonitoringResult {

	reply := &melodyiov1alpha1.MonitoringResult{
		PodMetrics:  nil,
		NodeMetrics: nil,
	}

	reply.PodMetrics = make([]melodyiov1alpha1.PodMetricSpec, 0)
	reply.NodeMetrics = make([]melodyiov1alpha1.NodeMetricSpec, 0)

	if response != nil {
		for _, metric := range response.Results {
			if metric.Name == inference.Name {
				reply.PodMetrics = append(reply.PodMetrics, melodyiov1alpha1.PodMetricSpec{
					Category: melodyiov1alpha1.Category(metric.Key),
					Value:    metric.Value,
				})
			} else {
				reply.NodeMetrics = append(reply.NodeMetrics, melodyiov1alpha1.NodeMetricSpec{
					NodeName: metric.Name,
					Metrics:  melodyiov1alpha1.NodeMetrics{Category: melodyiov1alpha1.Category(metric.Key), Value: metric.Value},
				})
			}
		}

	} else {
		log.Info("Get nil monitoring result of inference %s.%s, will save objective value as 0", inference.Name, inference.Namespace)

		reply.PodMetrics = append(reply.PodMetrics, melodyiov1alpha1.PodMetricSpec{
			Category: melodyiov1alpha1.CPUUsage,
			Value:    defaultMetricValue,
		})

		reply.PodMetrics = append(reply.PodMetrics, melodyiov1alpha1.PodMetricSpec{
			Category: melodyiov1alpha1.MemUsage,
			Value:    defaultMetricValue,
		})

		reply.PodMetrics = append(reply.PodMetrics, melodyiov1alpha1.PodMetricSpec{
			Category: melodyiov1alpha1.JobCompletionTime,
			Value:    defaultMetricValue,
		})

		for _, node := range inference.Spec.OptionalNodes {
			reply.NodeMetrics = append(reply.NodeMetrics, melodyiov1alpha1.NodeMetricSpec{
				NodeName: node,
				Metrics:  melodyiov1alpha1.NodeMetrics{Category: melodyiov1alpha1.CPUResource, Value: defaultMetricValue},
			})
			reply.NodeMetrics = append(reply.NodeMetrics, melodyiov1alpha1.NodeMetricSpec{
				NodeName: node,
				Metrics:  melodyiov1alpha1.NodeMetrics{Category: melodyiov1alpha1.MemResource, Value: defaultMetricValue},
			})
		}

	}

	return reply
}

func prepareDBRequest(inference *melodyiov1alpha1.Inference) *api_pb.GetResultRequest {
	request := &api_pb.GetResultRequest{
		Namespace:     inference.Namespace,
		InferenceName: inference.Name,
	}
	return request
}
