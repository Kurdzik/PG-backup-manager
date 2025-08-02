package backup_manager

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (b BackupManager) createPgDumpBackup(outputPath string) error {
	cmd := exec.Command("pg_dump",
		"-h", b.Host,
		"-p", b.Port,
		"-U", b.User,
		"-d", b.DBName,
		"-W",
		"-Fc",
		"-f", outputPath,
	)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGPASSWORD=%s", b.Password))

	log.Println(cmd)
	return cmd.Run()
}

func (b BackupManager) Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		b.Host, b.User, b.Password, b.DBName, b.Port)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return conn, nil

}

func (b BackupManager) ListAvaiableBackups(destination BackupDestination) []string {
	switch destination {
	case BackupFilesystem:
		files, err := os.ReadDir(fmt.Sprintf("%s/%s", LOCAL_BACKUP_DIR, b.DBName))
		if err != nil {
			return []string{}
		}

		filenames := []string{}
		for _, file := range files {
			filenames = append(filenames, file.Name())
		}
		return filenames
	}

	log.Println("Unable to list backups from ", destination)
	return []string{}
}

func (b BackupManager) RestoreFromBackup(destination BackupDestination, filename string) error {
	switch destination {
	case BackupFilesystem:
		log.Printf("Restoring database from backup: %s", filename)

		backupPath := filepath.Join(LOCAL_BACKUP_DIR, b.DBName, filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			log.Printf("Backup file does not exist: %s", backupPath)
			return fmt.Errorf("backup file not found: %s", filename)
		}

		conn, err := b.Connect()
		if err != nil {
			log.Printf("Unable to connect to database for restore: %v", err)
			return fmt.Errorf("database connection failed: %v", err)
		}

		db, _ := conn.DB()
		db.Close()

		cmd := exec.Command("pg_restore",
			"-h", b.Host,
			"-p", b.Port,
			"-U", b.User,
			"-d", b.DBName,
			"-c",
			"--if-exists",
			"-v",
			backupPath,
		)

		cmd.Env = append(os.Environ(),
			fmt.Sprintf("PGPASSWORD=%s", b.Password))

		if err := cmd.Run(); err != nil {
			log.Printf("Error restoring backup: %v", err)
			return fmt.Errorf("restore failed: %v", err)
		}

		log.Printf("Successfully restored database from backup: %s", filename)
		return nil
	}

	log.Printf("Unable to restore backup from destination: %s", destination)
	return fmt.Errorf("unsupported backup destination: %s", destination)
}

func (b BackupManager) CreateBackup(destination BackupDestination) error {

	conn, err := b.Connect()
	if err != nil {
		log.Printf("Unable to connect to a database")
		return err
	}

	db, _ := conn.DB()
	db.Close()

	switch destination {
	case BackupFilesystem:
		log.Println("Backing up database to a local filesystem...")

		timestamp := time.Now().Format("20060102_150405")

		os.MkdirAll(fmt.Sprintf("%s/%s/", LOCAL_BACKUP_DIR, b.DBName), 0755)
		outputFile := fmt.Sprintf("%s/%s/backup_%s.dump", LOCAL_BACKUP_DIR, b.DBName, timestamp)

		err = b.createPgDumpBackup(outputFile)
		if err != nil {
			log.Println("Error occurred: \n\n", err.Error())
			return err
		}
		return nil
	}

	log.Println("Unable to backup database to ", destination)
	return err

}
