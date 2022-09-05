package inference

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	melodyiov1alpha1 "melody/api/v1alpha1"
	consts "melody/controllers/const"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *InferenceReconciler) GetPodsForJob(job *melodyiov1alpha1.Inference) ([]*v1.Pod, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: r.GenLabels(job.Name),
	})
	// List all pods to include those that don't match the selector anymore
	// but have a ControllerRef pointing to this controller.
	podList := &v1.PodList{}
	err = r.Client.List(context.Background(), podList, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return nil, err
	}
	return ToPodPointerList(podList.Items), nil
}

func (r *InferenceReconciler) GenLabels(jobName string) map[string]string {
	labelInferenceName := consts.LabelInferenceName
	labelDeploymentName := consts.LabelDeploymentName
	return map[string]string{
		labelInferenceName:  jobName,
		labelDeploymentName: GenDeploymentName(jobName),
	}
}

func GenDeploymentName(jobName string) string {
	return jobName + "-" + consts.LabelDeploymentName
}

func ToPodPointerList(list []v1.Pod) []*v1.Pod {
	if list == nil {
		return nil
	}
	ret := make([]*v1.Pod, 0, len(list))
	for i := range list {
		ret = append(ret, &list[i])
	}
	return ret
}
