package core

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// newDB creates a new *gorm.DB instance and migrate the models
func newDB() *gorm.DB {
	logrus.Info("start mysql")
	url := viper.GetString("mysql")
	if len(url) == 0 {
		logrus.Panic("mysql url is empty")
	}
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		logrus.WithError(err).Panic("failed to create db")
	}
	if err := db.AutoMigrate(&Proxy{}); err != nil {
		logrus.WithError(err).Panic("failed to migrate model")
	}
	return db
}
