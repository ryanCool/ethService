package database

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/ryanCool/ethService/config"
	"gorm.io/gorm/logger"
	"time"

	"gorm.io/gorm"
	// PostgreSQL driver.
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
)

// postgresDB is the concrete PostgresSQL handle to a SQL database.
type postgresDB struct{ *gorm.DB }

var maxIdleConnsNum, maxOpenConnsNum, connMaxLifeMinutes int

func init() {
	maxIdleConnsNum = config.GetInt("SQL_MAX_IDLE_CONNS")
	maxOpenConnsNum = config.GetInt("SQL_MAX_OPEN_CONNS")
	connMaxLifeMinutes = config.GetInt("SQL_CONN_MAX_LIFE_MINUTES")
}

// initialize initializes the PostgreSQL database handle.
func (db *postgresDB) initialize(ctx context.Context, cfg dbConfig) {
	// Assemble PostgreSQL database source and setup database handle.
	dbSource := fmt.Sprintf(`host=%s port=%s user=%s password=%s dbname=%s`, cfg.Address, cfg.Port, cfg.Username, cfg.Password,
		cfg.DBName)

	// Connect to the PostgreSQL database.
	var err error
	db.DB, err = gorm.Open(postgres.Open(dbSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	// Get generic database object sql.DB to set optional params
	sqlDB, err := db.DB.DB()
	if err != nil {
		panic("Get generic database object sql.DB fail")
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(maxIdleConnsNum)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(maxOpenConnsNum)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifeMinutes) * time.Minute)

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
