# MultiCatálogo Backend - API Go/Fiber con PostgreSQL, Redis y JWT

Backend para el sistema de catálogo multinivel construido con Go, Fiber, PostgreSQL, Redis y autenticación JWT.

## Requisitos Previos

- Docker y Docker Compose
- Go 1.26+
- Windows/Linux/macOS

## Configuración Inicial

### 1. Iniciar PostgreSQL y Redis con Docker

```bash
cd multicatalogo-backend
docker-compose up -d
```

En una base nueva, Docker ejecuta en orden `init.sql`, `003_seguridad.sql` y `004_backoffice.sql`. Esto crea el esquema, activa bcrypt y prepara las tablas de back-office.

Esto iniciará:
- PostgreSQL 16 en el puerto 5434
- Redis en el puerto 6379
- Ejecutará automáticamente el script `database/init.sql` para crear las tablas y poblar los datos iniciales

### 2. Verificar que los contenedores estén corriendo

```bash
docker-compose ps
```

Deberías ver los contenedores `multicatalogo-postgres` y `multicatalogo-redis` con estado `Up`.

### 3. Compilar y ejecutar el servidor

```bash
go run .
```

O compilar primero:

```bash
go build -o multicatalogo-backend.exe .
./multicatalogo-backend.exe
```

El servidor iniciará en `http://localhost:3000`

## Endpoints Disponibles

### Autenticación
- `POST /api/login` - Autenticación de usuarios con JWT
  - Cuerpo: `{"email": "string", "password": "string"}`
  - Respuesta exitosa: `{"token": "string", "email": "string", "rol": "admin|cliente"}`
  - **Seguridad**: Passwords hasheados con bcrypt, JWT con expiración de 2 horas
- `POST /api/register` - Crear cuenta cliente; cuerpo `{nombre, email, password, ref?}`

### Productos (Públicos)
- `GET /api/productos?q=&categoria=&page=&limit=` - Catálogo paginado y búsqueda/filtro server-side (con caché Redis para la lista sin filtros)
- `GET /api/productos/:id` - Obtener un producto específico por ID

### Back-office (protegido con JWT de administrador)
- `GET /api/admin/productos?q=&page=&limit=`; `POST /api/admin/productos`; `PUT /api/admin/productos/:id`; `DELETE /api/admin/productos/:id`
- `GET /api/admin/usuarios?page=&limit=`; `PATCH /api/admin/usuarios/:id` para rol y estado de la cuenta
- `GET /api/pedidos?page=&limit=&estado=`; `PATCH /api/pedidos/:id/estado`; `POST /api/pedidos/:id/cancelar`
- `GET /api/comisiones/pendientes`; `POST /api/liquidaciones`; `GET /api/liquidaciones`
- `GET /api/reportes/resumen`; `GET /api/reportes/top-productos`; `GET /api/reportes/top-referidos`; `GET /api/reportes/comisiones-pendientes`

### Cuenta de cliente (protegido con JWT)
- `GET /api/pedidos/mios?page=&limit=`; `GET /api/pedidos/:id` (solo propietario o administrador)
- `GET /api/red/mia`; `GET /api/comisiones/mias`; `GET /api/liquidaciones/mias`

### Red Multinivel (Protegido)
- `GET /api/red` - Obtener la jerarquía de referidos del usuario (con caché Redis)
  - **Requiere**: Token JWT válido en header `Authorization: Bearer <token>`
  - El email se obtiene del token, no de query params

### Checkout (Protegido)
- `POST /api/checkout` - Procesar un pedido con transacciones ACID
  - **Requiere**: Token JWT válido en header `Authorization: Bearer <token>`
  - Cuerpo: `{"direccion": "string", "ciudad": "string", "provincia": "string", "codigo_postal": "string", "metodo_pago": "string", "items": [...]}`
  - Respuesta: `{"pedido_id": int, "total": float64, "estado": "string", "mensaje": "string"}`
  - **Seguridad**: Identidad del usuario obtenida del token JWT (no del body)
  - Los precios y el stock se vuelven a validar en PostgreSQL; pedido, stock y ledger de comisiones se guardan en una transacción
  - El caché de productos y de la red se invalida al actualizar datos

### Sistema
- `GET /api/health` - Verificar estado del servidor y dependencias
  - Respuesta: `{"status": "ok|degraded", "message": "...", "postgres": "up|down", "redis": "up|down"}`
  - Retorna 503 si PostgreSQL o Redis no están disponibles

## Estructura del Proyecto

```
multicatalogo-backend/
├── auth/
│   └── jwt.go                # Generación y validación de JWT
├── cache/
│   └── redisCache.go         # Implementación de caché Redis
├── config/
│   ├── db.go                 # Conexión a PostgreSQL
│   └── redis.go              # Conexión a Redis
├── controllers/
│   ├── authControllerWithDI.go      # Lógica de autenticación con DI
│   ├── prodControllerWithDI.go      # Lógica de productos con DI
│   ├── redControllerWithDI.go       # Lógica de red multinivel con DI
│   ├── checkoutController.go        # Lógica de checkout con transacciones ACID
│   └── healthController.go          # Health check con ping a dependencias
├── database/
│   ├── init.sql               # Script de inicialización de BD
│   ├── 003_seguridad.sql      # Migración de seguridad (bcrypt + índices)
│   └── 004_backoffice.sql     # Esquema de referidos, pedidos, comisiones y liquidaciones
├── interfaces/
│   └── interfaces.go          # Interfaces para inyección de dependencias
├── middleware/
│   └── auth.go               # Middleware RequireAuth y RequireAdmin
├── models/
│   ├── models.go             # Estructuras de datos generales
│   └── checkout.go           # Estructuras específicas de checkout
├── repository/
│   ├── repository.go          # Funciones legacy (compatibilidad)
│   └── postgresRepository.go # Implementación de DBRepository con DI
├── routes/
│   └── routes.go             # Definición de rutas con inyección de dependencias
├── main.go                   # Punto de entrada con inyección de dependencias
├── docker-compose.yml         # Configuración Docker (PostgreSQL + Redis)
├── .env / .env.example       # Configuración BD + Redis + JWT
└── README.md                 # Documentación
```

