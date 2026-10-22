-- Migración 003: Seguridad - bcrypt, índices y mejoras
-- Ejecutar manualmente: psql -h localhost -p 5434 -U multicatalogo -d multicatalogo -f database/003_seguridad.sql

-- Habilitar pgcrypto para bcrypt
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Actualizar tabla usuarios para usar bcrypt
ALTER TABLE usuarios ALTER COLUMN password TYPE TEXT;

-- Crear índices para mejorar rendimiento
CREATE INDEX IF NOT EXISTS idx_referidos_parent_id ON referidos(parent_id);
CREATE INDEX IF NOT EXISTS idx_pedidos_usuario_id ON pedidos(usuario_id);
CREATE INDEX IF NOT EXISTS idx_pedido_items_pedido_id ON pedido_items(pedido_id);
CREATE INDEX IF NOT EXISTS idx_productos_categoria ON productos(categoria);
CREATE INDEX IF NOT EXISTS idx_usuarios_email ON usuarios(email);

-- Función para hashear passwords con bcrypt
CREATE OR REPLACE FUNCTION hash_password(password TEXT) RETURNS TEXT AS $$
BEGIN
    RETURN crypt(password, gen_salt('bf'));
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Función para verificar passwords con bcrypt
CREATE OR REPLACE FUNCTION verify_password(password TEXT, hashed TEXT) RETURNS BOOLEAN AS $$
BEGIN
    RETURN (hashed = crypt(password, hashed));
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Actualizar passwords existentes con bcrypt
UPDATE usuarios SET password = hash_password(password) WHERE password NOT LIKE '$2a$%';

-- Verificar que los passwords estén hasheados
SELECT email, password FROM usuarios LIMIT 5;
