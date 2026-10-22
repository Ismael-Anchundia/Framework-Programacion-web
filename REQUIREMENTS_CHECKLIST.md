# Checklist de requisitos — PWA MultiCatálogo

Este checklist define el plan recomendado para completar la conversión del frontend a PWA. El archivo no existía en el repositorio; se creó en la raíz para registrar las fases solicitadas.

## Estado actual verificado

- [x] `vite-plugin-pwa` está instalado y genera manifest y service worker en el build de producción.
- [x] El manifest ya declara nombre, colores, modo `standalone` e iconos SVG de varios tamaños.
- [x] Workbox precachea recursos estáticos y hay un componente para detectar actualizaciones.
- [x] El carrito se conserva en `localStorage`.
- [ ] La app no ofrece una pantalla/estado offline ni una política de datos offline definida.
- [ ] El patrón de imágenes externas de productos no coincide necesariamente con la regla actual de caché por extensión.
- [ ] `autoUpdate` está habilitado mientras el componente presenta una actualización con acción manual; falta unificar el comportamiento.
- [ ] Se referencia `apple-touch-icon.png`, pero ese archivo no está en `public`; los iconos actuales son SVG.
- [ ] La API usa `http://localhost:3000` por defecto, no una dirección válida para una instalación en otro dispositivo o producción.

## Alcance recomendado

PWA instalable para la tienda, con shell y catálogo público consultables sin conexión cuando ya se hayan visitado. El carrito puede conservarse localmente. Inicio de sesión, red privada, pedidos, liquidaciones y checkout requieren conexión; nunca se cachean respuestas autenticadas. No se encolan ni simulan compras offline.

## Fase 0 — Alcance y configuración de despliegue

- [ ] Definir los navegadores y dispositivos objetivo y confirmar el alcance offline de este documento.
- [ ] Configurar `VITE_API_URL` por entorno; producción no debe apuntar a `localhost`.
- [ ] Configurar CORS del backend para el origen HTTPS de producción y documentar cómo ejecutar frontend y backend en desarrollo.
- [ ] Confirmar HTTPS en despliegue; para pruebas locales, usar `localhost`/`127.0.0.1`.

**Criterio de aceptación:** el build de producción apunta a la API correcta y la app abre desde un origen seguro sin errores de CORS.

## Fase 1 — Instalabilidad e identidad visual

- [ ] Alinear `lang` del manifest con español y completar metadatos de instalación.
- [ ] Generar iconos raster PNG de 192×192 y 512×512, más una variante `maskable`, conservando el SVG como favicon si se desea.
- [ ] Añadir `apple-touch-icon.png` y verificar que todas las rutas del manifest y `includeAssets` existan.
- [ ] Verificar `name`, `short_name`, `start_url`, `scope`, `display`, colores y propósito de cada icono.
- [ ] Decidir si se necesita una guía/botón de instalación dentro de la app además del prompt nativo del navegador.

**Criterio de aceptación:** el manifest se carga sin errores y Chrome/Edge permiten instalar la app con iconos correctos; Safari/iOS muestra el icono y nombre esperados.

## Fase 2 — Registro del service worker y actualización segura

- [ ] Confirmar el registro automático del service worker en el build de producción y documentar que `npm run dev` no representa esta validación.
- [ ] Elegir una sola estrategia de actualización: actualización automática o aviso con acción manual; alinear `registerType` y `ServiceWorkerUpdate`.
- [ ] Evitar recargas durante checkout o formularios con cambios sin guardar si se adopta actualización automática.
- [ ] Mostrar estado de actualización y errores de registro de forma comprensible.

**Criterio de aceptación:** una versión nueva se activa según la estrategia elegida, sin bucles de recarga ni pérdida inesperada del estado del usuario.

## Fase 3 — Shell offline y navegación

- [ ] Definir y verificar el precache de HTML, JS, CSS, fuentes e iconos de la app.
- [ ] Configurar fallback de navegación SPA para rutas profundas como `/catalogo`, `/mis-pedidos` y `/admin/productos`.
- [ ] Crear una vista offline clara con acción para reintentar conexión.
- [ ] Mantener fuera del caché las rutas autenticadas y mutaciones de la API (`POST`, `PUT`, `PATCH`, `DELETE`).

**Criterio de aceptación:** después de una primera carga online, el shell abre offline y las rutas SPA cargan sin quedar en una página del navegador; las vistas que requieren servidor explican que no hay conexión.

## Fase 4 — Catálogo público e imágenes en caché

- [ ] Definir estrategia y caducidad para `GET /api/productos` y detalle público: priorizar datos recientes online y usar la última respuesta válida offline.
- [ ] Excluir respuestas con autorización y todas las rutas privadas de usuario.
- [ ] Ajustar la regla de Workbox para las URLs reales de imágenes (incluidas URLs externas sin extensión) con límites y caducidad.
- [ ] Gestionar imágenes fallidas con un placeholder local.
- [ ] Mostrar cuándo el catálogo es una copia en caché y cuándo se actualizó por última vez, si la estrategia conserva datos antiguos.

**Criterio de aceptación:** tras visitar catálogo y detalle online, ambos muestran datos e imágenes disponibles offline; las peticiones autenticadas y respuestas de usuario no aparecen en Cache Storage.

## Fase 5 — Comportamiento offline de cuenta y compra

- [ ] Mostrar indicador de conexión y recuperarla ante eventos `online`/`offline`.
- [ ] Conservar y restaurar el carrito local (ya persiste en `localStorage`) y explicar que precio/stock se revalidan al volver a conectar.
- [ ] Deshabilitar o bloquear login, registro, consultas privadas y checkout cuando no hay red, con mensajes y reintento.
- [ ] No borrar el carrito ante una falla de red durante checkout; solo vaciarlo cuando el servidor confirme el pedido.
- [ ] Revisar el almacenamiento local de token y datos de usuario para no duplicar ni ampliar datos sensibles en Cache Storage.

**Criterio de aceptación:** una desconexión no causa pérdida del carrito ni presenta una compra como exitosa; al reconectar, el usuario puede reintentar y el backend sigue siendo la fuente de verdad.

## Fase 6 — Validación de instalación, offline y publicación

- [ ] Ejecutar `npm run build` y `npm run preview`; probar desde origen seguro y comprobar manifest/service worker en DevTools.
- [ ] Probar instalación, primera carga, recarga, rutas profundas, actualización y desinstalación en escritorio y móvil objetivo.
- [ ] Probar modo offline después de una visita online: shell, catálogo, imágenes, carrito y bloqueo de operaciones privadas.
- [ ] Verificar en DevTools que Cache Storage contiene solo recursos públicos previstos y que no almacena JWT ni respuestas privadas.
- [ ] Ejecutar Lighthouse/PWA audit y corregir errores de manifest, consola, rendimiento y accesibilidad que bloqueen el alcance definido.
- [ ] Documentar despliegue HTTPS, variables de entorno, CORS, actualización del service worker y límites del modo offline.

**Criterio de aceptación:** todas las pruebas anteriores pasan en el build desplegable y una instalación nueva puede iniciar, navegar por catálogo y recuperar conectividad sin datos privados obsoletos.
