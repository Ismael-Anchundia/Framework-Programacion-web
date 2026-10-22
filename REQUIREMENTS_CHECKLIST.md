# Checklist de requisitos — PWA MultiCatálogo

Este checklist define el plan recomendado para completar la conversión del frontend a PWA. El archivo no existía en el repositorio; se creó en la raíz para registrar las fases solicitadas.

## Estado actual verificado

- [x] `vite-plugin-pwa` está instalado y genera manifest y service worker en el build de producción.
- [x] El manifest declara idioma, nombre, colores, modo `standalone`, iconos PNG `any` y `maskable`, y el icono de Apple está presente.
- [x] Workbox precachea la interfaz y permite fallback SPA; hay caché de catálogo público e imágenes.
- [x] La actualización del service worker usa aviso manual y reporta errores de registro.
- [x] El carrito se conserva en `localStorage`.
- [x] La app muestra el estado offline; login, registro y checkout se bloquean sin conexión, y el checkout conserva el carrito ante fallos.
- [x] La API usa `localhost` solo en desarrollo y admite URL configurable o rutas relativas de mismo origen en producción.
- [x] El backend permite configurar los orígenes CORS de despliegue mediante `CORS_ALLOWED_ORIGINS`.

## Alcance recomendado

PWA instalable para la tienda, con shell y catálogo público consultables sin conexión cuando ya se hayan visitado. El carrito puede conservarse localmente. Inicio de sesión, red privada, pedidos, liquidaciones y checkout requieren conexión; nunca se cachean respuestas autenticadas. No se encolan ni simulan compras offline.

Objetivo de validación: Chrome/Edge en escritorio y Android, y Safari en iOS. Se usará la opción nativa del navegador para instalar; no se añadirá un prompt propio.

## Fase 0 — Alcance y configuración de despliegue

- [x] Definir los navegadores y dispositivos objetivo y el alcance offline recomendado.
- [x] Configurar `VITE_API_URL` por entorno; producción no apunta a `localhost` (API del mismo origen por defecto).
- [x] Permitir configurar CORS del backend y documentar frontend/API en desarrollo y despliegue.
- [ ] Confirmar HTTPS en despliegue; para pruebas locales, usar `localhost`/`127.0.0.1`.

**Criterio de aceptación:** el build de producción apunta a la API correcta y la app abre desde un origen seguro sin errores de CORS.

## Fase 1 — Instalabilidad e identidad visual

- [x] Alinear `lang` del manifest con español y completar metadatos de instalación.
- [x] Generar iconos raster PNG de 192×192 y 512×512, más una variante `maskable`, conservando el SVG como favicon.
- [x] Añadir `apple-touch-icon.png` y verificar que las rutas del manifest y `includeAssets` existan en el build.
- [x] Definir `name`, `short_name`, `start_url`, `scope`, `display`, colores y propósito de cada icono.
- [x] Usar las opciones de instalación nativas del navegador; no se añade un prompt propio.

**Criterio de aceptación:** el manifest se carga sin errores y Chrome/Edge permiten instalar la app con iconos correctos; Safari/iOS muestra el icono y nombre esperados.

## Fase 2 — Registro del service worker y actualización segura

- [x] Registrar el service worker en el build de producción y documentar que `npm run dev` no representa esta validación.
- [x] Elegir actualización con aviso manual y alinear `registerType` con `ServiceWorkerUpdate`.
- [x] Evitar recargas automáticas durante checkout y formularios; el usuario decide cuándo actualizar.
- [x] Mostrar el aviso de actualización/offline-ready y errores de registro comprensibles.

**Criterio de aceptación:** una versión nueva se activa según la estrategia elegida, sin bucles de recarga ni pérdida inesperada del estado del usuario.

## Fase 3 — Shell offline y navegación

- [x] Definir el precache de HTML, JS, CSS, fuentes e iconos de la app.
- [x] Configurar fallback de navegación SPA para rutas profundas y excluir `/api` del fallback.
- [x] Mostrar estado offline claro con acción para reintentar conexión.
- [x] Mantener fuera del caché las rutas privadas y mutaciones de la API (`POST`, `PUT`, `PATCH`, `DELETE`).

**Criterio de aceptación:** después de una primera carga online, el shell abre offline y las rutas SPA cargan sin quedar en una página del navegador; las vistas que requieren servidor explican que no hay conexión.

## Fase 4 — Catálogo público e imágenes en caché

- [x] Aplicar `NetworkFirst` con espera limitada y caducidad a `GET /api/productos` y detalle público.
- [x] Excluir solicitudes con `Authorization` y todas las rutas privadas.
- [x] Cachear imágenes locales y las URLs reales de Picsum, con límites y caducidad.
- [x] Gestionar imágenes fallidas con placeholders locales.
- [x] Avisar cuando se está offline que el catálogo disponible en caché puede estar desactualizado.

**Criterio de aceptación:** tras visitar catálogo y detalle online, ambos muestran datos e imágenes disponibles offline; las peticiones autenticadas y respuestas de usuario no aparecen en Cache Storage.

## Fase 5 — Comportamiento offline de cuenta y compra

- [x] Mostrar indicador de conexión y actualizarlo ante eventos `online`/`offline`.
- [x] Conservar/restaurar el carrito local e informar que el servidor revalida precio y stock al comprar.
- [x] Bloquear login, registro y checkout sin red; las vistas privadas explican que requieren conexión.
- [x] No borrar el carrito por una falla de red; vaciarlo solo tras confirmar el pedido en el servidor.
- [x] Evitar guardar tokens o datos privados en Cache Storage; el service worker solo enruta recursos estáticos, productos públicos sin autorización e imágenes.

**Criterio de aceptación:** una desconexión no causa pérdida del carrito ni presenta una compra como exitosa; al reconectar, el usuario puede reintentar y el backend sigue siendo la fuente de verdad.

## Fase 6 — Validación de instalación, offline y publicación

- [x] Ejecutar `npm run build` y `npm run preview`; comprobar HTTP 200 para manifest, service worker, iconos y ruta SPA profunda.
- [ ] Probar instalación, primera carga, recarga, rutas profundas, actualización y desinstalación en escritorio y móvil objetivo.
- [ ] Probar modo offline después de una visita online: shell, catálogo, imágenes, carrito y bloqueo de operaciones privadas.
- [ ] Verificar en DevTools que Cache Storage contiene solo recursos públicos previstos y que no almacena JWT ni respuestas privadas.
- [ ] Ejecutar Lighthouse/PWA audit y corregir errores de manifest, consola, rendimiento y accesibilidad que bloqueen el alcance definido.
- [x] Documentar despliegue HTTPS, variables de entorno, CORS, actualización del service worker y límites del modo offline.

**Criterio de aceptación:** todas las pruebas anteriores pasan en el build desplegable y una instalación nueva puede iniciar, navegar por catálogo y recuperar conectividad sin datos privados obsoletos.
