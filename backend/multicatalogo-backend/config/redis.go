// Paquete config maneja la configuración y conexión a Redis
package config

import (
	"context"
	"fmt"
	"log"

	// Importamos godotenv para cargar variables de entorno
	"github.com/joho/godotenv"
	// Importamos el cliente de Redis para Go
	"github.com/redis/go-redis/v9"
)

// RedisClient es el cliente global de Redis
var RedisClient *redis.Client

// InitRedis establece la conexión a Redis usando variables de entorno
func InitRedis() error {
	// Cargamos variables de entorno desde archivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: No se pudo cargar archivo .env, usando variables de entorno del sistema")
	}

	// Obtenemos las variables de entorno o usamos valores por defecto
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "")

	// Construimos la dirección de Redis
	addr := fmt.Sprintf("%s:%s", host, port)

	// Configuramos el cliente de Redis
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0, // Base de datos por defecto
	})

	// Verificamos que la conexión funcione
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error al verificar la conexión a Redis: %w", err)
	}

	log.Println("Conexión a Redis establecida exitosamente")
	return nil
}

// CloseRedis cierra la conexión a Redis
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
