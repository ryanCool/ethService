package database

import (
	"context"
	"fmt"
	"github.com/ryanCool/ethService/config"
	"runtime/debug"

	"gorm.io/gorm"
)

// DB is the interface handle to a SQL database.
type DB interface {
	initialize(ctx context.Context, cfg dbConfig)
	finalize(ctx context.Context)
	db() *gorm.DB
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

// Global database interface.
var dbIntf DB

// Initialize initializes the database module and instance.
func Initialize(ctx context.Context) {
	// Create database according to dialect.
	dialect := config.GetString("DATABASE_DIALECT")
	switch dialect {
	case "postgres", "cloudsqlpostgres":
		dbIntf = &postgresDB{}
	default:
		panic("invalid dialect")
	}

	// Get database configuration from environment variables.
	cfg := dbConfig{
		Dialect:  config.GetString("DATABASE_DIALECT"),
		Username: config.GetString("DATABASE_USERNAME"),
		Password: config.GetString("DATABASE_PASSWORD"),
		Address:  config.GetString("DATABASE_HOST"),
		Port:     config.GetString("DATABASE_PORT"),
		DBName:   config.GetString("DATABASE_NAME"),
	}

	// Initialize the database context.
	dbIntf.initialize(ctx, cfg)
}

// Finalize finalizes the database module and closes the database handles.
func Finalize(ctx context.Context) {
	// Make sure database instance has been initialized.
	if dbIntf == nil {
		panic("database has not been initialized")
	}

	// Finalize database instance.
	dbIntf.finalize(ctx)
}

// GetDB returns the GORM database instance.
func GetDB() *gorm.DB {
	return dbIntf.db()
}

// DBTransactionFunc is the function pointer type to pass to database
// transaction executor functions.
type DBTransactionFunc func(ctx context.Context, tx *gorm.DB) error

// Transaction executes the provided function as a transaction,
// and automatically performs commit / rollback accordingly.
func Transaction(ctx context.Context, db *gorm.DB, txFunc DBTransactionFunc) (err error) {
	// Obtain transaction handle.
	if err := db.Exec(`set transaction isolation level READ UNCOMMITTED`).Error; err != nil {
		return err
	}

	tx := db.Begin()
	if err = tx.Error; err != nil {
		fmt.Println("Failed to begin transaction: %v", err)
		return err
	}

	// Defer commit / rollback before we execute transaction.
	defer func() {
		// Recover from panic.
		var recovered interface{}
		if recovered = recover(); recovered != nil {
			// Assemble log string.
			message := fmt.Sprintf("\x1b[31m%v\n[Stack Trace]\n%s\x1b[m",
				recovered, debug.Stack())

			// Record the stack trace to logging service, or if we cannot
			// find a logging from this request, use the static logging.
			fmt.Println(message)
		}

		// Perform rollback if panic or if error is encountered.
		if recovered != nil || err != nil {
			if rerr := tx.Rollback().Error; rerr != nil {
				fmt.Println("Failed to rollback transaction: %v", rerr)
			}
		}
	}()

	// Execute transaction.
	if err = txFunc(ctx, tx); err != nil {
		fmt.Println("Failed to execute transaction: %v", err)
		return err
	}

	// Commit transaction.
	if err = tx.Commit().Error; err != nil {
		fmt.Println("Failed to commit transaction: %v", err)
		return err
	}

	return nil
}
