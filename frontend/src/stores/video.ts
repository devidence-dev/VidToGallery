import { defineStore } from 'pinia'
import { ref } from 'vue'
import { showLoadingToast, showSuccessToast, showFailToast } from 'vant'

interface VideoData {
  video_url: string
  title: string
  platform: string
  quality: string
  duration: number
  metadata: {
    method: string
    source: string
    tweet_id?: string
  }
  processed_at: string
}

interface QualitiesResponse {
  platform: string
  available_qualities: QualityOption[]
}

interface QualityOption {
  height: number
  label: string
  quality: string
  video_url?: string // Optional since it's empty until download
  width: number
}

export const useVideoStore = defineStore('video', () => {
  const url = ref('')
  const isLoading = ref(false)
  const videoData = ref<VideoData | null>(null)
  const qualitiesData = ref<QualitiesResponse | null>(null)
  const selectedQuality = ref('')
  const error = ref('')

  const getQualities = async (videoUrl: string) => {
    isLoading.value = true
    error.value = ''
    
    const toast = showLoadingToast({
      message: 'Checking qualities...',
      forbidClick: true,
      wordBreak: 'break-word',
    })
    
    try {
      const response = await fetch('/api/v1/qualities', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          url: videoUrl
        })
      })
      
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || `HTTP ${response.status}: Error getting video qualities`)
      }
      
      qualitiesData.value = await response.json()
      
      // Auto-select highest quality (first in list)
      if (qualitiesData.value?.available_qualities && qualitiesData.value.available_qualities.length > 0) {
        selectedQuality.value = qualitiesData.value.available_qualities[0].quality
      }
      
      toast.close()
      // showSuccessToast('Qualities loaded successfully')
    } catch (err) {
      toast.close()
      error.value = err instanceof Error ? err.message : 'Unknown error occurred'
      showFailToast({
        message: error.value,
        wordBreak: 'break-word',
      })
      throw err // Re-throw to let the component handle it if needed
    } finally {
      isLoading.value = false
    }
  }

  const downloadVideo = async (videoUrl: string, quality: string) => {
    isLoading.value = true
    error.value = ''
    
    const toast = showLoadingToast({
      message: 'Downloading...',
      forbidClick: true,
      wordBreak: 'break-word',
    })
    
    try {
      const response = await fetch('/api/v1/download', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          url: videoUrl,
          quality: quality
        })
      })
      
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || `HTTP ${response.status}: Error downloading video`)
      }
      
      videoData.value = await response.json()
      toast.close()
      // showSuccessToast({
      //   message: 'Video downloaded successfully',
      //   wordBreak: 'break-word',
      // })
    } catch (err) {
      toast.close()
      error.value = err instanceof Error ? err.message : 'Unknown error occurred'
      showFailToast({
        message: error.value,
        wordBreak: 'break-word',
      })
      throw err // Re-throw to let the component handle it if needed
    } finally {
      isLoading.value = false
    }
  }

  const shareVideo = async () => {
    if (videoData.value?.video_url && navigator.share) {
      try {
        await navigator.share({
          title: videoData.value.title || 'Video',
          url: videoData.value.video_url
        })
      } catch (err) {
        console.error('Error sharing video:', err)
      }
    }
  }

  const saveToGallery = async () => {
    if (!videoData.value?.video_url) return

    const toast = showLoadingToast({
      message: 'Saving to gallery...',
      forbidClick: true,
      wordBreak: 'break-word',
    })

    try {
      // Use the backend as proxy to download the video
      const response = await fetch('/api/v1/proxy-download', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          video_url: videoData.value.video_url
        })
      })
      
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || `HTTP ${response.status}: Error downloading video`)
      }
      
      const blob = await response.blob()
      
      // Check if device supports Web Share API with files
      if (navigator.canShare && navigator.canShare({ files: [new File([blob], 'video.mp4', { type: 'video/mp4' })] })) {
        const fileName = videoData.value.title 
          ? `${videoData.value.title.trim().replace(/[^\w\s-]/g, '').replace(/\s+/g, '_')}.mp4`
          : 'video.mp4'
        
        const file = new File([blob], fileName, { type: 'video/mp4' })
        
        await navigator.share({
          files: [file],
          title: videoData.value.title || 'Video'
        })
        
        toast.close()
        // showSuccessToast({
        //   message: 'Video saved to gallery',
        //   wordBreak: 'break-word',
        // })
      } else {
        throw new Error('Your device does not support saving to gallery')
      }
    } catch (err) {
      toast.close()
      const errorMessage = err instanceof Error ? err.message : 'Error saving to gallery'
      showFailToast({
        message: errorMessage,
        wordBreak: 'break-word',
      })
      console.error('Error saving to gallery:', err)
      throw new Error(errorMessage)
    }
  }

  const downloadFile = async () => {
    if (!videoData.value?.video_url) {
      throw new Error('No video URL available. Please download the video first.')
    }

    const toast = showLoadingToast({
      message: 'Downloading file...',
      forbidClick: true,
      wordBreak: 'break-word',
    })

    try {
      // Use the backend as proxy to download the video
      const response = await fetch('/api/v1/proxy-download', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          video_url: videoData.value.video_url
        })
      })
      
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || `HTTP ${response.status}: Error downloading video`)
      }
      
      const blob = await response.blob()
      
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      
      const fileName = videoData.value.title 
        ? `${videoData.value.title.trim().replace(/[^\w\s-]/g, '').replace(/\s+/g, '_')}.mp4`
        : 'video.mp4'
      
      link.download = fileName
      document.body.appendChild(link)
      link.click()
      
      // Clean up
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      
      toast.close()
      // showSuccessToast({
      //   message: 'File downloaded successfully',
      //   wordBreak: 'break-word',
      // })
    } catch (err) {
      toast.close()
      const errorMessage = err instanceof Error ? err.message : 'Error downloading file'
      showFailToast({
        message: errorMessage,
        wordBreak: 'break-word',
      })
      console.error('Error downloading file:', err)
      throw new Error(errorMessage)
    }
  }

  const reset = () => {
    url.value = ''
    videoData.value = null
    qualitiesData.value = null
    selectedQuality.value = ''
    error.value = ''
    isLoading.value = false
  }

  return {
    url,
    selectedQuality,
    isLoading,
    videoData,
    qualitiesData,
    error,
    getQualities,
    downloadVideo,
    shareVideo,
    saveToGallery,
    downloadFile,
    reset
  }
})
