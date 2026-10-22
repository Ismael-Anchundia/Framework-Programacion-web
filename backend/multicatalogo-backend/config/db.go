// Paquete config maneja la configuración y conexión a la base de datos
package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	// Importamos godotenv para cargar variables de entorno desde archivo .env
	"github.com/joho/godotenv"
	// Importamos el driver pgx para PostgreSQL
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB es la conexión global a la base de datos
var DB *sql.DB

// InitDB establece la conexión a PostgreSQL usando variables de entorno
func InitDB() error {
	// Cargamos variables de entorno desde archivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: No se pudo cargar archivo .env, usando variables de entorno del sistema")
	}

	// En Render se puede usar directamente la URL interna de PostgreSQL.
	// Para desarrollo local seguimos aceptando las variables DB_*.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5434")
		user := getEnv("DB_USER", "multicatalogo")
		password := getEnv("DB_PASSWORD", "multicatalogo123")
		dbname := getEnv("DB_NAME", "multicatalogo")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	}

	// Abrimos la conexión a la base de datos
	var err error
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("error al abrir la conexión a la base de datos: %w", err)
	}

	// Verificamos que la conexión funciona
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error al verificar la conexión a la base de datos: %w", err)
	}

	// Configuramos el pool de conexiones
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	log.Println("Conexión a PostgreSQL establecida exitosamente")
	return nil
}

// CloseDB cierra la conexión a la base de datos
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
