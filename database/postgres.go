package database

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
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

	//todo turn off log
	db.DB, err = gorm.Open(postgres.Open(dbSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		//Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

}

// finalize finalizes the PostgreSQL database handle.
func (db *postgresDB) finalize(ctx context.Context) {
	// Close the PostgreSQL database handle.
	d, _ := db.DB.DB()
	if err := d.Close(); err != nil {
		log.Printf("Failed to close database handle: %v\n", err)
	}
}

// db returns the PostgreSQL GORM database handle.
func (db *postgresDB) db() *gorm.DB {
	return db.DB
}
