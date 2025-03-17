package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	RedisAddress   string
	APIKey         string
	HostAPI        string
	AllowedOrigins map[string]bool
	Port           string
}

// LoadConfig carga las variables de entorno desde .env en la raíz del proyecto
func LoadConfig() *Config {
	// Obtener la ruta absoluta del directorio donde está `main.go`
	basePath, err := os.Getwd()
	if err != nil {
		log.Println("⚠️ No se pudo obtener el directorio actual:", err)
		basePath = "."
	}

	// Construir la ruta al archivo .env en la raíz del proyecto
	envPath := filepath.Join(basePath, ".env")
	log.Println("Cargando variables de entorno desde", envPath) // Imprimir la ruta al archivo .env(envPath)

	// Intentar cargar el .env
	err = godotenv.Load(envPath)
	if err != nil {
		log.Printf("⚠️ Advertencia: No se pudo cargar el archivo .env desde %s\n", envPath)
	}

	return &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		RedisAddress:   os.Getenv("REDIS_ADDR"),
		APIKey:         os.Getenv("API_KEY"),
		HostAPI:        os.Getenv("HOST_API"),
		AllowedOrigins: getOrigins(os.Getenv("ALLOWED_ORIGINS")),
		Port:           getPort(),
	}
}

// getOrigins devuelve los orígenes permitidos
func getOrigins(origins string) map[string]bool {
	DEFAULT_ALLOWED_ORIGINS := "http://localhost:8080"

	allowedOrigins := make(map[string]bool)

	// Si no hay variables de entorno, usar el valor por defecto
	if origins == "" {
		allowedOrigins[DEFAULT_ALLOWED_ORIGINS] = true
		return allowedOrigins
	}

	// Dividir la cadena de orígenes en un slice y agregarlos al mapa
	for _, origin := range strings.Split(origins, ",") {
		allowedOrigins[strings.TrimSpace(origin)] = true
	}

	return allowedOrigins
}

// getPort devuelve el puerto con un valor por defecto
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "3000"
	}
	return port
}
