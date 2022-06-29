package backends

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"k8s.io/klog"
	api_pb "melody/api/v1alpha1/grpc_proto/grpc_storage/go"
	"sync/atomic"
	"time"
)

const (
	initListSize = 512
	dbDriver     = "mysql"
	initInterval = 5 * time.Second
	initTimeout  = 60 * 10 * time.Second
)

func NewMysqlBackendService() StorageBackend {
	return &MysqlBackend{initialized: 0}
}

func NewMysqlBackend(mockDB *gorm.DB) (*MysqlBackend, error) {
	mysql := &MysqlBackend{initialized: 0}
	mysql.db = mockDB
	// Try create tables if they have not been created in database, or the storage service will not work.
	if !mysql.db.HasTable(&MonitoringResult{}) {
		klog.Infof("database has not table %s, try to create it", MonitoringResult{}.TableName())
		err := mysql.db.CreateTable(&MonitoringResult{}).Error
		if err != nil {
			return nil, err
		}
	}
	atomic.StoreInt32(&mysql.initialized, 1)
	return mysql, nil
}

var _ StorageBackend = &MysqlBackend{}

type MysqlBackend struct {
	db          *gorm.DB
	initialized int32
}

func (b *MysqlBackend) Initialize() error {
	if atomic.LoadInt32(&b.initialized) == 1 {
		return nil
	}
	err1 := b.init()
	if err1 != nil {
		klog.Errorf("Error Initialize DB: %v", err1)
		return err1
	}
	atomic.StoreInt32(&b.initialized, 1)
	return nil
}

func (b *MysqlBackend) Close() error {
	if b.db == nil {
		return nil
	}
	return b.db.Commit().Close()
}

func (b *MysqlBackend) Name() string {
	return "mysql"
}

func (b *MysqlBackend) SaveInferenceResult(request *api_pb.SaveResultRequest) error {
	klog.V(5).Infof("[mysql.SaveInferenceResult] namespace: %s, inference: %s", request.Namespace, request.InferenceName)

	existingResult := MonitoringResult{}
	saveQuery := &MonitoringResult{
		Namespace:     request.Namespace,
		InferenceName: request.InferenceName,
		//ExperimentName: request.ExperimentName,
	}

	result := b.db.Where(saveQuery).First(&existingResult)

	if request.Results != nil {
		saveQuery.Key = request.Results[0].Key
		saveQuery.Value = request.Results[0].Value
	}

	if result.Error != nil {
		if gorm.IsRecordNotFoundError(result.Error) {

			return b.createNewResult(saveQuery)
		}
		return result.Error
	}
	klog.Errorf("createNewResult error: %v", result.Error)
	return b.updateNewResult(saveQuery)
}

func (b *MysqlBackend) createNewResult(newResult *MonitoringResult) error {
	err := b.db.Create(newResult).Error
	if err != nil {
		klog.Errorf("saveInferenceResult error: %v", err)
	}
	return err
}

func (b *MysqlBackend) updateNewResult(newResult *MonitoringResult) error {
	result := b.db.Model(&MonitoringResult{}).Where(&MonitoringResult{
		Namespace:     newResult.Namespace,
		InferenceName: newResult.InferenceName,
		//ExperimentName: newResult.ExperimentName,
	}).Updates(newResult)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (b *MysqlBackend) GetInferenceResult(request *api_pb.GetResultRequest) (*api_pb.GetResultReply, error) {
	klog.V(5).Infof("[mysql.GetInferenceResult] namespace: %s, inference: %s", request.Namespace, request.InferenceName)
	existingResult := MonitoringResult{}
	getQuery := &MonitoringResult{
		Namespace:     request.Namespace,
		InferenceName: request.InferenceName,
		//ExperimentName: request.ExperimentName,
	}

	result := b.db.Where(getQuery).First(&existingResult)
	if result.Error != nil {
		return nil, result.Error
	}

	reply := &api_pb.GetResultReply{
		Namespace:     existingResult.Namespace,
		InferenceName: existingResult.InferenceName,
		//ExperimentName: existingResult.ExperimentName,
		Results: []*api_pb.KeyValue{{Key: existingResult.Key, Value: existingResult.Value}},
	}

	return reply, nil
}

func (b *MysqlBackend) openMysqlConnection(dbDriver, dbSource string) (db *gorm.DB, err error) {
	ticker := time.NewTicker(initInterval)
	defer ticker.Stop()
	timeoutC := time.After(initTimeout)
	for {
		select {
		case <-ticker.C:
			if db, err := gorm.Open(dbDriver, dbSource); err == nil {
				klog.Infof("Mysql db connected")
				return db, nil
			} else {
				klog.Infof("Open sql connection failed: %v", err)
			}
		case <-timeoutC:
			klog.Errorf("Open mysql connection failed (timeout)")
			return nil, fmt.Errorf("open mysql connection failed (timeout)")
		}
	}
}

func (b *MysqlBackend) init() error {
	dbSource, logMode, err := GetMysqlDBSource()
	if err != nil {
		klog.Errorf("Error init DB: %v", err)
		return err
	}
	if b.db, err = b.openMysqlConnection(dbDriver, dbSource); err != nil { // gorm.Open(dbDriver, dbSource)
		klog.Errorf("Error Open DB: %v", err)
		return err
	}
	b.db.LogMode(logMode == "debug")

	// Try create tables if they have not been created in database, or the storage service will not work.
	if !b.db.HasTable(&MonitoringResult{}) {
		klog.Infof("database has not table %s, try to create it", MonitoringResult{}.TableName())
		err = b.db.CreateTable(&MonitoringResult{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}