## Arquitectura

### Inyección de Dependencias
El proyecto utiliza inyección de dependencias para desacoplar la arquitectura:
- **Interfaces**: `DBRepository` y `CacheRepository` definen contratos
- **Implementaciones**: `PostgresRepository` y `RedisCache` implementan las interfaces
- **Controladores**: Reciben las dependencias como parámetros en sus constructores
- **main.go**: Instancia e inyecta todas las dependencias al iniciar

### Seguridad
- **bcrypt**: Passwords hasheados con pgcrypto (PostgreSQL)
- **JWT**: Tokens reales con expiración de 2 horas (configurable via `JWT_EXPIRATION_HOURS`)
- **Middleware**: `RequireAuth` verifica tokens JWT, `RequireAdmin` verifica rol admin
- **Endpoints protegidos**: `/api/red` y `/api/checkout` requieren autenticación
- **Identidad**: El checkout usa el email del token JWT, no del body del request

### Transacciones ACID
El endpoint `/api/checkout` implementa transacciones ACID en PostgreSQL:
- `BEGIN`: Inicia transacción
- Inserta pedido en tabla `pedidos`
- Inserta items en tabla `pedido_items`
- Resta stock de productos
- Si hay error: `ROLLBACK`
- Si todo OK: `COMMIT`

### Concurrencia
El cálculo de comisiones multinivel se ejecuta de forma asíncrona:
- No bloquea la respuesta al cliente
- `go calcularComisiones()` se ejecuta en background
- Actualiza ventas de referidos ascendentes (10%, 5%, 2%)

### Caché Redis
- **Productos**: Clave `productos:all` con TTL de 5 minutos
- **Red Multinivel**: Clave `red:tree:{email}` con TTL de 10 minutos
- **Invalidación**: La caché de la red se elimina automáticamente después de cada checkout
- **Degradación**: Si Redis no responde, la API consulta PostgreSQL directamente

### Health Check
El endpoint `/api/health` verifica:
- Estado del servidor
- Conexión a PostgreSQL (ping)
- Conexión a Redis (ping)
- Retorna `status: "degraded"` y código 503 si alguna dependencia falla

## Datos Iniciales

### Usuarios
- **Admin**: `admin@upse.edu.ec` / `123456`
- **Cliente**: `cliente@upse.edu.ec` / `123456`
- **Nota**: Los passwords están hasheados con bcrypt después de ejecutar la migración `003_seguridad.sql`

### Productos
- 8 productos de cosméticos con descripciones, categorías, galerías y stock

### Red Multinivel
- Estructura jerárquica con 7 referidos distribuidos en 3 niveles
- Sistema de comisiones: 10% (nivel 1), 5% (nivel 2), 2% (nivel 3)

## Variables de Entorno

Edita el archivo `.env` con tus configuraciones locales:

```env
# PostgreSQL
DB_HOST=localhost
DB_PORT=5434
DB_USER=multicatalogo
DB_PASSWORD=multicatalogo123
DB_NAME=multicatalogo

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=multicatalogo-secret-key-2026-very-secure
JWT_EXPIRATION_HOURS=2

# Entorno
APP_ENV=development
```

## Detener el Servidor

Para detener el servidor Go, presiona `Ctrl+C` en la terminal.

Para detener PostgreSQL y Redis:

```bash
docker-compose down
```

## Desarrollo

### Agregar nuevas dependencias

```bash
go get <paquete>
go mod tidy
```

### Aplicar migraciones a una base ya existente

```bash
psql -h localhost -p 5434 -U multicatalogo -d multicatalogo -f database/003_seguridad.sql
psql -h localhost -p 5434 -U multicatalogo -d multicatalogo -f database/004_backoffice.sql
```

## Troubleshooting

### Error de conexión a PostgreSQL
- Verifica que Docker esté corriendo: `docker ps`
- Verifica que el contenedor de PostgreSQL esté iniciado: `docker-compose ps`
- Revisa las credenciales en el archivo `.env`
- Confirma que el puerto 5434 no esté en uso

### Error de conexión a Redis
- Verifica que el contenedor de Redis esté iniciado: `docker-compose ps`
- Revisa las credenciales de Redis en `.env`
- Confirma que el puerto 6379 no esté en uso

### Errores de compilación
- Ejecuta `go mod tidy` para limpiar dependencias
- Asegúrate de tener Go 1.26+ instalado: `go version`

### Errores de autenticación
- Asegúrate de ejecutar la migración `003_seguridad.sql` para hashear passwords
- Verifica que `JWT_SECRET` esté configurado en `.env`
- Los endpoints protegidos requieren el header `Authorization: Bearer <token>`

## Separación de Responsabilidades

El proyecto sigue el patrón de separación de responsabilidades:

- **routes.go**: Define los endpoints HTTP e inyecta controladores
- **controllers**: Manejan la lógica HTTP (req/res) con dependencias inyectadas
- **repository**: Acceso a datos y persistencia (implementa interfaces)
- **interfaces**: Definen contratos para inyección de dependencias
- **middleware**: Autenticación y autorización
- **auth**: Generación y validación de JWT
- **cache**: Implementación de caché Redis
- **models**: Definen las estructuras de datos
- **config**: Configuración y conexiones (DB y Redis)
