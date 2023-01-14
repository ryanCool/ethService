package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// PostgreSQL driver.
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
)

// postgresDB is the concrete PostgresSQL handle to a SQL database.
type postgresDB struct{ *gorm.DB }

// initialize initializes the PostgreSQL database handle.
func (db *postgresDB) initialize(ctx context.Context, cfg dbConfig) {
	// Assemble PostgreSQL database source and setup database handle.
	dbSource := fmt.Sprintf(`host=%s port=%s user=%s password=%s dbname=%s`, cfg.Address, cfg.Port, cfg.Username, cfg.Password,
		cfg.DBName)

	// Connect to the PostgreSQL database.
	var err error
	db.DB, err = gorm.Open(postgres.Open(dbSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

}

// finalize finalizes the PostgreSQL database handle.
func (db *postgresDB) finalize(ctx context.Context) {
	// Close the PostgreSQL database handle.
	postgreDB, _ := db.DB.DB()
	if err := postgreDB.Close(); err != nil {
		fmt.Println("Failed to close database handle: %v", err)
	}
}

// db returns the PostgreSQL GORM database handle.
func (db *postgresDB) db() *gorm.DB {
	return db.DB
}
