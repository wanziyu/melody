package dbclient

import (
	"fmt"
	client "github.com/influxdata/influxdb1-client/v2"
	logger "github.com/sirupsen/logrus"
	melodyiov1alpha1 "melody/api/v1alpha1"
	consts "melody/controllers/const"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

var (
	log                = logf.Log.WithName("inference-db-client")
	timeout            = 60 * time.Second
	defaultMetricValue = string("0.0")
)

type DBClient interface {
	GetMonitorResult(inference *melodyiov1alpha1.Inference) (*melodyiov1alpha1.MonitoringResult, error)
}

type InferenceDBClient struct {
}

func NewInferenceDBClient() DBClient {
	return &InferenceDBClient{}
}

func (t InferenceDBClient) GetMonitorResult(inference *melodyiov1alpha1.Inference) (*melodyiov1alpha1.MonitoringResult, error) {
	// Prepare db request
	conn := connInflux()
	podMetricRequest := prepareDBRequest(inference)

	res, err := queryDB(conn, podMetricRequest)
	logger.Info("Query data: ", res)
	if err != nil {
		log.Error(err, "Failed to get monitor result from db storage")
		return nil, err
	}
	// Validate and convert response
	reply := validateDBResult(inference, res)
	return reply, nil
}

func validateDBResult(inference *melodyiov1alpha1.Inference, response []client.Result) *melodyiov1alpha1.MonitoringResult {

	reply := &melodyiov1alpha1.MonitoringResult{}
	reply.PodMetrics.PodName = inference.Name
	reply.PodMetrics.Metrics = make([]melodyiov1alpha1.PodMetrics, 0)
	values := make([]string, 0)
	reply.PodMetrics.Metrics = append(reply.PodMetrics.Metrics, melodyiov1alpha1.PodMetrics{Category: melodyiov1alpha1.CPUUsage, Value: values})
	reply.PodMetrics.Metrics = append(reply.PodMetrics.Metrics, melodyiov1alpha1.PodMetrics{Category: melodyiov1alpha1.MemUsage, Value: values})
	reply.PodMetrics.Metrics = append(reply.PodMetrics.Metrics, melodyiov1alpha1.PodMetrics{Category: melodyiov1alpha1.JobCompletionTime, Value: values})
	reply.NodeMetrics = make([]melodyiov1alpha1.NodeMetricSpec, 0)

	if response != nil {
		for _, row := range response[0].Series[0].Values {
			for i, metric := range reply.PodMetrics.Metrics {
				if metric.Category == melodyiov1alpha1.CPUUsage {
					reply.PodMetrics.Metrics[i].Value = append(reply.PodMetrics.Metrics[i].Value, fmt.Sprintf("%v", row[1]))
				} else if metric.Category == melodyiov1alpha1.MemUsage {
					reply.PodMetrics.Metrics[i].Value = append(reply.PodMetrics.Metrics[i].Value, fmt.Sprintf("%v", row[4]))
				} else if metric.Category == melodyiov1alpha1.JobCompletionTime {
					reply.PodMetrics.Metrics[i].Value = append(reply.PodMetrics.Metrics[i].Value, fmt.Sprintf("%v", row[3]))
				}
			}
		}

	} else {
		log.Info("Get nil monitoring result of inference %s.%s, will save objective value as 0", inference.Name, inference.Namespace)
		cpu := melodyiov1alpha1.PodMetrics{Category: melodyiov1alpha1.CPUUsage, Value: values}
		mem := melodyiov1alpha1.PodMetrics{Category: melodyiov1alpha1.MemUsage, Value: values}
		jct := melodyiov1alpha1.PodMetrics{Category: melodyiov1alpha1.JobCompletionTime, Value: values}
		reply.PodMetrics.Metrics = []melodyiov1alpha1.PodMetrics{cpu, mem, jct}

		/*		for _, node := range inference.Spec.OptionalNodes {
				reply.NodeMetrics = append(reply.NodeMetrics, melodyiov1alpha1.NodeMetricSpec{
					NodeName: node,
					Metrics:  melodyiov1alpha1.NodeMetrics{Category: melodyiov1alpha1.CPUResource, Value: defaultMetricValue},
				})
				reply.NodeMetrics = append(reply.NodeMetrics, melodyiov1alpha1.NodeMetricSpec{
					NodeName: node,
					Metrics:  melodyiov1alpha1.NodeMetrics{Category: melodyiov1alpha1.MemResource, Value: defaultMetricValue},
				})
			}*/

	}

	return reply
}

func prepareDBRequest(inference *melodyiov1alpha1.Inference) string {
	request := fmt.Sprintf("SELECT * FROM %s WHERE \"inference\" = '%s' LIMIT %d", consts.DefaultMelodyPodMetricMeasurement, inference.Name, 5)
	return request
}
