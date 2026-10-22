// Paquete repository maneja el acceso a datos y la lógica de persistencia
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
	"time"
)

// Usuario representa un usuario en el sistema
type Usuario struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Rol      string `json:"rol"`
}

// Referido representa un referido en la red multinivel
type Referido struct {
	ID       int        `json:"id"`
	Nombre   string     `json:"nombre"`
	Nivel    int        `json:"nivel"`
	Ventas   float64    `json:"ventas"`
	ParentID *int       `json:"parent_id,omitempty"`
	Hijos    []Referido `json:"hijos,omitempty"`
}

// BuscarUsuarioPorCredenciales busca un usuario por email y password
func BuscarUsuarioPorCredenciales(email, password string) (*Usuario, error) {
	query := `SELECT id, email, password, rol FROM usuarios WHERE email = $1 AND password = $2`

	var usuario Usuario
	err := config.DB.QueryRow(query, email, password).Scan(
		&usuario.ID,
		&usuario.Email,
		&usuario.Password,
		&usuario.Rol,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Usuario no encontrado
		}
		return nil, err
	}

	return &usuario, nil
}

// ObtenerProductos retorna todos los productos del catálogo con caché en Redis (con búsqueda opcional)
func ObtenerProductos(busqueda string) ([]models.Producto, error) {
	ctx := context.Background()
	cacheKey := "productos:all"

	// Si hay búsqueda, no usar caché y hacer búsqueda en BD
	if busqueda != "" {
		query := `
			SELECT id, nombre, descripcion, precio, categoria, img, galeria
			FROM productos
			WHERE nombre ILIKE $1 OR descripcion ILIKE $1 OR categoria ILIKE $1
			ORDER BY id
		`

		searchPattern := "%" + busqueda + "%"
		rows, err := config.DB.Query(query, searchPattern)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var productos []models.Producto
		for rows.Next() {
			var p models.Producto
			var galeria []string

			err := rows.Scan(
				&p.ID,
				&p.Nombre,
				&p.Descripcion,
				&p.Precio,
				&p.Categoria,
				&p.Img,
				&galeria,
			)
			if err != nil {
				return nil, err
			}

			productos = append(productos, p)
		}

		return productos, nil
	}

	// Paso 1: Intentar leer desde Redis (Cache Hit)
	if config.RedisClient != nil {
		cachedData, err := config.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var productos []models.Producto
			if err := json.Unmarshal([]byte(cachedData), &productos); err == nil {
				log.Println("Cache HIT: Productos obtenidos desde Redis")
				return productos, nil
			}
		}
	}

	// Paso 2: Cache Miss - Consultar PostgreSQL
	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria FROM productos ORDER BY id`

	rows, err := config.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var productos []models.Producto
	for rows.Next() {
		var p models.Producto
		var galeria []string

		err := rows.Scan(
			&p.ID,
			&p.Nombre,
			&p.Descripcion,
			&p.Precio,
			&p.Categoria,
			&p.Img,
			&galeria,
		)
		if err != nil {
			return nil, err
		}

		productos = append(productos, p)
	}

	// Paso 3: Guardar en Redis con TTL de 5 minutos
	if config.RedisClient != nil {
		jsonData, err := json.Marshal(productos)
		if err == nil {
			err = config.RedisClient.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err()
			if err != nil {
				log.Printf("Advertencia: Error al guardar productos en caché: %v", err)
			} else {
				log.Println("Productos guardados en caché (TTL: 5 minutos)")
			}
		}
	}

	return productos, nil
}

// ObtenerProductoPorID retorna un producto específico por su ID
func ObtenerProductoPorID(id int) (*models.Producto, error) {
	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria FROM productos WHERE id = $1`

	var p models.Producto
	var galeria []string

	err := config.DB.QueryRow(query, id).Scan(
		&p.ID,
		&p.Nombre,
		&p.Descripcion,
		&p.Precio,
		&p.Categoria,
		&p.Img,
		&galeria,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Producto no encontrado
		}
		return nil, err
	}

	return &p, nil
}

// ObtenerRed construye la jerarquía multinivel de referidos con caché en Redis
func ObtenerRed(email string) (*Referido, error) {
	ctx := context.Background()
	cacheKey := "red:tree:" + email

	// Paso 1: Intentar leer desde Redis (Cache Hit)
	if config.RedisClient != nil {
		cachedData, err := config.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var red Referido
			if err := json.Unmarshal([]byte(cachedData), &red); err == nil {
				log.Println("Cache HIT: Red obtenida desde Redis")
				return &red, nil
			}
		}
	}

	// Paso 2: Cache Miss - Consultar PostgreSQL y construir el árbol
	query := `SELECT id, nombre, nivel, ventas, parent_id FROM referidos ORDER BY id`

	rows, err := config.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Mapa para almacenar todos los referidos por ID
	referidosMap := make(map[int]*Referido)
	var referidosList []*Referido

	for rows.Next() {
		var r Referido
		var parentID sql.NullInt64

		err := rows.Scan(
			&r.ID,
			&r.Nombre,
			&r.Nivel,
			&r.Ventas,
			&parentID,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			r.ParentID = &pid
		}

		referidosMap[r.ID] = &r
		referidosList = append(referidosList, &r)
	}

	// Construimos el árbol jerárquico
	// El usuario raíz (nivel 0) se crea virtualmente
	raiz := &Referido{
		ID:     0,
		Nombre: "Tú",
		Nivel:  0,
		Ventas: 2400, // Ventas propias del usuario
	}

	// Reconstruimos la jerarquía asignando hijos a sus padres
	for _, ref := range referidosList {
		if ref.ParentID != nil {
			if padre, exists := referidosMap[*ref.ParentID]; exists {
				padre.Hijos = append(padre.Hijos, *ref)
			}
		} else {
			// Si no tiene parent_id, es un referido directo del usuario raíz
			raiz.Hijos = append(raiz.Hijos, *ref)
		}
	}

	// Paso 3: Guardar en Redis con TTL de 10 minutos
	if config.RedisClient != nil {
		jsonData, err := json.Marshal(raiz)
		if err == nil {
			err = config.RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute).Err()
			if err != nil {
				log.Printf("Advertencia: Error al guardar red en caché: %v", err)
			} else {
				log.Println("Red guardada en caché (TTL: 10 minutos)")
			}
		}
	}

	return raiz, nil
}

// ObtenerUsuarioPorEmail busca un usuario por email (para validaciones)
func ObtenerUsuarioPorEmail(email string) (*Usuario, error) {
	query := `SELECT id, email, password, rol FROM usuarios WHERE email = $1`

	var usuario Usuario
	err := config.DB.QueryRow(query, email).Scan(
		&usuario.ID,
		&usuario.Email,
		&usuario.Password,
		&usuario.Rol,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Usuario no encontrado
		}
		return nil, err
	}

	return &usuario, nil
}

// InicializarRepository inicializa el repository (verifica conexión)
func InicializarRepository() error {
	if config.DB == nil {
		log.Println("Advertencia: Conexión a base de datos no inicializada")
		return nil
	}

	// Verificar conexión
	if err := config.DB.Ping(); err != nil {
		return err
	}

	log.Println("Repository inicializado correctamente")
	return nil
}
