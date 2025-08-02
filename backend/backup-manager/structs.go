package backup_manager

const LOCAL_BACKUP_DIR = "/etc/backups"

type BackupDestination string

const (
	BackupFilesystem BackupDestination = "local"
)

type BackupManager struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
}
