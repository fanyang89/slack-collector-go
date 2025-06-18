package main

import (
	"time"

	zerologgorm "github.com/go-mods/zerolog-gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time { return time.Now() },
		Logger: &zerologgorm.GormLogger{
			FieldsExclude: []string{zerologgorm.DurationFieldName, zerologgorm.FileFieldName},
		},
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Message{}, &Channel{}, &User{}, &Collected{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
