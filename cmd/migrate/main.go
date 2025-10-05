package main

import (
	"fmt"
	"sample/internal/app/ds"
	"sample/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		fmt.Printf("Error details: %v\n", err)
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.Patient{},
		&ds.InsulinCalculation{},
		&ds.InsulinCalculationPatients{},
		&ds.User{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}

