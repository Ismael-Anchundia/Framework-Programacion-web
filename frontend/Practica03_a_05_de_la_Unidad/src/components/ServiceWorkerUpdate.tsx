import { useState } from 'react'
import { useRegisterSW } from 'virtual:pwa-register/react'

export default function ServiceWorkerUpdate() {
  const [registrationError, setRegistrationError] = useState('')
  const {
    needRefresh: [needRefresh, setNeedRefresh],
    offlineReady: [offlineReady, setOfflineReady],
    updateServiceWorker,
  } = useRegisterSW({
    onRegisterError(error) {
      console.error('No se pudo registrar el service worker:', error)
      setRegistrationError('No se pudo activar el modo PWA en este navegador. Puedes seguir usando la tienda con conexión.')
    },
  })

  if (!needRefresh && !offlineReady && !registrationError) return null

  return (
    <div className="fixed bottom-24 right-4 bg-indigo-600 text-white px-6 py-4 rounded-lg shadow-lg z-50 max-w-sm">
      {registrationError ? (
        <>
          <p role="alert" className="text-sm">{registrationError}</p>
          <button type="button" onClick={() => setRegistrationError('')} className="mt-2 text-sm underline">Cerrar</button>
        </>
      ) : needRefresh ? (
        <>
          <p className="font-semibold mb-2">Nueva versión disponible</p>
          <p className="text-sm mb-3 opacity-90">Actualiza cuando termines lo que estás haciendo.</p>
          <button type="button" onClick={() => void updateServiceWorker(true)} className="w-full bg-white text-indigo-600 font-semibold py-2 px-4 rounded hover:bg-indigo-50 transition-colors">Actualizar</button>
          <button type="button" onClick={() => setNeedRefresh(false)} className="mt-2 w-full text-sm text-white/90 hover:text-white">Ahora no</button>
        </>
      ) : (
        <>
          <p className="font-semibold mb-2">App lista para usar sin conexión</p>
          <button type="button" onClick={() => setOfflineReady(false)} className="text-sm underline">Entendido</button>
        </>
      )}
    </div>
  )
}
