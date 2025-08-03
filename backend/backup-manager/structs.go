package backup_manager

import "pg_bckup_mgr/db"

const LOCAL_BACKUP_DIR = "/etc/backups"

type BackupDestination string

const (
	BackupFilesystem BackupDestination = "local"
	BackupS3Bucket   BackupDestination = "s3"
)

type BackupManager struct {
	Host              string
	Port              string
	DBName            string
	User              string
	Password          string
	BackupDestination *db.Destination
}
