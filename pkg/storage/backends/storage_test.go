package backends

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	api_pb "melody/api/v1alpha1/grpc_proto/grpc_storage/go"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbInterface StorageBackend
)

const dbHost = "10.1.0.20"

func TestMain(m *testing.M) {
	err := os.Setenv("MYSQL_HOST", dbHost)
	if err != nil {
		fmt.Println(err)
	}
	dbInterface = NewMysqlBackendService()
	if err = dbInterface.Initialize(); err != nil {
		fmt.Println(err)
	}
	os.Exit(m.Run())
}

func TestGetDbName(t *testing.T) {
	dbName := "root:melody@tcp(melody-mysql:3306)/melody?timeout=35s"
	dbSource, _, err := GetMysqlDBSource()
	if err != nil {
		t.Errorf("GetMysqlDBSource returns err %v", err)
	}
	if dbSource != dbName {
		t.Errorf("GetMysqlDBSource returns wrong value %v", dbSource)
	}
}

func _TestAddToDB(t *testing.T) {

	testCases := map[string]struct {
		addRequest   *api_pb.SaveResultRequest
		queryRequest *api_pb.GetResultRequest
	}{
		"result_1": {
			addRequest: &api_pb.SaveResultRequest{
				Namespace:     "melody-system",
				InferenceName: "test-inference-1",
				//ExperimentName: "test-pe",
				Results: []*api_pb.KeyValue{{Name: "worker1", Key: "cpu", Value: "20"}},
			},
			queryRequest: &api_pb.GetResultRequest{
				Namespace:     "melody-system",
				InferenceName: "test-inference-1",
				//ExperimentName: "test-pe",
			},
		},
	}

	// Test add row and get row
	var err error
	for name, tc := range testCases {
		t.Run(fmt.Sprintf("%s", name), func(t *testing.T) {
			err = dbInterface.SaveInferenceResult(tc.addRequest)
			if err != nil {
				fmt.Println(err)
			}

			result, err := dbInterface.GetInferenceResult(tc.queryRequest)
			if err != nil {
				fmt.Println(err)
			}
			assert.Equal(t, result.Results[0].Key, tc.addRequest.Results[0].Key)
			assert.Equal(t, result.Results[0].Value, tc.addRequest.Results[0].Value)
		})
	}
}
