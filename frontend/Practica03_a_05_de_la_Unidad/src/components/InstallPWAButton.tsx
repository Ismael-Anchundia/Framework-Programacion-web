import { useEffect, useState } from 'react'

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed'; platform: string }>
}

export default function InstallPWAButton() {
  const [installPrompt, setInstallPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [isInstalled, setIsInstalled] = useState(false)
  const [showInstructions, setShowInstructions] = useState(false)

  useEffect(() => {
    const standalone = window.matchMedia('(display-mode: standalone)').matches
      || ('standalone' in navigator && navigator.standalone === true)
    setIsInstalled(standalone)

    const handleBeforeInstallPrompt = (event: Event) => {
      event.preventDefault()
      setInstallPrompt(event as BeforeInstallPromptEvent)
    }
    const handleAppInstalled = () => {
      setIsInstalled(true)
      setInstallPrompt(null)
    }

    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
    window.addEventListener('appinstalled', handleAppInstalled)

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
      window.removeEventListener('appinstalled', handleAppInstalled)
    }
  }, [])

  const handleInstall = async () => {
    if (!installPrompt) {
      setShowInstructions(true)
      return
    }

    const prompt = installPrompt
    setInstallPrompt(null)
    await prompt.prompt()
    const choice = await prompt.userChoice
    if (choice.outcome === 'accepted') setIsInstalled(true)
    else setShowInstructions(true)
  }

  if (isInstalled) return null

  const isAppleMobile = /iPad|iPhone|iPod/.test(navigator.userAgent)

  return (
    <>
      <button
        type="button"
        onClick={() => void handleInstall()}
        className="fixed bottom-5 right-5 z-40 inline-flex items-center gap-2 rounded-full bg-indigo-600 px-5 py-3 font-semibold text-white shadow-xl transition hover:bg-indigo-700 focus:outline-none focus:ring-4 focus:ring-indigo-300"
        aria-label="Instalar MultiCatálogo como aplicación"
      >
        <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v12m0 0 4-4m-4 4-4-4M5 17v3h14v-3" />
        </svg>
        <span>Instalar app</span>
      </button>

      {showInstructions && (
        <div className="fixed inset-0 z-[60] flex items-end justify-center bg-slate-950/50 p-4 sm:items-center">
          <section
            role="dialog"
            aria-modal="true"
            aria-labelledby="pwa-install-title"
            className="w-full max-w-md rounded-2xl bg-white p-6 text-slate-800 shadow-2xl"
          >
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 id="pwa-install-title" className="text-lg font-bold">Instala MultiCatálogo</h2>
                <p className="mt-2 text-sm leading-6 text-slate-600">
                  {isAppleMobile
                    ? 'En Safari, toca Compartir y luego “Añadir a pantalla de inicio”.'
                    : 'Abre el menú del navegador y elige “Instalar aplicación” o “Añadir a pantalla de inicio”. Si no aparece, prueba con Chrome o Edge.'}
                </p>
              </div>
              <button
                type="button"
                onClick={() => setShowInstructions(false)}
                className="rounded-lg p-2 text-slate-500 hover:bg-slate-100 hover:text-slate-800"
                aria-label="Cerrar instrucciones"
              >
                <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="m6 6 12 12M18 6 6 18" />
                </svg>
              </button>
            </div>
          </section>
        </div>
      )}
    </>
  )
}
