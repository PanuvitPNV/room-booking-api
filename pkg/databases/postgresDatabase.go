package databases

import (
	"fmt"
	"log"
	"sync"

	"github.com/panuvitpnv/room-booking-api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type postgresDatabase struct {
	*gorm.DB
}

var (
	postgresDatabaseInstance *postgresDatabase
	once                     sync.Once
)

func NewPostgresDatabase(conf *config.Database) Database {
	once.Do(func() {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s search_path=%s",
			conf.Host,
			conf.User,
			conf.Password,
			conf.DBName,
			conf.Port,
			conf.SSLMode,
			conf.Schema,
		)

		gormConfig := &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true, // Disable during migration
			Logger:                                   logger.Default.LogMode(logger.Info),
		}

		conn, err := gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			panic(err)
		}

		log.Printf("Connected to database %s", conf.DBName)
		postgresDatabaseInstance = &postgresDatabase{conn}
	})

	return postgresDatabaseInstance
}

func (db *postgresDatabase) Connect() *gorm.DB {
	if postgresDatabaseInstance == nil {
		log.Fatal("Database connection not initialized")
	}
	return postgresDatabaseInstance.DB
}

// Additional helper methods
func (db *postgresDatabase) Close() error {
	sqlDB, err := postgresDatabaseInstance.DB.DB()
	if err != nil {
		return fmt.Errorf("error getting database instance: %v", err)
	}
	return sqlDB.Close()
}

func (db *postgresDatabase) Ping() error {
	sqlDB, err := postgresDatabaseInstance.DB.DB()
	if err != nil {
		return fmt.Errorf("error getting database instance: %v", err)
	}
	return sqlDB.Ping()
}
