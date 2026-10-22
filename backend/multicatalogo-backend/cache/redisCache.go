// Paquete cache implementa el repositorio de caché usando Redis
package cache

import (
	"context"
	"fmt"
	"log"
	"multicatalogo-backend/config"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implementa la interfaz CacheRepository
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache crea una nueva instancia de RedisCache
func NewRedisCache() *RedisCache {
	return &RedisCache{
		client: config.RedisClient,
	}
}

// Get obtiene un valor del caché por su clave
func (r *RedisCache) Get(key string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("cliente Redis no inicializado")
	}

	ctx := context.Background()
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return value, nil
}

// Set guarda un valor en el caché con un tiempo de expiración
func (r *RedisCache) Set(key string, value string, ttlSeconds int) error {
	if r.client == nil {
		return fmt.Errorf("cliente Redis no inicializado")
	}

	ctx := context.Background()
	ttl := time.Duration(ttlSeconds) * time.Second
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		log.Printf("Error al guardar en caché: %v", err)
		return err
	}

	return nil
}

// Delete elimina una clave del caché
func (r *RedisCache) Delete(key string) error {
	if r.client == nil {
		return fmt.Errorf("cliente Redis no inicializado")
	}

	ctx := context.Background()
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Error al eliminar del caché: %v", err)
		return err
	}

	log.Printf("Clave eliminada del caché: %s", key)
	return nil
}
