package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestConnection(conn Connection) bool {
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

	err = conn.AutoMigrate(&Connection{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return conn, nil
}

func AddCredentials(conn *gorm.DB, obj Connection) error {
	result := conn.Create(&obj)
	if result.Error != nil {
		return fmt.Errorf("failed to create credentials: %w", result.Error)
	}
	return nil
}

func GetCredentialsById(conn *gorm.DB, id string) (Connection, error) {
	var connection Connection
	result := conn.First(&connection, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return connection, fmt.Errorf("credentials with id %s not found", id)
		}
		return connection, fmt.Errorf("failed to get credentials: %w", result.Error)
	}
	return connection, nil
}

func ListAllCredentials(conn *gorm.DB) ([]Connection, error) {
	var connections []Connection
	result := conn.Find(&connections)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", result.Error)
	}
	return connections, nil
}

func DeleteCredentialsById(conn *gorm.DB, id string) error {
	result := conn.Delete(&Connection{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete credentials: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("credentials with id %s not found", id)
	}
	return nil
}

func UpdateCredentials(conn *gorm.DB, obj Connection) error {
	result := conn.Save(&obj)
	if result.Error != nil {
		return fmt.Errorf("failed to update credentials: %w", result.Error)
	}
	return nil
}

func GetBackupDestinationByID(conn *gorm.DB, id string) (Destination, error) {
	var destinations Destination
	result := conn.First(&destinations, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return destinations, fmt.Errorf("backup destinations with id %s not found", id)
		}
		return destinations, fmt.Errorf("failed to get backup destinations: %w", result.Error)
	}
	return destinations, nil
}
