package db

import (
	"time"
)

type Connection struct {
	ID               uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	PostgresHost     string    `json:"postgres_host" gorm:"type:varchar(255);not null"`
	PostgresPort     string    `json:"postgres_port" gorm:"not null;default:5432"`
	PostgresDBName   string    `json:"postgres_db_name" gorm:"type:varchar(255);not null"`
	PostgresUser     string    `json:"postgres_user" gorm:"type:varchar(255);not null"`
	PostgresPassword string    `json:"postgres_password" gorm:"type:varchar(255);not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Destination     []Destination    `json:"destinations,omitempty" gorm:"foreignKey:ConnectionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	BackupSchedules []BackupSchedule `json:"backup_schedules,omitempty" gorm:"foreignKey:ConnectionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

// TableName overrides the table name used by Connection to `connections`
func (Connection) TableName() string {
	return "connections"
}

type Destination struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ConnectionID    uint      `json:"connection_id" gorm:"not null;index"`
	Name            string    `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
	EndpointURL     string    `json:"endpoint_url" gorm:"type:varchar(500);not null"`
	Region          string    `json:"region" gorm:"type:varchar(100)"`
	BucketName      string    `json:"bucket_name" gorm:"type:varchar(255);not null"`
	AccessKeyID     string    `json:"access_key_id" gorm:"type:varchar(255);not null"`
	SecretAccessKey string    `json:"secret_access_key" gorm:"type:varchar(255);not null"`
	PathPrefix      string    `json:"path_prefix" gorm:"type:varchar(500);default:''"`
	UseSSL          bool      `json:"use_ssl" gorm:"default:true"`
	VerifySSL       bool      `json:"verify_ssl" gorm:"default:true"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Connection      Connection       `json:"connection,omitempty" gorm:"foreignKey:ConnectionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	BackupSchedules []BackupSchedule `json:"backup_schedules,omitempty" gorm:"foreignKey:DestinationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

// TableName overrides the table name used by Destination to `destinations`
func (Destination) TableName() string {
	return "destinations"
}

type BackupSchedule struct {
	ID            uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	ConnectionID  uint       `json:"connection_id" gorm:"not null;index"`
	DestinationID uint       `json:"destination_id" gorm:"not null;index"`
	Schedule      string     `json:"schedule" gorm:"type:varchar(255);not null"` // Cron expression
	Enabled       bool       `json:"enabled" gorm:"default:true;index"`
	LastRun       *time.Time `json:"last_run,omitempty"`
	NextRun       *time.Time `json:"next_run,omitempty" gorm:"index"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Connection  Connection  `json:"connection,omitempty" gorm:"foreignKey:ConnectionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Destination Destination `json:"destination,omitempty" gorm:"foreignKey:DestinationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

// TableName overrides the table name used by BackupSchedule to `backup_schedules`
func (BackupSchedule) TableName() string {
	return "backup_schedules"
}
