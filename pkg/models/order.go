package models

type Order struct {
	ID                int    `json:"id"`
	UserID            string `json:"user_id" validate:"required,uuid"`
	ClusterName       string `json:"cluster_name" validate:"required,min=1,max=63,isvalidclustername"`
	HasControlPlane   bool   `json:"has_control_plane"`
	HasMonitoring     bool   `json:"has_monitoring"`
	HasAlerting       bool   `json:"has_alerting"`
	ImageStorage      int    `json:"images_storage" validate:"required"`
	MonitoringStorage int    `json:"monitoring_storage" validate:"required"`
}
