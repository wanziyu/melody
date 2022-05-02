package consts

import "os"

const (
	// LabelExperimentName is the label of experiment name.
	LabelInferenceName = "inference"
	// LabelTrialName is the label of trial name.
	LabelSchedulingDecesionName = "schedulingdecesion"
	// LabelDeploymentName is the label of deployment name.
	LabelDeploymentName = "deployment"
	LabelDomainName     = "domain"
	// DefaultServicePort is the default port of sampling_client service.
	InferenceServicePort   = 8500
	InferenceContainerPort = 8300
	// DefaultServicePortName is the default port name of sampling_client service.
	InferenceServicePortName = "inference-service"
	// DefaultMetricValue is the default trial result value, set for failed trials
	DefaultMetricValue = "0.0"
	// DefaultSamplingService is the default algorithm k8s service name
	DefaultSamplingService = "morphling-algorithm-server"
	// DefaultSamplingPort is the default port of algorithm service.
	DefaultSamplingPort = 9996
	// DefaultMorphlingMySqlServiceName is the default mysql k8s service name
	DefaultMorphlingMySqlServiceName = "morphling-mysql"
	// DefaultMorphlingMySqlServicePort is the default mysql k8s service port
	DefaultMorphlingMySqlServicePort = "3306"
	// DefaultMorphlingDBManagerServiceName is the default db-manager k8s service name
	DefaultMorphlingDBManagerServiceName = "morphling-db-manager"

	//DefaultMorphlingDBManagerServicePort = 6799
	//DefaultMorphlingNamespace            = "morphling-system"
)

var (
	DefaultControllerNamespace = GetEnvOrDefault("MORPHLING_CORE_NAMESPACE", "morphling-system")
	// DefaultMorphlingDBManagerServicePort is the default db-manager k8s service port
	DefaultMorphlingDBManagerServicePort = GetEnvOrDefault("DB_PORT", "6799")
)

func GetEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
