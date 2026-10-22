# MultiCatálogo PWA

## Desarrollo

Desde esta carpeta, inicia el backend en `http://localhost:3000` y ejecuta:

```powershell
npm install
npm run dev
```

Vite usa `http://localhost:3000` por defecto durante desarrollo. El service worker no se registra en `npm run dev`; esto evita que el caché de la PWA oculte cambios locales.

## Probar la versión PWA de producción

```powershell
$env:VITE_API_URL = "http://localhost:3000"
npm run build
npm run preview
```

Abre la dirección local que muestre Vite Preview, normalmente `http://localhost:4173`. Después de cargarla una vez con internet:

1. En Chrome/Edge, usa la opción **Instalar MultiCatálogo** del navegador. En Safari de iOS, usa **Compartir → Añadir a pantalla de inicio**.
2. Abre DevTools → Application para revisar Manifest, Service Worker y Cache Storage.
3. En DevTools → Network, activa **Offline** y recarga. El shell, las rutas SPA, el catálogo ya visitado, sus imágenes disponibles y el carrito local deben seguir funcionando.
4. Login, registro, pedidos, red, liquidaciones y checkout necesitan conexión. El checkout offline se bloquea y el carrito se conserva.
5. Para probar actualizaciones, genera y sirve un nuevo build; la app debe mostrar el aviso y esperar a que el usuario pulse **Actualizar**.

La instalación y el service worker requieren un contexto seguro: HTTPS en despliegue o `localhost` para pruebas locales. Un teléfono accediendo por HTTP a la IP LAN del ordenador no es un contexto seguro.

## URL de API y publicación

Copia `.env.example` como `.env.production` y define `VITE_API_URL`:

- Si frontend y backend comparten origen y el servidor/proxy dirige `/api` al backend, déjalo vacío.
- Si la API tiene un origen público distinto, usa su URL HTTPS, por ejemplo `https://api.ejemplo.edu`.

En el backend, establece `CORS_ALLOWED_ORIGINS` con el origen exacto del frontend. Admite varios orígenes separados por comas, por ejemplo `https://tienda.ejemplo.edu,https://admin.ejemplo.edu`. Localmente se permiten Vite (`5173`) y Vite Preview (`4173`) en `localhost` y `127.0.0.1`.

El service worker cachea recursos de la interfaz, respuestas públicas de productos sin cabecera `Authorization` e imágenes. No cachea rutas privadas, tokens ni operaciones de escritura. Los datos de catálogo offline pueden estar desactualizados; PostgreSQL sigue siendo la fuente de verdad para stock y precio durante checkout.
