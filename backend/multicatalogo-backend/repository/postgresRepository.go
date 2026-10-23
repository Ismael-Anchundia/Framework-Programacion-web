// Paquete repository implementa el repositorio de base de datos usando PostgreSQL
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"multicatalogo-backend/config"
	"multicatalogo-backend/domain"
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/models"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	query := `SELECT id, email, password, rol, activo, referido_id FROM usuarios WHERE email = $1 AND verify_password($2, password) = true`

	var usuario interfaces.UsuarioDB
	var referidoID sql.NullInt64
	err := r.db.QueryRow(query, email, password).Scan(
		&usuario.ID,
		&usuario.Email,
		&usuario.Password,
		&usuario.Rol,
		&usuario.Activo,
		&referidoID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Usuario no encontrado o password incorrecto
		}
		return nil, err
	}

	// Verificar que el usuario esté activo
	if !usuario.Activo {
		return nil, fmt.Errorf("cuenta desactivada")
	}

	if referidoID.Valid {
		rid := int(referidoID.Int64)
		usuario.ReferidoID = &rid
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

func (r *PostgresRepository) ObtenerUsuariosPaginados(page, limit int) ([]interfaces.UsuarioDB, int, error) {
	page, limit = validarPagina(page, limit)
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM usuarios").Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`SELECT id, email, rol, COALESCE(nombre, ''), COALESCE(activo, true), referido_id, creado_en FROM usuarios ORDER BY id LIMIT $1 OFFSET $2`, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	usuarios := make([]interfaces.UsuarioDB, 0)
	for rows.Next() {
		var u interfaces.UsuarioDB
		var referidoID sql.NullInt64
		var creadoEn sql.NullTime
		if err := rows.Scan(&u.ID, &u.Email, &u.Rol, &u.Nombre, &u.Activo, &referidoID, &creadoEn); err != nil {
			return nil, 0, err
		}
		if referidoID.Valid {
			id := int(referidoID.Int64)
			u.ReferidoID = &id
		}
		if creadoEn.Valid {
			u.CreadoEn = creadoEn.Time.Format("2006-01-02 15:04:05")
		}
		usuarios = append(usuarios, u)
	}
	return usuarios, total, rows.Err()
}

// ObtenerProductos retorna todos los productos del catálogo con caché en Redis (con búsqueda opcional)
func (r *PostgresRepository) ObtenerProductos(busqueda string) ([]models.Producto, error) {
	ctx := context.Background()
	cacheKey := "productos:all"

	// Si hay búsqueda, no usar caché y hacer búsqueda en BD
	if busqueda != "" {
		query := `
			SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock, activo
			FROM productos
			WHERE activo = true AND (nombre ILIKE $1 OR descripcion ILIKE $1 OR categoria ILIKE $1)
			ORDER BY id
		`

		searchPattern := "%" + busqueda + "%"
		rows, err := r.db.Query(query, searchPattern)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var productos []models.Producto
		for rows.Next() {
			var p models.Producto
			var galeria []string
			var activo bool

			err := rows.Scan(
				&p.ID,
				&p.Nombre,
				&p.Descripcion,
				&p.Precio,
				&p.Categoria,
				&p.Img,
				&galeria,
				&p.Stock,
				&activo,
			)
			if err != nil {
				return nil, err
			}
			p.Galeria = galeria
			p.Activo = activo

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
	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock, activo FROM productos WHERE activo = true ORDER BY id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var productos []models.Producto
	for rows.Next() {
		var p models.Producto
		var galeria []string
		var activo bool

		err := rows.Scan(
			&p.ID,
			&p.Nombre,
			&p.Descripcion,
			&p.Precio,
			&p.Categoria,
			&p.Img,
			&galeria,
			&p.Stock,
			&activo,
		)
		if err != nil {
			return nil, err
		}
		p.Galeria = galeria
		p.Activo = activo

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

func (r *PostgresRepository) ObtenerProductosPaginados(busqueda, categoria string, page, limit int) ([]models.Producto, int, error) {
	page, limit = validarPagina(page, limit)
	busqueda = strings.TrimSpace(busqueda)
	categoria = strings.TrimSpace(categoria)
	if busqueda == "" && categoria == "" {
		productos, err := r.ObtenerProductos("")
		if err != nil {
			return nil, 0, err
		}
		data, total := paginaSlice(productos, page, limit)
		return data, total, nil
	}
	pattern := "%" + busqueda + "%"
	where := "activo = true AND ($1 = '' OR nombre ILIKE $2 OR descripcion ILIKE $2 OR categoria ILIKE $2) AND ($3 = '' OR LOWER(BTRIM(categoria)) = LOWER(BTRIM($3)))"
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM productos WHERE "+where, busqueda, pattern, categoria).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock, activo FROM productos WHERE `+where+` ORDER BY id LIMIT $4 OFFSET $5`, busqueda, pattern, categoria, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	productos, err := scanProductos(rows)
	return productos, total, err
}

// ObtenerProductoPorID retorna un producto específico por su ID
func (r *PostgresRepository) ObtenerProductoPorID(id int) (*models.Producto, error) {
	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock, activo FROM productos WHERE id = $1 AND activo = true`

	var p models.Producto
	var galeria []string
	var activo bool

	err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Nombre,
		&p.Descripcion,
		&p.Precio,
		&p.Categoria,
		&p.Img,
		&galeria,
		&p.Stock,
		&activo,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Producto no encontrado
		}
		return nil, err
	}
	p.Galeria = galeria
	p.Activo = activo

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
	defer tx.Rollback()

	// Reservar el stock y calcular precios desde la base de datos. Nunca se
	// confía en los importes que envía el navegador.
	required := make(map[int]int)
	for i := range items {
		if items[i].Cantidad < 1 {
			return fmt.Errorf("la cantidad debe ser mayor que cero")
		}
		required[items[i].ProductoID] += items[i].Cantidad
	}
	productIDs := make([]int, 0, len(required))
	for id := range required {
		productIDs = append(productIDs, id)
	}
	sort.Ints(productIDs)
	prices := make(map[int]float64, len(required))
	for _, productID := range productIDs {
		var precio float64
		err = tx.QueryRow(`
			SELECT precio FROM productos
			WHERE id = $1 AND activo = true AND stock >= $2
			FOR UPDATE
		`, productID, required[productID]).Scan(&precio)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("producto %d no disponible o sin stock suficiente", productID)
			}
			return err
		}
		prices[productID] = precio
	}
	var total float64
	for i := range items {
		items[i].Precio = prices[items[i].ProductoID]
		items[i].Subtotal = items[i].Precio * float64(items[i].Cantidad)
		total += items[i].Subtotal
	}
	pedido.Total = total

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

	}
	for _, productID := range productIDs {
		if _, err = tx.Exec("UPDATE productos SET stock = stock - $1 WHERE id = $2", required[productID], productID); err != nil {
			return err
		}
	}

	if err = crearComisionesTx(tx, pedido.ID, pedido.UsuarioID, pedido.Total); err != nil {
		return err
	}

	// Commit de la transacción
	err = tx.Commit()
	if err != nil {
		return err
	}
	if config.RedisClient != nil {
		_ = config.RedisClient.Del(context.Background(), "productos:all").Err()
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

// CrearUsuario crea un nuevo usuario con password hasheado
func (r *PostgresRepository) CrearUsuario(usuario *interfaces.UsuarioDB, referidoID *int) error {
	// Verificar si el email ya existe
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE email = $1)", usuario.Email).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email ya existe")
	}

	// Hashear password con bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(usuario.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Iniciar transacción para crear usuario y su nodo en referidos
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Insertar usuario
	queryUsuario := `
		INSERT INTO usuarios (email, password, rol, nombre, activo, referido_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err = tx.QueryRow(
		queryUsuario,
		usuario.Email,
		string(hashedPassword),
		usuario.Rol,
		usuario.Nombre,
		usuario.Activo,
		referidoID,
	).Scan(&usuario.ID)

	if err != nil {
		return err
	}

	// Crear nodo en referidos
	nivel := 1
	if referidoID != nil {
		// Obtener nivel del referente
		var nivelReferente int
		err = tx.QueryRow("SELECT nivel FROM referidos WHERE id = $1", *referidoID).Scan(&nivelReferente)
		if err == nil {
			nivel = nivelReferente + 1
		}
	}

	queryReferido := `
		INSERT INTO referidos (nombre, nivel, ventas, parent_id, codigo)
		VALUES ($1, $2, 0, $3, $4)
		RETURNING id
	`

	var referidoIDNuevo int
	codigo := fmt.Sprintf("REF%d", usuario.ID) // Generar código único
	err = tx.QueryRow(queryReferido, usuario.Nombre, nivel, referidoID, codigo).Scan(&referidoIDNuevo)
	if err != nil {
		return err
	}

	// Actualizar usuario con su referido_id
	_, err = tx.Exec("UPDATE usuarios SET referido_id = $1 WHERE id = $2", referidoIDNuevo, usuario.ID)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// ObtenerUsuarios retorna todos los usuarios (admin)
func (r *PostgresRepository) ObtenerUsuarios() ([]interfaces.UsuarioDB, error) {
	query := `SELECT id, email, rol, COALESCE(nombre, ''), COALESCE(activo, true), referido_id, creado_en FROM usuarios ORDER BY id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []interfaces.UsuarioDB
	for rows.Next() {
		var u interfaces.UsuarioDB
		var referidoID sql.NullInt64
		var creadoEn sql.NullTime

		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.Rol,
			&u.Nombre,
			&u.Activo,
			&referidoID,
			&creadoEn,
		)
		if err != nil {
			return nil, err
		}

		if referidoID.Valid {
			rid := int(referidoID.Int64)
			u.ReferidoID = &rid
		}

		if creadoEn.Valid {
			u.CreadoEn = creadoEn.Time.Format("2006-01-02 15:04:05")
		}

		usuarios = append(usuarios, u)
	}

	return usuarios, nil
}

// ActualizarUsuarioEstado activa/desactiva un usuario
func (r *PostgresRepository) ActualizarUsuarioEstado(usuarioID int, activo bool) error {
	_, err := r.db.Exec("UPDATE usuarios SET activo = $1 WHERE id = $2", activo, usuarioID)
	return err
}

func (r *PostgresRepository) ActualizarUsuario(usuarioID int, activo *bool, rol *string) error {
	if rol != nil && *rol != "admin" && *rol != "cliente" {
		return fmt.Errorf("rol inválido")
	}
	result, err := r.db.Exec(`
		UPDATE usuarios SET
			activo = COALESCE($1, activo),
			rol = COALESCE($2, rol)
		WHERE id = $3
	`, activo, rol, usuarioID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ObtenerRedPorReferidoID retorna la red desde un referido específico
func (r *PostgresRepository) ObtenerRedPorReferidoID(referidoID int) (*interfaces.ReferidoDB, error) {
	query := `SELECT id, nombre, nivel, ventas, parent_id, codigo FROM referidos WHERE id = $1`

	var r1 interfaces.ReferidoDB
	var parentID sql.NullInt64
	var codigo sql.NullString

	err := r.db.QueryRow(query, referidoID).Scan(
		&r1.ID,
		&r1.Nombre,
		&r1.Nivel,
		&r1.Ventas,
		&parentID,
		&codigo,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if parentID.Valid {
		pid := int(parentID.Int64)
		r1.ParentID = &pid
	}

	if codigo.Valid {
		r1.Codigo = codigo.String
	}

	// Obtener hijos recursivamente
	hijos, err := r.obtenerHijosReferidos(referidoID)
	if err != nil {
		return nil, err
	}
	r1.Hijos = hijos

	return &r1, nil
}

// obtenerHijosReferidos obtiene los hijos de un referido recursivamente
func (r *PostgresRepository) obtenerHijosReferidos(parentID int) ([]interfaces.ReferidoDB, error) {
	query := `SELECT id, nombre, nivel, ventas, parent_id, codigo FROM referidos WHERE parent_id = $1 ORDER BY id`

	rows, err := r.db.Query(query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hijos []interfaces.ReferidoDB
	for rows.Next() {
		var r1 interfaces.ReferidoDB
		var parentID sql.NullInt64
		var codigo sql.NullString

		err := rows.Scan(
			&r1.ID,
			&r1.Nombre,
			&r1.Nivel,
			&r1.Ventas,
			&parentID,
			&codigo,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			r1.ParentID = &pid
		}

		if codigo.Valid {
			r1.Codigo = codigo.String
		}

		// Recursión para obtener hijos
		nietos, err := r.obtenerHijosReferidos(r1.ID)
		if err != nil {
			return nil, err
		}
		r1.Hijos = nietos

		hijos = append(hijos, r1)
	}

	return hijos, nil
}

// ObtenerReferidoPorCodigo busca un referido por su código
func (r *PostgresRepository) ObtenerReferidoPorCodigo(codigo string) (*interfaces.ReferidoDB, error) {
	query := `SELECT id, nombre, nivel, ventas, parent_id, codigo FROM referidos WHERE codigo = $1`

	var r1 interfaces.ReferidoDB
	var parentID sql.NullInt64

	err := r.db.QueryRow(query, codigo).Scan(
		&r1.ID,
		&r1.Nombre,
		&r1.Nivel,
		&r1.Ventas,
		&parentID,
		&r1.Codigo,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if parentID.Valid {
		pid := int(parentID.Int64)
		r1.ParentID = &pid
	}

	return &r1, nil
}

// Métodos stub para implementación futura (FASE 2-7)

// CrearProducto crea un nuevo producto
func (r *PostgresRepository) CrearProducto(producto *models.Producto) error {
	query := `
		INSERT INTO productos (nombre, descripcion, precio, categoria, img, galeria, stock, activo)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var galeria []string
	if len(producto.Galeria) > 0 {
		galeria = producto.Galeria
	}

	err := r.db.QueryRow(
		query,
		producto.Nombre,
		producto.Descripcion,
		producto.Precio,
		producto.Categoria,
		producto.Img,
		galeria,
		producto.Stock,
		true, // activo por defecto
	).Scan(&producto.ID)

	if err != nil {
		return err
	}

	// Invalidar caché de productos
	if config.RedisClient != nil {
		ctx := context.Background()
		config.RedisClient.Del(ctx, "productos:all")
		log.Println("Caché de productos invalidada tras crear producto")
	}

	return nil
}

// ActualizarProducto actualiza un producto existente
func (r *PostgresRepository) ActualizarProducto(id int, producto *models.Producto) error {
	query := `
		UPDATE productos
		SET nombre = $1, descripcion = $2, precio = $3, categoria = $4, img = $5, galeria = $6, stock = $7, activo = $8
		WHERE id = $9
	`

	var galeria []string
	if len(producto.Galeria) > 0 {
		galeria = producto.Galeria
	}

	_, err := r.db.Exec(
		query,
		producto.Nombre,
		producto.Descripcion,
		producto.Precio,
		producto.Categoria,
		producto.Img,
		galeria,
		producto.Stock,
		producto.Activo,
		id,
	)

	if err != nil {
		return err
	}

	// Invalidar caché de productos
	if config.RedisClient != nil {
		ctx := context.Background()
		config.RedisClient.Del(ctx, "productos:all")
		log.Println("Caché de productos invalidada tras actualizar producto")
	}

	return nil
}

// DesactivarProducto desactiva un producto (borrado lógico)
func (r *PostgresRepository) DesactivarProducto(id int) error {
	_, err := r.db.Exec("UPDATE productos SET activo = false WHERE id = $1", id)
	if err != nil {
		return err
	}

	// Invalidar caché de productos
	if config.RedisClient != nil {
		ctx := context.Background()
		config.RedisClient.Del(ctx, "productos:all")
		log.Println("Caché de productos invalidada tras desactivar producto")
	}

	return nil
}

// ObtenerProductosAdmin retorna todos los productos incluyendo inactivos
func (r *PostgresRepository) ObtenerProductosAdmin(busqueda string) ([]models.Producto, error) {
	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock, activo FROM productos
		WHERE ($1 = '' OR nombre ILIKE $2 OR descripcion ILIKE $2 OR categoria ILIKE $2) ORDER BY id`
	pattern := "%" + busqueda + "%"
	rows, err := r.db.Query(query, busqueda, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var productos []models.Producto
	for rows.Next() {
		var p models.Producto
		var galeria []string
		var activo bool

		err := rows.Scan(
			&p.ID,
			&p.Nombre,
			&p.Descripcion,
			&p.Precio,
			&p.Categoria,
			&p.Img,
			&galeria,
			&p.Stock,
			&activo,
		)
		if err != nil {
			return nil, err
		}
		p.Galeria = galeria
		p.Activo = activo

		productos = append(productos, p)
	}

	return productos, nil
}

func (r *PostgresRepository) ObtenerProductosAdminPaginados(busqueda string, page, limit int) ([]models.Producto, int, error) {
	page, limit = validarPagina(page, limit)
	busqueda = strings.TrimSpace(busqueda)
	pattern := "%" + busqueda + "%"
	where := "($1 = '' OR nombre ILIKE $2 OR descripcion ILIKE $2 OR categoria ILIKE $2)"
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM productos WHERE "+where, busqueda, pattern).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock, activo FROM productos WHERE `+where+` ORDER BY id LIMIT $3 OFFSET $4`, busqueda, pattern, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	productos, err := scanProductos(rows)
	return productos, total, err
}

func paginaSlice[T any](items []T, page, limit int) ([]T, int) {
	total := len(items)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	start := (page - 1) * limit
	if start >= total {
		return []T{}, total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total
}

func validarPagina(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return page, limit
}

func scanProductos(rows *sql.Rows) ([]models.Producto, error) {
	productos := make([]models.Producto, 0)
	for rows.Next() {
		var p models.Producto
		var galeria []string
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &galeria, &p.Stock, &p.Activo); err != nil {
			return nil, err
		}
		p.Galeria = galeria
		productos = append(productos, p)
	}
	return productos, rows.Err()
}

// ObtenerPedidosMios retorna los pedidos del usuario autenticado (paginado)
func (r *PostgresRepository) ObtenerPedidosMios(usuarioID int, page, limit int) ([]interfaces.PedidoDB, int, error) {
	offset := (page - 1) * limit

	// Contar total
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM pedidos WHERE usuario_id = $1", usuarioID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Obtener pedidos paginados
	query := `
		SELECT id, usuario_id, total, estado, direccion, metodo_pago, ciudad, provincia, codigo_postal, creado_en
		FROM pedidos
		WHERE usuario_id = $1
		ORDER BY creado_en DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, usuarioID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var pedidos []interfaces.PedidoDB
	for rows.Next() {
		var p interfaces.PedidoDB
		var creadoEn sql.NullTime

		err := rows.Scan(
			&p.ID,
			&p.UsuarioID,
			&p.Total,
			&p.Estado,
			&p.Direccion,
			&p.MetodoPago,
			&p.Ciudad,
			&p.Provincia,
			&p.CodigoPostal,
			&creadoEn,
		)
		if err != nil {
			return nil, 0, err
		}

		if creadoEn.Valid {
			p.CreadoEn = creadoEn.Time.Format("2006-01-02 15:04:05")
		}

		pedidos = append(pedidos, p)
	}

	return pedidos, total, nil
}

// ObtenerPedidosTodos retorna todos los pedidos (admin, paginado + filtro estado)
func (r *PostgresRepository) ObtenerPedidosTodos(page, limit int, estado string) ([]interfaces.PedidoDB, int, error) {
	offset := (page - 1) * limit

	// Construir query dinámica
	query := `SELECT id, usuario_id, total, estado, direccion, metodo_pago, ciudad, provincia, codigo_postal, creado_en FROM pedidos`
	var args []interface{}
	argCount := 1

	if estado != "" {
		query += " WHERE estado = $" + fmt.Sprintf("%d", argCount)
		args = append(args, estado)
		argCount++
	}

	query += " ORDER BY creado_en DESC LIMIT $" + fmt.Sprintf("%d", argCount) + " OFFSET $" + fmt.Sprintf("%d", argCount+1)
	args = append(args, limit, offset)

	// Contar total
	countQuery := "SELECT COUNT(*) FROM pedidos"
	if estado != "" {
		countQuery += " WHERE estado = $1"
	}
	var total int
	if estado != "" {
		err := r.db.QueryRow(countQuery, estado).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.db.QueryRow(countQuery).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var pedidos []interfaces.PedidoDB
	for rows.Next() {
		var p interfaces.PedidoDB
		var creadoEn sql.NullTime

		err := rows.Scan(
			&p.ID,
			&p.UsuarioID,
			&p.Total,
			&p.Estado,
			&p.Direccion,
			&p.MetodoPago,
			&p.Ciudad,
			&p.Provincia,
			&p.CodigoPostal,
			&creadoEn,
		)
		if err != nil {
			return nil, 0, err
		}

		if creadoEn.Valid {
			p.CreadoEn = creadoEn.Time.Format("2006-01-02 15:04:05")
		}

		pedidos = append(pedidos, p)
	}

	return pedidos, total, nil
}

// ObtenerPedidoPorID retorna un pedido con sus items
func (r *PostgresRepository) ObtenerPedidoPorID(id int) (*interfaces.PedidoDB, error) {
	query := `
		SELECT id, usuario_id, total, estado, direccion, metodo_pago, ciudad, provincia, codigo_postal, creado_en
		FROM pedidos
		WHERE id = $1
	`

	var p interfaces.PedidoDB
	var creadoEn sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.UsuarioID,
		&p.Total,
		&p.Estado,
		&p.Direccion,
		&p.MetodoPago,
		&p.Ciudad,
		&p.Provincia,
		&p.CodigoPostal,
		&creadoEn,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if creadoEn.Valid {
		p.CreadoEn = creadoEn.Time.Format("2006-01-02 15:04:05")
	}

	itemRows, err := r.db.Query(`
		SELECT pi.id, pi.pedido_id, pi.producto_id, p.nombre, pi.cantidad, pi.precio, pi.subtotal
		FROM pedido_items pi JOIN productos p ON p.id = pi.producto_id
		WHERE pi.pedido_id = $1 ORDER BY pi.id
	`, id)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var item interfaces.PedidoItemDB
		if err := itemRows.Scan(&item.ID, &item.PedidoID, &item.ProductoID, &item.Nombre, &item.Cantidad, &item.Precio, &item.Subtotal); err != nil {
			return nil, err
		}
		p.Items = append(p.Items, item)
	}
	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	return &p, nil
}

// ActualizarEstadoPedido actualiza el estado de un pedido
func (r *PostgresRepository) ActualizarEstadoPedido(id int, estado string) error {
	// Validar transiciones válidas
	estadosValidos := map[string][]string{
		"pendiente":  {"preparando"},
		"preparando": {"enviado"},
		"enviado":    {"entregado"},
		"cancelado":  {},
		"entregado":  {},
	}

	// Obtener estado actual
	var estadoActual string
	err := r.db.QueryRow("SELECT estado FROM pedidos WHERE id = $1", id).Scan(&estadoActual)
	if err != nil {
		return err
	}

	// Verificar que la transición sea válida
	transicionesPermitidas, existe := estadosValidos[estadoActual]
	if !existe {
		return fmt.Errorf("estado actual '%s' no permite transiciones", estadoActual)
	}

	valido := false
	for _, t := range transicionesPermitidas {
		if t == estado {
			valido = true
			break
		}
	}

	if !valido {
		return fmt.Errorf("transición inválida de '%s' a '%s'", estadoActual, estado)
	}

	_, err = r.db.Exec("UPDATE pedidos SET estado = $1 WHERE id = $2", estado, id)
	return err
}

// CancelarPedido cancela un pedido y devuelve stock en transacción
func (r *PostgresRepository) CancelarPedido(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var estado string
	if err = tx.QueryRow("SELECT estado FROM pedidos WHERE id = $1 FOR UPDATE", id).Scan(&estado); err != nil {
		return err
	}
	if estado != "pendiente" && estado != "preparando" {
		return fmt.Errorf("el pedido no puede cancelarse desde el estado %q", estado)
	}

	// Devolver stock por cada item
	queryItems := `
		SELECT producto_id, cantidad
		FROM pedido_items
		WHERE pedido_id = $1
	`

	rows, err := tx.Query(queryItems, id)
	if err != nil {
		return err
	}
	var productos []interfaces.PedidoItemDB
	for rows.Next() {
		var item interfaces.PedidoItemDB
		if err = rows.Scan(&item.ProductoID, &item.Cantidad); err != nil {
			rows.Close()
			return err
		}
		productos = append(productos, item)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for _, item := range productos {
		if _, err = tx.Exec("UPDATE productos SET stock = stock + $1 WHERE id = $2", item.Cantidad, item.ProductoID); err != nil {
			return err
		}
	}
	if _, err = tx.Exec("UPDATE pedidos SET estado = 'cancelado' WHERE id = $1", id); err != nil {
		return err
	}

	// Revertir comisiones (marcar como revertida)
	_, err = tx.Exec("UPDATE comisiones SET estado = 'revertida' WHERE pedido_id = $1", id)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	if config.RedisClient != nil {
		_ = config.RedisClient.Del(context.Background(), "productos:all").Err()
	}
	return nil
}

// ComisionConReferido representa una comisión con su referido
type ComisionConReferido struct {
	ReferidoID int
	Nivel      int
	Porcentaje float64
	Monto      float64
}

// CrearComisiones crea comisiones para un pedido dentro de la transacción
func (r *PostgresRepository) CrearComisiones(pedidoID int, usuarioID int, total float64) error {
	return crearComisionesTx(r.db, pedidoID, usuarioID, total)
}

type commissionDB interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

func crearComisionesTx(db commissionDB, pedidoID, usuarioID int, total float64) error {
	var referidoID sql.NullInt64
	if err := db.QueryRow("SELECT referido_id FROM usuarios WHERE id = $1", usuarioID).Scan(&referidoID); err != nil {
		return err
	}
	if !referidoID.Valid {
		return nil
	}

	comisiones := domain.RepartirComisiones(total)
	currentID := int(referidoID.Int64)
	for i, comision := range comisiones {
		var parentID sql.NullInt64
		if err := db.QueryRow("SELECT parent_id FROM referidos WHERE id = $1", currentID).Scan(&parentID); err != nil {
			return err
		}
		if !parentID.Valid {
			break
		}
		if _, err := db.Exec(`
			INSERT INTO comisiones (pedido_id, referido_id, nivel, porcentaje, monto, estado)
			VALUES ($1, $2, $3, $4, $5, 'pendiente')
		`, pedidoID, int(parentID.Int64), i+1, comision.Porcentaje, comision.Monto); err != nil {
			return err
		}
		if _, err := db.Exec("UPDATE referidos SET ventas = ventas + $1 WHERE id = $2", total, int(parentID.Int64)); err != nil {
			return err
		}
		currentID = int(parentID.Int64)
	}
	return nil
}

// obtenerCadenaReferidos obtiene la cadena de referidos ascendentes hasta 3 niveles
func (r *PostgresRepository) obtenerCadenaReferidos(usuarioID int) []ComisionConReferido {
	// Primero obtener el referido_id del usuario
	var referidoID sql.NullInt64
	err := r.db.QueryRow("SELECT referido_id FROM usuarios WHERE id = $1", usuarioID).Scan(&referidoID)
	if err != nil || !referidoID.Valid {
		return []ComisionConReferido{}
	}

	// Obtener la cadena de referidos (máximo 3 niveles)
	var referidos []ComisionConReferido
	porcentajes := []float64{10.0, 5.0, 2.0}

	currentID := int(referidoID.Int64)
	for i := 0; i < 3; i++ {
		if currentID == 0 {
			break
		}

		var parentID sql.NullInt64
		err := r.db.QueryRow("SELECT parent_id FROM referidos WHERE id = $1", currentID).Scan(&parentID)
		if err != nil || !parentID.Valid {
			break
		}

		referidos = append(referidos, ComisionConReferido{
			ReferidoID: int(parentID.Int64),
			Nivel:      i + 1,
			Porcentaje: porcentajes[i],
			Monto:      0, // Se calculará según el total
		})

		currentID = int(parentID.Int64)
	}

	return referidos
}

// ObtenerComisionesMias retorna las comisiones del usuario
func (r *PostgresRepository) ObtenerComisionesMias(referidoID int) ([]interfaces.ComisionDB, error) {
	query := `
		SELECT id, pedido_id, referido_id, nivel, porcentaje, monto, estado, liquidacion_id, creado_en
		FROM comisiones
		WHERE referido_id = $1
		ORDER BY creado_en DESC
	`

	rows, err := r.db.Query(query, referidoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comisiones []interfaces.ComisionDB
	for rows.Next() {
		var c interfaces.ComisionDB
		var liquidacionID sql.NullInt64
		var creadoEn sql.NullTime

		err := rows.Scan(
			&c.ID,
			&c.PedidoID,
			&c.ReferidoID,
			&c.Nivel,
			&c.Porcentaje,
			&c.Monto,
			&c.Estado,
			&liquidacionID,
			&creadoEn,
		)
		if err != nil {
			return nil, err
		}

		if liquidacionID.Valid {
			lid := int(liquidacionID.Int64)
			c.LiquidacionID = &lid
		}

		if creadoEn.Valid {
			c.CreadoEn = creadoEn.Time.Format("2006-01-02 15:04:05")
		}

		comisiones = append(comisiones, c)
	}

	return comisiones, nil
}

// ObtenerComisionesPendientes retorna comisiones pendientes de pago
func (r *PostgresRepository) ObtenerComisionesPendientes() ([]interfaces.ComisionDB, error) {
	query := `
		SELECT c.id, c.pedido_id, c.referido_id, c.nivel, c.porcentaje, c.monto, c.estado, c.liquidacion_id, c.creado_en, r.nombre
		FROM comisiones c
		JOIN referidos r ON c.referido_id = r.id
		WHERE c.estado = 'pendiente'
		ORDER BY c.creado_en DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comisiones []interfaces.ComisionDB
	for rows.Next() {
		var c interfaces.ComisionDB
		var liquidacionID sql.NullInt64
		var creadoEn sql.NullTime

		err := rows.Scan(
			&c.ID,
			&c.PedidoID,
			&c.ReferidoID,
			&c.Nivel,
			&c.Porcentaje,
			&c.Monto,
			&c.Estado,
			&liquidacionID,
			&creadoEn,
			&c.ReferidoNombre,
		)
		if err != nil {
			return nil, err
		}

		if liquidacionID.Valid {
			lid := int(liquidacionID.Int64)
			c.LiquidacionID = &lid
		}

		if creadoEn.Valid {
			c.CreadoEn = creadoEn.Time.Format("2006-01-02 15:04:05")
		}

		comisiones = append(comisiones, c)
	}

	return comisiones, nil
}

// CrearLiquidacion crea una liquidación para un referido
func (r *PostgresRepository) CrearLiquidacion(referidoID int, total float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Calcular y bloquear las comisiones pendientes; el total del cliente no es fuente de verdad.
	rows, err := tx.Query(`SELECT id, monto FROM comisiones WHERE referido_id = $1 AND estado = 'pendiente' FOR UPDATE`, referidoID)
	if err != nil {
		return err
	}
	var comisionIDs []int
	var totalPendiente float64
	for rows.Next() {
		var id int
		var monto float64
		if err = rows.Scan(&id, &monto); err != nil {
			rows.Close()
			return err
		}
		comisionIDs = append(comisionIDs, id)
		totalPendiente += monto
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	if len(comisionIDs) == 0 || totalPendiente <= 0 {
		return fmt.Errorf("no hay comisiones pendientes para liquidar")
	}

	// Crear liquidación
	query := `
		INSERT INTO liquidaciones (referido_id, total, estado)
		VALUES ($1, $2, 'procesando')
		RETURNING id
	`

	var liquidacionID int
	err = tx.QueryRow(query, referidoID, totalPendiente).Scan(&liquidacionID)
	if err != nil {
		return err
	}

	// Marcar comisiones como pagadas
	_, err = tx.Exec(`
		UPDATE comisiones
		SET estado = 'pagada', liquidacion_id = $1
		WHERE referido_id = $2 AND estado = 'pendiente'
	`, liquidacionID, referidoID)

	if err != nil {
		return err
	}

	// Actualizar estado de liquidación
	_, err = tx.Exec("UPDATE liquidaciones SET estado = 'pagada', pagada_en = NOW() WHERE id = $1", liquidacionID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ObtenerLiquidacionesMias retorna las liquidaciones del usuario
func (r *PostgresRepository) ObtenerLiquidacionesMias(referidoID int) ([]interfaces.LiquidacionDB, error) {
	query := `
		SELECT id, referido_id, total, estado, creado_en, pagada_en
		FROM liquidaciones
		WHERE referido_id = $1
		ORDER BY creado_en DESC
	`

	rows, err := r.db.Query(query, referidoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var liquidaciones []interfaces.LiquidacionDB
	for rows.Next() {
		var l interfaces.LiquidacionDB
		var creadoEn sql.NullTime
		var pagadaEn sql.NullTime

		err := rows.Scan(
			&l.ID,
			&l.ReferidoID,
			&l.Total,
			&l.Estado,
			&creadoEn,
			&pagadaEn,
		)
		if err != nil {
			return nil, err
		}

		if creadoEn.Valid {
			l.CreadoEn = creadoEn.Time.Format("2006-01-02 15:04:05")
		}

		if pagadaEn.Valid {
			l.PagadaEn = pagadaEn.Time.Format("2006-01-02 15:04:05")
		}

		liquidaciones = append(liquidaciones, l)
	}

	return liquidaciones, nil
}

// ObtenerLiquidacionesTodas retorna todas las liquidaciones (admin)
func (r *PostgresRepository) ObtenerLiquidacionesTodas() ([]interfaces.LiquidacionDB, error) {
	query := `
		SELECT l.id, l.referido_id, l.total, l.estado, l.creado_en, l.pagada_en, r.nombre
		FROM liquidaciones l
		JOIN referidos r ON l.referido_id = r.id
		ORDER BY l.creado_en DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var liquidaciones []interfaces.LiquidacionDB
	for rows.Next() {
		var l interfaces.LiquidacionDB
		var creadoEn sql.NullTime
		var pagadaEn sql.NullTime
		var nombre string

		err := rows.Scan(
			&l.ID,
			&l.ReferidoID,
			&l.Total,
			&l.Estado,
			&creadoEn,
			&pagadaEn,
			&nombre,
		)
		if err != nil {
			return nil, err
		}

		if creadoEn.Valid {
			l.CreadoEn = creadoEn.Time.Format("2006-01-02 15:04:05")
		}

		if pagadaEn.Valid {
			l.PagadaEn = pagadaEn.Time.Format("2006-01-02 15:04:05")
		}

		liquidaciones = append(liquidaciones, l)
	}

	return liquidaciones, nil
}

// ObtenerReporteResumen retorna reporte de resumen de ventas
func (r *PostgresRepository) ObtenerReporteResumen(desde, hasta string) (map[string]interface{}, error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(p.total), 0), COALESCE(AVG(p.total), 0),
			COALESCE((SELECT SUM(c.monto) FROM comisiones c JOIN pedidos pc ON pc.id = c.pedido_id
				WHERE pc.estado <> 'cancelado' AND c.estado <> 'revertida'
				AND ($1 = '' OR pc.creado_en >= $1::date)
				AND ($2 = '' OR pc.creado_en < ($2::date + INTERVAL '1 day'))), 0)
		FROM pedidos p
		WHERE p.estado <> 'cancelado'
		AND ($1 = '' OR p.creado_en >= $1::date)
		AND ($2 = '' OR p.creado_en < ($2::date + INTERVAL '1 day'))
	`

	result := make(map[string]interface{})
	var totalPedidos int
	var totalVendido, ticketPromedio, comisionesDevengadas sql.NullFloat64

	err := r.db.QueryRow(query, desde, hasta).Scan(
		&totalPedidos,
		&totalVendido,
		&ticketPromedio,
		&comisionesDevengadas,
	)
	if err != nil {
		return nil, err
	}

	result["total_pedidos"] = totalPedidos
	result["total_vendido"] = totalVendido.Float64
	result["ticket_promedio"] = ticketPromedio.Float64
	result["comisiones_devengadas"] = comisionesDevengadas.Float64

	return result, nil
}

// ObtenerTopProductos retorna los productos más vendidos
func (r *PostgresRepository) ObtenerTopProductos(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT
			p.id,
			p.nombre,
			COALESCE(p.categoria, ''),
			SUM(pi.cantidad) as total_vendido,
			SUM(pi.subtotal) as total_ingresos
		FROM pedido_items pi
		JOIN productos p ON pi.producto_id = p.id
		JOIN pedidos ped ON pi.pedido_id = ped.id
		WHERE ped.estado != 'cancelado'
		GROUP BY p.id, p.nombre, p.categoria
		ORDER BY total_ingresos DESC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var productos []map[string]interface{}
	for rows.Next() {
		var id int
		var nombre, categoria string
		var totalVendido int
		var totalIngresos float64

		err := rows.Scan(&id, &nombre, &categoria, &totalVendido, &totalIngresos)
		if err != nil {
			return nil, err
		}

		productos = append(productos, map[string]interface{}{
			"id":             id,
			"nombre":         nombre,
			"categoria":      categoria,
			"total_vendido":  totalVendido,
			"total_ingresos": totalIngresos,
		})
	}

	return productos, nil
}

// ObtenerTopReferidos retorna los referidos con más ventas
func (r *PostgresRepository) ObtenerTopReferidos(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT
			r.id,
			r.nombre,
			r.nivel,
			COALESCE(r.ventas, 0),
			COUNT(c.id) as total_comisiones
		FROM referidos r
		LEFT JOIN comisiones c ON r.id = c.referido_id
		GROUP BY r.id, r.nombre, r.nivel, r.ventas
		ORDER BY r.ventas DESC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var referidos []map[string]interface{}
	for rows.Next() {
		var id int
		var nombre string
		var nivel int
		var ventas float64
		var totalComisiones int

		err := rows.Scan(&id, &nombre, &nivel, &ventas, &totalComisiones)
		if err != nil {
			return nil, err
		}

		referidos = append(referidos, map[string]interface{}{
			"id":               id,
			"nombre":           nombre,
			"nivel":            nivel,
			"ventas":           ventas,
			"total_comisiones": totalComisiones,
		})
	}

	return referidos, nil
}
