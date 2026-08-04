package migrations

import (
	"redirect-service/pkg/logger"
	"redirect-service/platform/database"

	"go.uber.org/zap"
)

func ExecuteMigrations() (err error) {
	conf := database.NewDatabaseConfig()
	err = database.Connect(conf)
	if err != nil {
		logger.Log.Error("failed to connect to database",
			zap.Error(err),
		)
		return
	}

	// AutoMigrate the models
	if err := database.DB.
		AutoMigrate(
		//&models.UserCredentials{},
		); err != nil {
		logger.Log.Fatal("failed to auto-migrate",
			zap.Error(err),
		)
		return err
	}

	logger.Log.Info("Migration ran successfully")

	return nil
}
