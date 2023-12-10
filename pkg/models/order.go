package models

type Order struct {
	ID              int    `json:"id"`
	UserID          int    `json:"user_id"`
	ClusterName     string `json:"cluster_name"`
	HasControlPlane bool   `json:"has_control_plane"`
	HasMonitoring   bool   `json:"has_monitoring"`
	HasAlerting     bool   `json:"has_alerting"`
	StorageSize     int    `json:"storage_size"`
}
