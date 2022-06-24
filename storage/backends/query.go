package backends

import (
	"github.com/jinzhu/gorm"
)

type MonitoringResult struct {
	//gorm.Model
	Namespace     string `gorm:"type:varchar(128);column:namespace" json:"namespace"`
	InferenceName string `gorm:"type:varchar(128);column:inference_name" json:"inference_name"`
	//ExperimentName string    `gorm:"type:varchar(128);column:experiment_name" json:"experiment_name"`
	NodeName string `gorm:"type:varchar(128);column:nodename" json:"nodename"`
	Key      string `gorm:"type:varchar(128);column:key" json:"key"`
	Value    string `gorm:"type:varchar(128);column:value" json:"value"`
	//GmtModified    time.Time `gorm:"type:datetime;column:gmt_modified" json:"gmt_modified"`
}

func (tr MonitoringResult) TableName() string {
	return "monitoring_result_info"
}

// BeforeCreate update gmt_modified timestamp.
func (tr *MonitoringResult) BeforeCreate(scope *gorm.Scope) error {
	return nil //scope.SetColumn("gmt_modified", time.Now().UTC())
}

// BeforeUpdate update gmt_modified timestamp.
func (tr *MonitoringResult) BeforeUpdate(scope *gorm.Scope) error {
	return nil //scope.SetColumn("gmt_modified", time.Now().UTC())
}
