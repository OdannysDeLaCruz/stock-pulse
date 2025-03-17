package main

import (
	"log"

	"github.com/OdannysDeLaCruz/stock-pulse/config"
	"github.com/OdannysDeLaCruz/stock-pulse/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()


	// Conectar a la base de datos
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("No se pudo conectar a la base de datos:", err)
	}

	// Ejecutar migraciones
	if err := migrations.RunMigrations(db); err != nil {
		log.Fatal("Error ejecutando migraciones:", err)
	}

	log.Println("Migraciones completadas exitosamente")
}