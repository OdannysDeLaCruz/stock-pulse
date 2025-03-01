package main

import (
	"log"
	"os"

	"github.com/OdannysDeLaCruz/stock-tracker/migrations"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error cargando el archivo .env")
	}

	// Conectar a la base de datos
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("No se pudo conectar a la base de datos:", err)
	}

	// Ejecutar migraciones
	if err := migrations.RunMigrations(db); err != nil {
		log.Fatal("Error ejecutando migraciones:", err)
	}

	log.Println("Migraciones completadas exitosamente")
} 