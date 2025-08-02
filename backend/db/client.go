package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
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

// Alternative delete by ID
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

// Update credentials
func UpdateCredentials(conn *gorm.DB, obj Connections) error {
	result := conn.Save(&obj)
	if result.Error != nil {
		return fmt.Errorf("failed to update credentials: %w", result.Error)
	}
	return nil
}
