package backup_manager

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"pg_bckup_mgr/auth"
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
	decryptedPassword, _ := auth.DecryptString(b.Password)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGPASSWORD=%s", decryptedPassword))

	return cmd.Run()
}

func (b BackupManager) Connect() (*gorm.DB, error) {
	decryptedPassword, _ := auth.DecryptString(b.Password)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		b.Host, b.User, decryptedPassword, b.DBName, b.Port)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return conn, nil

}

func (b BackupManager) ListAvaiableBackups(destination BackupDestination) []string {

	switch destination {
	case BackupFilesystem:
		log.Println("Searching for backups in local filesystem...")

		backupDirName := fmt.Sprintf("%s-%s-%s", b.DBName, b.Host, b.User)
		files, err := os.ReadDir(fmt.Sprintf("%s/%s", LOCAL_BACKUP_DIR, backupDirName))

		if err != nil {
			log.Println("Error occured: ", err.Error())
			return []string{}
		}

		filenames := []string{}
		for _, file := range files {
			filenames = append(filenames, file.Name())
		}
		log.Println("found: ", len(filenames), " files")

		return filenames

	case BackupS3Bucket:
		log.Println("Searching for backups in S3 bucket...")

		S3Client, _ := NewS3Client(b.BackupDestination.Name,
			b.BackupDestination.EndpointURL,
			b.BackupDestination.Region,
			b.BackupDestination.BucketName,
			b.BackupDestination.AccessKeyID,
			b.BackupDestination.SecretAccessKey,
			b.BackupDestination.UseSSL,
			b.BackupDestination.VerifySSL,
		)

		filenames, err := S3Client.ListFiles()
		if err != nil {
			return []string{}
		}
		log.Println("found: ", len(filenames), " files")

		return filenames
	}

	log.Println("Unable to list backups from ", destination)
	return []string{}
}

func (b BackupManager) RestoreFromBackup(destination BackupDestination, filename string) error {
	switch destination {
	case BackupFilesystem:
		log.Printf("Restoring database from backup: %s", filename)
		backupDirName := fmt.Sprintf("%s-%s-%s", b.DBName, b.Host, b.User)
		backupPath := filepath.Join(LOCAL_BACKUP_DIR, backupDirName, filename)

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

		decryptedPassword, _ := auth.DecryptString(b.Password)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("PGPASSWORD=%s", decryptedPassword))

		if err := cmd.Run(); err != nil {
			log.Printf("Error restoring backup: %v", err)
			return fmt.Errorf("restore failed: %v", err)
		}

		log.Printf("Successfully restored database from backup: %s", filename)
		return nil

	case BackupS3Bucket:
		log.Printf("Restoring database from S3 backup: %s", filename)

		S3Client, err := NewS3Client(b.BackupDestination.Name,
			b.BackupDestination.EndpointURL,
			b.BackupDestination.Region,
			b.BackupDestination.BucketName,
			b.BackupDestination.AccessKeyID,
			b.BackupDestination.SecretAccessKey,
			b.BackupDestination.UseSSL,
			b.BackupDestination.VerifySSL,
		)
		if err != nil {
			log.Printf("Error creating S3 client: %v", err)
			return fmt.Errorf("S3 client creation failed: %v", err)
		}

		// Create temporary directory for download
		backupDirName := fmt.Sprintf("%s-%s-%s", b.DBName, b.Host, b.User)
		os.MkdirAll(fmt.Sprintf("%s/%s/", LOCAL_BACKUP_DIR, backupDirName), 0755)

		tempBackupPath := filepath.Join(LOCAL_BACKUP_DIR, backupDirName, filename)
		log.Printf("Downloading backup from S3 to: %s", tempBackupPath)

		// Download file from S3
		err = S3Client.DownloadFile(filename, tempBackupPath)
		if err != nil {
			log.Printf("Error downloading backup from S3: %v", err)
			return fmt.Errorf("S3 download failed: %v", err)
		}
		log.Println("Successfully downloaded backup from S3")

		// Verify downloaded file exists
		if _, err := os.Stat(tempBackupPath); os.IsNotExist(err) {
			log.Printf("Downloaded backup file does not exist: %s", tempBackupPath)
			return fmt.Errorf("downloaded backup file not found: %s", filename)
		}

		conn, err := b.Connect()
		if err != nil {
			log.Printf("Unable to connect to database for restore: %v", err)
			return fmt.Errorf("database connection failed: %v", err)
		}

		db, _ := conn.DB()
		db.Close()

		log.Println("Starting database restore from downloaded backup...")
		cmd := exec.Command("pg_restore",
			"-h", b.Host,
			"-p", b.Port,
			"-U", b.User,
			"-d", b.DBName,
			"-c",
			"--if-exists",
			"-v",
			tempBackupPath,
		)

		decryptedPassword, _ := auth.DecryptString(b.Password)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("PGPASSWORD=%s", decryptedPassword))

		if err := cmd.Run(); err != nil {
			log.Printf("Error restoring backup: %v", err)
			// Clean up temp file
			os.Remove(tempBackupPath)
			return fmt.Errorf("restore failed: %v", err)
		}

		// Clean up temp file
		log.Println("Cleaning up temporary backup file...")
		os.Remove(tempBackupPath)

		log.Printf("Successfully restored database from S3 backup: %s", filename)
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

	timestamp := time.Now().Format("20060102_150405")
	backupDirName := fmt.Sprintf("%s-%s-%s", b.DBName, b.Host, b.User)
	os.MkdirAll(fmt.Sprintf("%s/%s/", LOCAL_BACKUP_DIR, backupDirName), 0755)
	outputFile := fmt.Sprintf("%s/%s/backup_%s.dump", LOCAL_BACKUP_DIR, backupDirName, timestamp)
	err = b.createPgDumpBackup(outputFile)
	if err != nil {
		log.Println("Error occurred: \n\n", err.Error())
		return err
	}

	switch destination {
	case BackupFilesystem:
		log.Println("Backing up database to a local filesystem...")
		return nil

	case BackupS3Bucket:
		log.Println("Backing up database to a remote S3 bucket...")

		S3Client, _ := NewS3Client(b.BackupDestination.Name,
			b.BackupDestination.EndpointURL,
			b.BackupDestination.Region,
			b.BackupDestination.BucketName,
			b.BackupDestination.AccessKeyID,
			b.BackupDestination.SecretAccessKey,
			b.BackupDestination.UseSSL,
			b.BackupDestination.VerifySSL,
		)

		err = S3Client.UploadFile(outputFile)
		if err != nil {
			log.Println("Error Ocurred durig file upload ", err.Error())
		}

		os.RemoveAll(outputFile)
		return nil
	}

	log.Println("Unable to backup database to ", destination)
	return err

}

