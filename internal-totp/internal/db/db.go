package db

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"internal-totp/internal/models"
)

var DB *gorm.DB

func Init(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// automigrate
	if err := DB.AutoMigrate(&models.MFATOTPSeed{}, &models.MFABackupCode{}, &models.MFAAuditLog{}); err != nil {
		return err
	}
	log.Println("DB connected and migrated")
	return nil
}
