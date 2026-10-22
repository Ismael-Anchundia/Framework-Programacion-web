import { useEffect, useState } from 'react'

export default function ServiceWorkerUpdate() {
  const [showReload, setShowReload] = useState(false)
  const [waitingWorker, setWaitingWorker] = useState<ServiceWorker | null>(null)

  const onSWUpdate = (registration: ServiceWorkerRegistration) => {
    const w = registration.waiting
    if (w) {
      setWaitingWorker(w)
      setShowReload(true)
    }
  }

  useEffect(() => {
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.ready.then((registration) => {
        registration.addEventListener('updatefound', () => {
          const installingWorker = registration.installing
          if (installingWorker) {
            installingWorker.addEventListener('statechange', () => {
              if (installingWorker.state === 'installed' && navigator.serviceWorker.controller) {
                onSWUpdate(registration)
              }
            })
          }
        })
      })
    }
  }, [])

  const reloadPage = () => {
    waitingWorker?.postMessage({ type: 'SKIP_WAITING' })
    setShowReload(false)
    window.location.reload()
  }

  if (!showReload) return null

  return (
    <div className="fixed bottom-4 right-4 bg-indigo-600 text-white px-6 py-4 rounded-lg shadow-lg z-50 max-w-sm">
      <p className="font-semibold mb-2">Nueva versión disponible</p>
      <p className="text-sm mb-3 opacity-90">
        Hay una actualización lista. Haz clic para recargar.
      </p>
      <button
        onClick={reloadPage}
        className="w-full bg-white text-indigo-600 font-semibold py-2 px-4 rounded hover:bg-indigo-50 transition-colors"
      >
        Actualizar
      </button>
    </div>
  )
}
