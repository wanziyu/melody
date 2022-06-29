package backends

import (
	api_pb "melody/api/v1alpha1/grpc_proto/grpc_storage/go"
)

// StorageBackend provides a collection of abstract methods to
// interact with different storage backends, write/read pod and job objects.
type StorageBackend interface {
	// Initialize initializes a backend storage service with local or remote database.
	Initialize() error
	// Close shutdown backend storage service.
	Close() error
	// Name returns backend name.
	Name() string
	// SaveInferenceResult append or update a pod record to backend.
	SaveInferenceResult(observationLog *api_pb.SaveResultRequest) error
	// GetTrialResult retrieve a TrialResult from backend.
	GetInferenceResult(request *api_pb.GetResultRequest) (*api_pb.GetResultReply, error)
}
