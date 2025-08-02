package db

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestConnection(conn Connections) bool {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		conn.PostgresUser,
		conn.PostgresPassword,
		conn.PostgresHost,
		conn.PostgresPort,
		conn.PostgresDBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return false
	}

	sqlDB, err := db.DB()
	if err != nil {
		return false
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return false
	}

	return true
}

func Connect() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = conn.AutoMigrate(&Connections{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return conn, nil
}

func AddCredentials(conn *gorm.DB, obj Connections) error {
	result := conn.Create(&obj)
	if result.Error != nil {
		return fmt.Errorf("failed to create credentials: %w", result.Error)
	}
	return nil
}

func GetCredentialsById(conn *gorm.DB, id int) (Connections, error) {
	var connection Connections
	result := conn.First(&connection, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return connection, fmt.Errorf("credentials with id %d not found", id)
		}
		return connection, fmt.Errorf("failed to get credentials: %w", result.Error)
	}
	return connection, nil
}

func ListAllCredentials(conn *gorm.DB) ([]Connections, error) {
	var connections []Connections
	result := conn.Find(&connections)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", result.Error)
	}
	return connections, nil
}

func DeleteCredentialsById(conn *gorm.DB, id int) error {
	result := conn.Delete(&Connections{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete credentials: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("credentials with id %d not found", id)
	}
	return nil
}

func UpdateCredentials(conn *gorm.DB, obj Connections) error {
	result := conn.Save(&obj)
	if result.Error != nil {
		return fmt.Errorf("failed to update credentials: %w", result.Error)
	}
	return nil
}

type Destinations struct {
	ID              uint      `json:"destination_id" gorm:"primaryKey;autoIncrement"`
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
}

func GetBackupDestinationByID(conn *gorm.DB, id string) (Destinations, error) {
	var destinations Destinations
	result := conn.First(&destinations, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return destinations, fmt.Errorf("backup destinations with id %d not found", id)
		}
		return destinations, fmt.Errorf("failed to get backup destinations: %w", result.Error)
	}
	return destinations, nil
}
