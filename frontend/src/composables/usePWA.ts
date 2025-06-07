import { ref } from 'vue'

interface BeforeInstallPromptEvent extends Event {
  prompt(): Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

interface NavigatorWithStandalone extends Navigator {
  standalone?: boolean
}

export function usePWA() {
  const deferredPrompt = ref<BeforeInstallPromptEvent | null>(null)
  const showInstallPrompt = ref(false)

  // Listen for the beforeinstallprompt event
  window.addEventListener('beforeinstallprompt', (e) => {
    // Prevent the mini-infobar from appearing on mobile
    e.preventDefault()
    // Stash the event so it can be triggered later
    deferredPrompt.value = e as BeforeInstallPromptEvent
    showInstallPrompt.value = true
  })

  // Handle the install button click
  const installApp = async () => {
    if (!deferredPrompt.value) return

    // Show the install prompt
    deferredPrompt.value.prompt()
    
    // Wait for the user to respond to the prompt
    const { outcome } = await deferredPrompt.value.userChoice
    
    // Clear the deferredPrompt so it can only be used once
    deferredPrompt.value = null
    showInstallPrompt.value = false
    
    return outcome
  }

  // Check if app is already installed
  const isInstalled = ref(
    window.matchMedia('(display-mode: standalone)').matches ||
    (window.navigator as NavigatorWithStandalone).standalone === true
  )

  return {
    showInstallPrompt,
    installApp,
    isInstalled
  }
}
