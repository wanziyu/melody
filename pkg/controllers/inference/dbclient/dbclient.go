package dbclient

import (
	melodyiov1alpha1 "melody/api/v1alpha1"
	"melody/pkg/controllers/util"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
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
	request := prepareDBRequest(trial)

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
		log.Error(err, "Failed to get trial result from db storage")
		return nil, err
	}

	// Validate and convert response
	reply := validateDBResult(trial, response)
	return reply, nil
}
