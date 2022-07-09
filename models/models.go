package models

import "github.com/jinzhu/gorm"

// Model ...
type model interface {
	createIndexes(db *gorm.DB) error
	createUniqueIndexes(db *gorm.DB) error
	createForeignKeys(db *gorm.DB) error
}

// models ...
var models = []model{}

// registerModelForMigration...
func registerModelForAutoMigration(model model) {
	models = append(models, model)
}

// AutoMigrate ...
func AutoMigrate(db *gorm.DB) error {
	// Turn on logging for migration.
	db = db.Debug()

	// Perform migration on models.
	for _, model := range models {
		err := db.AutoMigrate(model).Error
		if err != nil {
			return err
		}
	}

	// Create indexes, and foreign keys for each model.
	for _, model := range models {
		// Create indexes.
		if err := model.createIndexes(db); err != nil {
			return err
		}

		// Create unique indexes.
		if err := model.createUniqueIndexes(db); err != nil {
			return err
		}

		// Create foreign keys.
		if err := model.createForeignKeys(db); err != nil {
			return err
		}
	}

	return nil
}
