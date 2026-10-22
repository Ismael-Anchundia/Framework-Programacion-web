-- Script de inicialización de la base de datos multicatalogo
-- Crea las tablas necesarias y popula con los datos iniciales

-- Tabla de referidos (se crea antes de usuarios porque usuarios la referencia)
CREATE TABLE IF NOT EXISTS referidos (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    nivel INTEGER NOT NULL CHECK (nivel >= 0),
    ventas DECIMAL(10, 2) DEFAULT 0,
    parent_id INTEGER REFERENCES referidos(id) ON DELETE SET NULL,
    codigo VARCHAR(20) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de usuarios
CREATE TABLE IF NOT EXISTS usuarios (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    rol VARCHAR(50) NOT NULL CHECK (rol IN ('admin', 'cliente')),
    nombre VARCHAR(255),
    activo BOOLEAN DEFAULT true,
    referido_id INTEGER REFERENCES referidos(id),
    creado_en TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de productos
CREATE TABLE IF NOT EXISTS productos (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    descripcion TEXT,
    precio DECIMAL(10, 2) NOT NULL,
    categoria VARCHAR(100),
    img VARCHAR(500),
    galeria TEXT[], -- Array de URLs de imágenes
    stock INTEGER DEFAULT 100, -- Stock disponible
    activo BOOLEAN DEFAULT true, -- Borrado lógico
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de pedidos
CREATE TABLE IF NOT EXISTS pedidos (
    id SERIAL PRIMARY KEY,
    usuario_id INTEGER REFERENCES usuarios(id) ON DELETE CASCADE,
    total DECIMAL(10, 2) NOT NULL,
    estado VARCHAR(50) DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado')),
    direccion VARCHAR(500),
    metodo_pago VARCHAR(100),
    ciudad VARCHAR(100),
    provincia VARCHAR(100),
    codigo_postal VARCHAR(20),
    creado_en TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de items de pedido
CREATE TABLE IF NOT EXISTS pedido_items (
    id SERIAL PRIMARY KEY,
    pedido_id INTEGER REFERENCES pedidos(id) ON DELETE CASCADE,
    producto_id INTEGER REFERENCES productos(id) ON DELETE RESTRICT,
    cantidad INTEGER NOT NULL CHECK (cantidad > 0),
    precio DECIMAL(10, 2) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de liquidaciones
CREATE TABLE IF NOT EXISTS liquidaciones (
    id SERIAL PRIMARY KEY,
    referido_id INTEGER REFERENCES referidos(id) ON DELETE CASCADE,
    total DECIMAL(10, 2) NOT NULL,
    estado VARCHAR(50) DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'procesando', 'pagada')),
    creado_en TIMESTAMP DEFAULT NOW(),
    pagada_en TIMESTAMP
);

-- Índices para liquidaciones
CREATE INDEX IF NOT EXISTS idx_liquidaciones_referido ON liquidaciones(referido_id);
CREATE INDEX IF NOT EXISTS idx_liquidaciones_estado ON liquidaciones(estado);

-- Tabla de comisiones (se crea después de liquidaciones porque la referencia)
CREATE TABLE IF NOT EXISTS comisiones (
    id SERIAL PRIMARY KEY,
    pedido_id INTEGER REFERENCES pedidos(id) ON DELETE CASCADE,
    referido_id INTEGER REFERENCES referidos(id) ON DELETE CASCADE,
    nivel INTEGER NOT NULL CHECK (nivel >= 1 AND nivel <= 3),
    porcentaje DECIMAL(5, 2) NOT NULL,
    monto DECIMAL(10, 2) NOT NULL,
    estado VARCHAR(50) DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'aprobada', 'pagada', 'revertida')),
    liquidacion_id INTEGER REFERENCES liquidaciones(id) ON DELETE SET NULL,
    creado_en TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comisiones_referido_pedido ON comisiones(referido_id, pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_pedido ON comisiones(pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_estado ON comisiones(estado);

-- Insertar productos iniciales (basados en el frontend)
INSERT INTO productos (id, nombre, descripcion, precio, categoria, img, galeria, stock) VALUES
    (1, 'Serum Revitalizante', 'Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce manchas y aporta hidratación profunda desde la primera aplicación.', 45.00, 'Serum', 'https://picsum.photos/seed/serum/600', ARRAY['https://picsum.photos/seed/serum1/600', 'https://picsum.photos/seed/serum2/600', 'https://picsum.photos/seed/serum3/600'], 50),
    (2, 'Crema Hidratante Pro', 'Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.', 32.50, 'Crema', 'https://picsum.photos/seed/crema/600', ARRAY['https://picsum.photos/seed/crema1/600', 'https://picsum.photos/seed/crema2/600', 'https://picsum.photos/seed/crema3/600'], 75),
    (3, 'Tónico Purificante', 'Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.', 28.00, 'Tónico', 'https://picsum.photos/seed/tonico/600', ARRAY['https://picsum.photos/seed/tonico1/600', 'https://picsum.photos/seed/tonico2/600', 'https://picsum.photos/seed/tonico3/600'], 60),
    (4, 'Mascarilla Nocturna', 'Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.', 50.00, 'Mascarilla', 'https://picsum.photos/seed/mascarilla/600', ARRAY['https://picsum.photos/seed/mascarilla1/600', 'https://picsum.photos/seed/mascarilla2/600', 'https://picsum.photos/seed/mascarilla3/600'], 40),
    (5, 'Serum Anti-Edad Retinol', 'Serum de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.', 58.00, 'Serum', 'https://picsum.photos/seed/retinol/600', ARRAY['https://picsum.photos/seed/retinol1/600', 'https://picsum.photos/seed/retinol2/600', 'https://picsum.photos/seed/retinol3/600'], 35),
    (6, 'Crema Contorno de Ojos', 'Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.', 26.00, 'Crema', 'https://picsum.photos/seed/ojos/600', ARRAY['https://picsum.photos/seed/ojos1/600', 'https://picsum.photos/seed/ojos2/600', 'https://picsum.photos/seed/ojos3/600'], 80),
    (7, 'Tónico Exfoliante AHA-BHA', 'Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.', 34.00, 'Tónico', 'https://picsum.photos/seed/exfoliante/600', ARRAY['https://picsum.photos/seed/exfoliante1/600', 'https://picsum.photos/seed/exfoliante2/600', 'https://picsum.photos/seed/exfoliante3/600'], 55),
    (8, 'Kit Rutina Completa', 'Kit integral con serum, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.', 129.00, 'Kit', 'https://picsum.photos/seed/kit/600', ARRAY['https://picsum.photos/seed/kit1/600', 'https://picsum.photos/seed/kit2/600', 'https://picsum.photos/seed/kit3/600'], 25)
ON CONFLICT (id) DO NOTHING;

-- Insertar usuarios con relación vacía; se vincula al referido después del seed.
INSERT INTO usuarios (email, password, rol, nombre, referido_id) VALUES
    ('admin@upse.edu.ec', '123456', 'admin', 'Administrador', NULL),
    ('cliente@upse.edu.ec', '123456', 'cliente', 'Cliente Demo', NULL)
ON CONFLICT (email) DO NOTHING;

-- Insertar referidos iniciales (estructura jerárquica aplanada con parent_id)
-- Basado en la estructura del frontend red.ts
INSERT INTO referidos (id, nombre, nivel, ventas, parent_id, codigo) VALUES
    (1, 'Ana García', 1, 1200.00, NULL, 'ANA'),
    (2, 'Luis Poveda', 1, 850.00, NULL, 'LUIS'),
    (3, 'Marta Sánchez', 1, 430.00, NULL, 'MARTA'),
    (4, 'Carlos Ruiz', 2, 500.00, 1, 'CARLOS'),
    (5, 'Sofía León', 2, 430.00, 1, 'SOFIA'),
    (6, 'Marco Díaz', 2, 380.00, 2, 'MARCO'),
    (7, 'Diana Paz', 3, 300.00, 4, 'DIANA'),
    (8, 'Cliente Demo', 2, 0.00, 2, 'CLIENTE')
ON CONFLICT (id) DO NOTHING;

UPDATE usuarios SET referido_id = 8 WHERE email = 'cliente@upse.edu.ec';

SELECT setval(pg_get_serial_sequence('productos', 'id'), COALESCE((SELECT MAX(id) FROM productos), 1), true);
SELECT setval(pg_get_serial_sequence('referidos', 'id'), COALESCE((SELECT MAX(id) FROM referidos), 1), true);
SELECT setval(pg_get_serial_sequence('usuarios', 'id'), COALESCE((SELECT MAX(id) FROM usuarios), 1), true);
SELECT setval(pg_get_serial_sequence('pedidos', 'id'), COALESCE((SELECT MAX(id) FROM pedidos), 1), true);
SELECT setval(pg_get_serial_sequence('pedido_items', 'id'), COALESCE((SELECT MAX(id) FROM pedido_items), 1), true);
