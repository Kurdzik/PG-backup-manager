package backup_manager

import (
	"errors"
	"fmt"
	"log"
	"pg_bckup_mgr/db"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func CreateSchedule(conn *gorm.DB, connectionId string, destinationId string, schedule string) error {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	cronSchedule, err := parser.Parse(schedule)
	if err != nil {
		log.Printf("Invalid cron expression '%s': %v", schedule, err)
		return errors.New("invalid cron expression")
	}

	dest, err := db.GetBackupDestinationByID(conn, destinationId)
	if err != nil {
		log.Printf("Error getting destination: %v", err)
		return err
	}

	creds, err := db.GetCredentialsById(conn, connectionId)
	if err != nil {
		log.Printf("Error getting credentials: %v", err)
		return err
	}

	now := time.Now()
	nextRun := cronSchedule.Next(now)

	newSchedule := db.BackupSchedule{
		ConnectionID:  creds.ID,
		DestinationID: dest.ID,
		Schedule:      schedule,
		Enabled:       true,
		NextRun:       &nextRun,
	}

	if err := conn.Create(&newSchedule).Error; err != nil {
		log.Printf("Error creating schedule: %v", err)
		return err
	}

	log.Printf("Created new schedule: %+v, Next run: %v", newSchedule, nextRun.Format(time.RFC3339))
	return nil
}

func UpdateSchedule(conn *gorm.DB, scheduleId string, updates map[string]interface{}) error {
	id, err := strconv.ParseUint(scheduleId, 10, 32)
	if err != nil {
		log.Printf("Invalid schedule ID: %v", err)
		return errors.New("invalid schedule ID")
	}

	var existingSchedule db.BackupSchedule
	if err := conn.First(&existingSchedule, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Schedule not found with ID: %s", scheduleId)
			return errors.New("schedule not found")
		}
		log.Printf("Error finding schedule: %v", err)
		return err
	}

	allowedFields := map[string]bool{
		"schedule":       true,
		"enabled":        true,
		"connection_id":  true,
		"destination_id": true,
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		} else {
			log.Printf("Warning: Ignoring update to non-allowed field: %s", key)
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("no valid fields to update")
	}

	if newScheduleStr, ok := filteredUpdates["schedule"]; ok {
		parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		cronSchedule, err := parser.Parse(newScheduleStr.(string))
		if err != nil {
			log.Printf("Invalid cron expression '%s': %v", newScheduleStr, err)
			return errors.New("invalid cron expression")
		}

		now := time.Now()
		nextRun := cronSchedule.Next(now)
		filteredUpdates["next_run"] = &nextRun
		log.Printf("Updated next run time to: %v", nextRun.Format(time.RFC3339))
	}

	if err := conn.Model(&existingSchedule).Updates(filteredUpdates).Error; err != nil {
		log.Printf("Error updating schedule: %v", err)
		return err
	}

	log.Printf("Updated schedule ID %s with: %+v", scheduleId, filteredUpdates)
	return nil
}

func DeleteSchedule(conn *gorm.DB, scheduleId string) error {
	id, err := strconv.ParseUint(scheduleId, 10, 32)
	if err != nil {
		log.Printf("Invalid schedule ID: %v", err)
		return errors.New("invalid schedule ID")
	}

	var schedule db.BackupSchedule
	if err := conn.First(&schedule, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Schedule not found with ID: %s", scheduleId)
			return errors.New("schedule not found")
		}
		log.Printf("Error finding schedule: %v", err)
		return err
	}

	if err := conn.Delete(&schedule).Error; err != nil {
		log.Printf("Error deleting schedule: %v", err)
		return err
	}

	log.Printf("Deleted schedule ID: %s", scheduleId)
	return nil
}

func ListSchedules(conn *gorm.DB, filters map[string]interface{}) ([]db.BackupSchedule, error) {
	var schedules []db.BackupSchedule

	query := conn.Model(&db.BackupSchedule{})

	if connectionId, ok := filters["connection_id"]; ok {
		query = query.Where("connection_id = ?", connectionId)
	}

	if destinationId, ok := filters["destination_id"]; ok {
		query = query.Where("destination_id = ?", destinationId)
	}

	if enabled, ok := filters["enabled"]; ok {
		query = query.Where("enabled = ?", enabled)
	}

	query = query.Preload("Connection").Preload("Destination")

	if err := query.Find(&schedules).Error; err != nil {
		log.Printf("Error listing schedules: %v", err)
		return nil, err
	}

	log.Printf("Found %d schedules", len(schedules))
	return schedules, nil
}

func GetScheduleByID(conn *gorm.DB, scheduleId string) (*db.BackupSchedule, error) {
	id, err := strconv.ParseUint(scheduleId, 10, 32)
	if err != nil {
		log.Printf("Invalid schedule ID: %v", err)
		return nil, errors.New("invalid schedule ID")
	}

	var schedule db.BackupSchedule
	if err := conn.Preload("Connection").Preload("Destination").First(&schedule, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Schedule not found with ID: %s", scheduleId)
			return nil, errors.New("schedule not found")
		}
		log.Printf("Error finding schedule: %v", err)
		return nil, err
	}

	return &schedule, nil
}

func ListSchedulesByConnection(conn *gorm.DB, connectionId string) ([]db.BackupSchedule, error) {
	id, err := strconv.ParseUint(connectionId, 10, 32)
	if err != nil {
		log.Printf("Invalid connection ID: %v", err)
		return nil, errors.New("invalid connection ID")
	}

	return ListSchedules(conn, map[string]interface{}{
		"connection_id": uint(id),
	})
}

