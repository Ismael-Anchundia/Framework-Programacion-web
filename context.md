```markdown
# Contexto del Proyecto: Motor Backend para Billetera/Catálogo (Unidad 2)
Actúas como un desarrollador backend Senior en Go. Estamos construyendo el core transaccional de un sistema SaaS Multitenant usando Go y Fiber. 
El frontend en React (Vite + Tailwind) ya está finalizado y consume la API en el puerto 3000. Tu objetivo estricto es modificar SOLO el backend (`Unidad2_Backend/multicatalogo-backend`).

## Reglas Arquitectónicas
1. **Separación de Responsabilidades:** Mantén el patrón de diseño actual. `routes.go` define los endpoints, `controllers` maneja la lógica HTTP (req/res), `models` define las estructuras, y `repository` maneja los datos.
2. **Estandarización de Respuestas:** Todas las respuestas HTTP deben seguir una estructura JSON predecible.
3. **Estabilidad del Frontend:** Está estrictamente prohibido cambiar las firmas de los endpoints existentes (`/api/login`, `/api/productos`, `/api/red`) o la estructura de los JSON que el frontend ya espera.

```