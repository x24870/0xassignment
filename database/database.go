package database

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"

	"main/config"
)

// DB is the interface handle to a SQL database.
type DB interface {
	initialize(ctx context.Context, cfg dbConfig)
	finalize()
	db() interface{}
}

// dbConfig is the config to connect to a SQL database.
type dbConfig struct {
	// The dialect of the SQL database.
	Dialect string

	// The username used to login to the database.
	Username string

	// The password used to login to the database.
	Password string

	// The address of the database service to connect to.
	Address string

	// The port of the database service to connect to.
	Port string

	// The name of the database to connect to.
	DBName string
}

// Global database interfaces.
var DBIntf DB

// Database root context.
var dbRootCtx context.Context

// Connection pool configuration
var maxIdleConns int
var maxOpenConns int
var maxConnLifetime time.Duration

func init() {
	maxIdleConns = config.GetInt("DATABASE_MAX_IDLE_CONNECTIONS")
	maxOpenConns = config.GetInt("DATABASE_MAX_OPEN_CONNECTIONS")
	maxConnLifetime = config.GetMilliseconds("DATABASE_MAX_CONN_LIFETIME_MS")
}

// Initialize initializes the database module and instance.
func Initialize(ctx context.Context) {
	// Save database root context.
	dbRootCtx = ctx

	// Create database according to dialect.
	Dialect := config.GetString("DATABASE_DIALECT")
	switch Dialect {
	case "postgres", "cloudsqlpostgres":
		DBIntf = &postgresDB{}
	default:
		panic("invalid dialect")
	}

	// Get database configuration from environment variables.
	DBConfig := dbConfig{
		Dialect:  config.GetString("DATABASE_DIALECT"),
		Username: config.GetString("DATABASE_USERNAME"),
		Password: config.GetString("DATABASE_PASSWORD"),
		Address:  config.GetString("DATABASE_HOST"),
		Port:     config.GetString("DATABASE_PORT"),
		DBName:   config.GetString("DATABASE_NAME"),
	}

	// Initialize the database context.
	DBIntf.initialize(ctx, DBConfig)
}

// Finalize finalizes the database module and closes the database handles.
func Finalize() {
	// Make sure database instance has been initialized.
	if DBIntf == nil {
		panic("database has not been initialized")
	}

	// Finalize database instance.
	DBIntf.finalize()
}

// GetDB returns the database instance.
func GetDB() interface{} {
	return DBIntf.db()
}

// GetSQL returns the SQL database instance.
func GetSQL() *gorm.DB {
	return GetDB().(*gorm.DB)
}

// DBTransactionFunc is the function pointer type to pass to database
// transaction executor functions.
type DBTransactionFunc func(tx *gorm.DB) error
