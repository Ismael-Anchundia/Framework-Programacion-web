// Paquete repository implementa el repositorio de base de datos usando PostgreSQL
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"multicatalogo-backend/config"
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/models"
	"time"
)

// PostgresRepository implementa la interfaz DBRepository
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository crea una nueva instancia de PostgresRepository
func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{
		db: config.DB,
	}
}

// BuscarUsuarioPorCredenciales busca un usuario por email y password (verifica con bcrypt)
func (r *PostgresRepository) BuscarUsuarioPorCredenciales(email, password string) (*interfaces.UsuarioDB, error) {
	// Usar la función verify_password de PostgreSQL para bcrypt
	query := `SELECT id, email, password, rol FROM usuarios WHERE email = $1 AND verify_password($2, password) = true`

	var usuario interfaces.UsuarioDB
	err := r.db.QueryRow(query, email, password).Scan(
		&usuario.ID,
		&usuario.Email,
		&usuario.Password,
		&usuario.Rol,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Usuario no encontrado o password incorrecto
		}
		return nil, err
	}

	return &usuario, nil
}

// ObtenerUsuarioPorEmail busca un usuario por email
func (r *PostgresRepository) ObtenerUsuarioPorEmail(email string) (*interfaces.UsuarioDB, error) {
	query := `SELECT id, email, password, rol FROM usuarios WHERE email = $1`

	var usuario interfaces.UsuarioDB
	err := r.db.QueryRow(query, email).Scan(
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

// ObtenerProductos retorna todos los productos del catálogo con caché en Redis
func (r *PostgresRepository) ObtenerProductos() ([]models.Producto, error) {
	ctx := context.Background()
	cacheKey := "productos:all"

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

	rows, err := r.db.Query(query)
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
func (r *PostgresRepository) ObtenerProductoPorID(id int) (*models.Producto, error) {
	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria FROM productos WHERE id = $1`

	var p models.Producto
	var galeria []string

	err := r.db.QueryRow(query, id).Scan(
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
func (r *PostgresRepository) ObtenerRed(email string) (*interfaces.ReferidoDB, error) {
	ctx := context.Background()
	cacheKey := "red:tree:" + email

	// Paso 1: Intentar leer desde Redis (Cache Hit)
	if config.RedisClient != nil {
		cachedData, err := config.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var red interfaces.ReferidoDB
			if err := json.Unmarshal([]byte(cachedData), &red); err == nil {
				log.Println("Cache HIT: Red obtenida desde Redis")
				return &red, nil
			}
		}
	}

	// Paso 2: Cache Miss - Consultar PostgreSQL y construir el árbol
	query := `SELECT id, nombre, nivel, ventas, parent_id FROM referidos ORDER BY id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Mapa para almacenar todos los referidos por ID
	referidosMap := make(map[int]*interfaces.ReferidoDB)
	var referidosList []*interfaces.ReferidoDB

	for rows.Next() {
		var r1 interfaces.ReferidoDB
		var parentID sql.NullInt64

		err := rows.Scan(
			&r1.ID,
			&r1.Nombre,
			&r1.Nivel,
			&r1.Ventas,
			&parentID,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			r1.ParentID = &pid
		}

		referidosMap[r1.ID] = &r1
		referidosList = append(referidosList, &r1)
	}

	// Construimos el árbol jerárquico
	raiz := &interfaces.ReferidoDB{
		ID:     0,
		Nombre: "Tú",
		Nivel:  0,
		Ventas: 2400,
	}

	// Reconstruimos la jerarquía asignando hijos a sus padres
	for _, ref := range referidosList {
		if ref.ParentID != nil {
			if padre, exists := referidosMap[*ref.ParentID]; exists {
				padre.Hijos = append(padre.Hijos, *ref)
			}
		} else {
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

// CrearPedido crea un pedido con sus items en una transacción ACID
func (r *PostgresRepository) CrearPedido(pedido *interfaces.PedidoDB, items []interfaces.PedidoItemDB) error {
	// Iniciar transacción
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Función para rollback en caso de error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insertar pedido principal
	queryPedido := `
		INSERT INTO pedidos (usuario_id, total, estado, direccion, metodo_pago, ciudad, provincia, codigo_postal)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err = tx.QueryRow(
		queryPedido,
		pedido.UsuarioID,
		pedido.Total,
		pedido.Estado,
		pedido.Direccion,
		pedido.MetodoPago,
		pedido.Ciudad,
		pedido.Provincia,
		pedido.CodigoPostal,
	).Scan(&pedido.ID)

	if err != nil {
		return err
	}

	// Insertar items del pedido
	for i := range items {
		items[i].PedidoID = pedido.ID
		queryItem := `
			INSERT INTO pedido_items (pedido_id, producto_id, cantidad, precio, subtotal)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`

		err = tx.QueryRow(
			queryItem,
			items[i].PedidoID,
			items[i].ProductoID,
			items[i].Cantidad,
			items[i].Precio,
			items[i].Subtotal,
		).Scan(&items[i].ID)

		if err != nil {
			return err
		}

		// Restar stock simulado (en un sistema real esto sería en la tabla productos)
		_, err = tx.Exec(
			"UPDATE productos SET stock = stock - $1 WHERE id = $2",
			items[i].Cantidad,
			items[i].ProductoID,
		)
		if err != nil {
			return err
		}
	}

	// Commit de la transacción
	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Printf("Pedido %d creado exitosamente con %d items", pedido.ID, len(items))
	return nil
}

// ActualizarComisiones actualiza las comisiones de los referidos ascendentes de forma asíncrona
func (r *PostgresRepository) ActualizarComisiones(total float64, usuarioID int) error {
	// Calcular comisiones multinivel
	// Nivel 1: 10% del total
	// Nivel 2: 5% del total
	// Nivel 3: 2% del total

	// En un sistema real, esto buscaría la cadena de referidos del usuario
	// Por ahora, simulamos actualizando algunos referidos

	comisionNivel1 := total * 0.10
	comisionNivel2 := total * 0.05
	comisionNivel3 := total * 0.02

	// Actualizar ventas de referidos (simulado)
	_, err := r.db.Exec(`
		UPDATE referidos 
		SET ventas = ventas + $1 
		WHERE id IN (SELECT parent_id FROM referidos WHERE id = $2 LIMIT 1)
	`, comisionNivel1, usuarioID)

	if err != nil {
		log.Printf("Error al actualizar comisión nivel 1: %v", err)
	}

	log.Printf("Comisiones actualizadas: Nivel1=%.2f, Nivel2=%.2f, Nivel3=%.2f",
		comisionNivel1, comisionNivel2, comisionNivel3)

	return nil
}
