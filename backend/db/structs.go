package db

import "time"

type Connections struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	PostgresHost     string    `json:"postgres_host" gorm:"column:postgres_host"`
	PostgresPort     string    `json:"postgres_port" gorm:"column:postgres_port"`
	PostgresDBName   string    `json:"postgres_db_name" gorm:"column:postgres_db_name"`
	PostgresUser     string    `json:"postgres_user" gorm:"column:postgres_user"`
	PostgresPassword string    `json:"postgres_password" gorm:"column:postgres_password"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
