package util

import (
	"fmt"
	melodyiov1alpha1 "melody/api/v1alpha1"
	consts "melody/controllers/const"
)

func GetServiceDeploymentName(t *melodyiov1alpha1.Inference) string {
	return t.Name + "-" + "deployment"
}

func GetServiceName(t *melodyiov1alpha1.Inference) string {
	return t.Name + "-" + "service"
}

func GetContainerName(t *melodyiov1alpha1.Inference) string {
	return t.Name + "-" + "container"
}

func GetMonitorJobName(t *melodyiov1alpha1.Inference) string {
	return t.Name + "-" + "client-job"
}

func GetServiceEndpoint(t *melodyiov1alpha1.Inference) string {
	return fmt.Sprintf("%s:%d",
		GetServiceName(t),
		consts.InferenceServicePort)
}

func GetDBStorageEndpoint() string {
	return fmt.Sprintf("%s.%s:%s",
		consts.DefaultMelodyDBManagerServiceName,
		consts.DefaultControllerNamespace,
		consts.DefaultMelodyDBManagerServicePort)
}