func ListSchedulesByDestination(conn *gorm.DB, destinationId string) ([]db.BackupSchedule, error) {
	id, err := strconv.ParseUint(destinationId, 10, 32)
	if err != nil {
		log.Printf("Invalid destination ID: %v", err)
		return nil, errors.New("invalid destination ID")
	}

	return ListSchedules(conn, map[string]interface{}{
		"destination_id": uint(id),
	})
}

func EnableSchedule(conn *gorm.DB, scheduleId string) error {
	return UpdateSchedule(conn, scheduleId, map[string]interface{}{
		"enabled": true,
	})
}

func DisableSchedule(conn *gorm.DB, scheduleId string) error {
	return UpdateSchedule(conn, scheduleId, map[string]interface{}{
		"enabled": false,
	})
}

func RecalculateNextRun(conn *gorm.DB, scheduleId string) error {
	schedule, err := GetScheduleByID(conn, scheduleId)
	if err != nil {
		return err
	}

	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	cronSchedule, err := parser.Parse(schedule.Schedule)
	if err != nil {
		log.Printf("Invalid cron expression '%s': %v", schedule.Schedule, err)
		return errors.New("invalid cron expression")
	}

	now := time.Now()
	nextRun := cronSchedule.Next(now)

	if err := conn.Model(schedule).Update("next_run", &nextRun).Error; err != nil {
		log.Printf("Error updating next run time: %v", err)
		return err
	}

	log.Printf("Recalculated next run time for schedule ID %s: %v", scheduleId, nextRun.Format(time.RFC3339))
	return nil
}

var scheduler *cron.Cron
var schedulerMu sync.Mutex

func loadBackupSchedules(conn *gorm.DB) []db.BackupSchedule {
	var schedules []db.BackupSchedule

	if err := conn.Preload("Connection").Preload("Destination").Where("enabled = ?", true).Find(&schedules).Error; err != nil {
		log.Printf("Error loading backup schedules: %v", err)
		return nil
	}

	return schedules
}

func executeBackup(conn *gorm.DB, schedule db.BackupSchedule) {
	log.Printf("Executing backup for schedule ID: %d", schedule.ID)

	updateScheduleRunTimes(conn, schedule)

	manager := BackupManager{
		Host:              schedule.Connection.PostgresHost,
		Port:              schedule.Connection.PostgresPort,
		DBName:            schedule.Connection.PostgresDBName,
		User:              schedule.Connection.PostgresUser,
		Password:          schedule.Connection.PostgresPassword,
		BackupDestination: &schedule.Destination,
	}

	//TODO add local
	manager.CreateBackup(BackupDestination("s3"))

	log.Printf("Backup executed for schedule ID: %d", schedule.ID)
}

func updateScheduleRunTimes(conn *gorm.DB, schedule db.BackupSchedule) {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	cronSchedule, err := parser.Parse(schedule.Schedule)
	if err != nil {
		log.Printf("Invalid cron expression for schedule %d: %v", schedule.ID, err)
		return
	}

	now := time.Now()
	nextRun := cronSchedule.Next(now)

	updates := map[string]interface{}{
		"last_run": &now,
		"next_run": &nextRun,
	}

	if err := conn.Model(&schedule).Updates(updates).Error; err != nil {
		log.Printf("Error updating run times for schedule %d: %v", schedule.ID, err)
		return
	}

	log.Printf("Updated schedule %d: last_run=%s, next_run=%s",
		schedule.ID, now.Format(time.RFC3339), nextRun.Format(time.RFC3339))
}

func RegisterBackupSchedules(conn *gorm.DB) {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()

	if scheduler != nil {
		log.Println("Stopping existing backup scheduler...")
		scheduler.Stop()
	}

	scheduler = cron.New()

	schedules := loadBackupSchedules(conn)
	if schedules == nil {
		log.Println("No backup schedules loaded")
		return
	}

	log.Printf("Registering %d backup schedules", len(schedules))

	for _, schedule := range schedules {
		s := schedule

		log.Printf("Registering backup schedule ID %d with cron: %s", s.ID, s.Schedule)

		_, err := scheduler.AddFunc(s.Schedule, func() {
			executeBackup(conn, s)
		})

		if err != nil {
			log.Printf("Failed to register backup job for schedule %d: %v", s.ID, err)
			continue
		}
	}

	log.Println("All backup schedules registered, scheduler started.")
	scheduler.Start()
}

func StopBackupScheduler() {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()

	if scheduler != nil {
		log.Println("Stopping backup scheduler...")
		scheduler.Stop()
		scheduler = nil
	}
}

func GetSchedulerStatus() map[string]interface{} {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()

	status := map[string]interface{}{
		"running": scheduler != nil,
		"entries": 0,
	}

	if scheduler != nil {
		status["entries"] = len(scheduler.Entries())
	}

	return status
}

func RestartBackupScheduler(conn *gorm.DB) {
	log.Println("Restarting backup scheduler...")
	StopBackupScheduler()
	RegisterBackupSchedules(conn)
}

func AddSingleSchedule(conn *gorm.DB, scheduleID uint) error {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()

	if scheduler == nil {
		return fmt.Errorf("scheduler is not running")
	}

	var schedule db.BackupSchedule
	if err := conn.Preload("Connection").Preload("Destination").First(&schedule, scheduleID).Error; err != nil {
		return fmt.Errorf("error loading schedule %d: %v", scheduleID, err)
	}

	if !schedule.Enabled {
		return fmt.Errorf("schedule %d is disabled", scheduleID)
	}

	_, err := scheduler.AddFunc(schedule.Schedule, func() {
		executeBackup(conn, schedule)
	})

	if err != nil {
		return fmt.Errorf("failed to add schedule %d to scheduler: %v", scheduleID, err)
	}

	log.Printf("Added single schedule %d to running scheduler", scheduleID)
	return nil
}
