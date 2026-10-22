-- Migración 004: Back-office MultiCatálogo
-- Migración idempotente para implementar funcionalidades de back-office
-- Ejecutar manualmente: psql -h localhost -p 5434 -U multicatalogo -d multicatalogo -f database/004_backoffice.sql

-- ALTER usuarios: agregar campos para back-office
ALTER TABLE usuarios
ADD COLUMN IF NOT EXISTS nombre VARCHAR(255),
ADD COLUMN IF NOT EXISTS activo BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS referido_id INTEGER REFERENCES referidos(id),
ADD COLUMN IF NOT EXISTS creado_en TIMESTAMP DEFAULT NOW();

-- ALTER referidos: agregar código de invitación único
ALTER TABLE referidos
ADD COLUMN IF NOT EXISTS codigo VARCHAR(20) UNIQUE;

-- Generar códigos para el seed sin pisar los ya existentes
UPDATE referidos SET codigo = CASE id
    WHEN 1 THEN 'ANA'
    WHEN 2 THEN 'LUIS'
    WHEN 3 THEN 'MARTA'
    WHEN 4 THEN 'CARLOS'
    WHEN 5 THEN 'SOFIA'
    WHEN 6 THEN 'MARCO'
    WHEN 7 THEN 'DIANA'
    ELSE codigo
END
WHERE codigo IS NULL AND id BETWEEN 1 AND 7;

-- ALTER productos: agregar activo para borrado lógico
ALTER TABLE productos
ADD COLUMN IF NOT EXISTS activo BOOLEAN DEFAULT true;

-- ALTER pedidos: agregar estado y usuario_id si no existen
ALTER TABLE pedidos
ADD COLUMN IF NOT EXISTS estado VARCHAR(50) DEFAULT 'pendiente',
ADD COLUMN IF NOT EXISTS usuario_id INTEGER REFERENCES usuarios(id),
ADD COLUMN IF NOT EXISTS creado_en TIMESTAMP DEFAULT NOW();

UPDATE pedidos SET creado_en = created_at WHERE creado_en IS NULL;
ALTER TABLE pedidos DROP CONSTRAINT IF EXISTS pedidos_estado_check;
UPDATE pedidos SET estado = 'preparando' WHERE estado = 'procesando';
UPDATE pedidos SET estado = 'entregado' WHERE estado = 'completado';
ALTER TABLE pedidos ADD CONSTRAINT pedidos_estado_check CHECK (estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado'));

-- Crear tabla liquidaciones
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

-- Crear tabla comisiones después de liquidaciones porque la referencia
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

-- Enlazar usuarios con referidos del seed
UPDATE usuarios SET nombre = 'Administrador', referido_id = NULL WHERE email = 'admin@upse.edu.ec';
-- Bases existentes pueden tener una secuencia atrasada respecto a los IDs cargados manualmente.
SELECT setval(pg_get_serial_sequence('referidos', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM referidos), 1), 1), true);
INSERT INTO referidos (nombre, nivel, ventas, parent_id, codigo)
SELECT 'Cliente Demo', 2, 0, 2, 'CLIENTE'
WHERE NOT EXISTS (SELECT 1 FROM referidos WHERE codigo = 'CLIENTE');
UPDATE usuarios SET nombre = 'Cliente Demo', referido_id = (SELECT id FROM referidos WHERE codigo = 'CLIENTE')
WHERE email = 'cliente@upse.edu.ec';

SELECT setval(pg_get_serial_sequence('referidos', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM referidos), 1), 1), true);
SELECT setval(pg_get_serial_sequence('usuarios', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM usuarios), 1), 1), true);

-- Verificación
SELECT 'usuarios' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'usuarios' ORDER BY ordinal_position;
SELECT 'referidos' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'referidos' ORDER BY ordinal_position;
SELECT 'pedidos' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'pedidos' ORDER BY ordinal_position;
SELECT 'comisiones' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'comisiones' ORDER BY ordinal_position;
SELECT 'liquidaciones' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'liquidaciones' ORDER BY ordinal_position;