func (b BackupManager) DeleteBackup(destination BackupDestination, filename string) error {
	switch destination {
	case BackupFilesystem:
		log.Printf("Deleting backup file: %s", filename)

		backupDirName := fmt.Sprintf("%s-%s-%s", b.DBName, b.Host, b.User)
		backupPath := filepath.Join(LOCAL_BACKUP_DIR, backupDirName, filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			log.Printf("Backup file does not exist: %s", backupPath)
			return fmt.Errorf("backup file not found: %s", filename)
		}

		if err := os.Remove(backupPath); err != nil {
			log.Printf("Error deleting backup file: %v", err)
			return fmt.Errorf("failed to delete backup file: %v", err)
		}

		log.Printf("Successfully deleted backup file: %s", filename)
		return nil

	case BackupS3Bucket:
		log.Printf("Deleting backup file from S3: %s", filename)

		S3Client, err := NewS3Client(b.BackupDestination.Name,
			b.BackupDestination.EndpointURL,
			b.BackupDestination.Region,
			b.BackupDestination.BucketName,
			b.BackupDestination.AccessKeyID,
			b.BackupDestination.SecretAccessKey,
			b.BackupDestination.UseSSL,
			b.BackupDestination.VerifySSL,
		)
		if err != nil {
			log.Printf("Error creating S3 client: %v", err)
			return fmt.Errorf("S3 client creation failed: %v", err)
		}

		log.Println("Attempting to delete file from S3 bucket...")
		err = S3Client.DeleteFile(filename)
		if err != nil {
			log.Printf("Error deleting backup from S3: %v", err)
			return fmt.Errorf("failed to delete S3 backup: %v", err)
		}

		log.Printf("Successfully deleted backup file from S3: %s", filename)
		return nil
	}

	log.Printf("Unable to delete backup from destination: %s", destination)
	return fmt.Errorf("unsupported backup destination: %s", destination)
}
