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

-- Generar códigos para referidos existentes
UPDATE referidos SET codigo = 'TU' WHERE id = 1;
UPDATE referidos SET codigo = 'ANA' WHERE id = 2;
UPDATE referidos SET codigo = 'LUIS' WHERE id = 3;
UPDATE referidos SET codigo = 'MARTA' WHERE id = 4;
UPDATE referidos SET codigo = 'CARLOS' WHERE id = 5;
UPDATE referidos SET codigo = 'SOFIA' WHERE id = 6;
UPDATE referidos SET codigo = 'MARCO' WHERE id = 7;

-- ALTER productos: agregar activo para borrado lógico
ALTER TABLE productos
ADD COLUMN IF NOT EXISTS activo BOOLEAN DEFAULT true;

-- ALTER pedidos: agregar estado y usuario_id si no existen
ALTER TABLE pedidos
ADD COLUMN IF NOT EXISTS estado VARCHAR(50) DEFAULT 'pendiente',
ADD COLUMN IF NOT EXISTS usuario_id INTEGER REFERENCES usuarios(id);

-- Crear tabla comisiones (ledger de comisiones)
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

-- Índices para comisiones
CREATE INDEX IF NOT EXISTS idx_comisiones_referido_pedido ON comisiones(referido_id, pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_pedido ON comisiones(pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_estado ON comisiones(estado);

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

-- Enlazar usuarios con referidos del seed
UPDATE usuarios SET nombre = 'Administrador', referido_id = 0 WHERE email = 'admin@upse.edu.ec';
UPDATE usuarios SET nombre = 'Cliente Demo', referido_id = 2 WHERE email = 'cliente@upse.edu.ec';

-- Verificación
SELECT 'usuarios' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'usuarios' ORDER BY ordinal_position;
SELECT 'referidos' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'referidos' ORDER BY ordinal_position;
SELECT 'pedidos' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'pedidos' ORDER BY ordinal_position;
SELECT 'comisiones' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'comisiones' ORDER BY ordinal_position;
SELECT 'liquidaciones' as tabla, column_name, data_type FROM information_schema.columns WHERE table_name = 'liquidaciones' ORDER BY ordinal_position;
