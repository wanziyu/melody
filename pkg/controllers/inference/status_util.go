package inference

import melodyiov1alpha1 "melody/api/v1alpha1"

type updateStatusFunc func(instance *melodyiov1alpha1.Inference) error
