<template>
  <div class="video-input">
    <div class="input-row">
      <van-cell-group class="input-field">
        <van-field v-model="localUrl" label="" placeholder="Paste video URL here" @input="onInput"
          @update:model-value="onModelValueUpdate">
          <template #left-icon>
            <van-icon name="video" />
          </template>
        </van-field>
      </van-cell-group>

      <div class="field-buttons">
        <van-button plain type="primary" @click="pasteFromClipboard" size="small">
          <van-icon name="notes" />
          Paste
        </van-button>

        <van-button plain type="primary" @click="resetForm" size="small">
          <van-icon name="replay" />
          Clear
        </van-button>
      </div>
    </div>

    <div class="button-container">
      <van-button type="primary" block :disabled="!localUrl" @click="handleGetQualities">
        <van-icon name="search" />
        Check Qualities
      </van-button>
    </div>

    <div v-if="availableQualities.length > 0" class="quality-section">
      <van-cell-group>
        <van-field name="quality" label="Available quality" class="quality-field">
          <template #left-icon>
            <van-icon name="play-circle" />
          </template>
          <template #input>
            <van-radio-group v-model="selectedQuality" direction="vertical">
              <van-radio v-for="qualityOption in availableQualities" :key="qualityOption.quality"
                :name="qualityOption.quality">
                {{ qualityOption.label }}
              </van-radio>
            </van-radio-group>
          </template>
        </van-field>
      </van-cell-group>

      <div class="download-container">
        <van-button type="primary" block :disabled="!selectedQuality" @click="handleDownload">
          <van-icon name="setting" />
          Process Video
        </van-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useVideoStore } from '@/stores/video'
import { showSuccessToast, showFailToast } from 'vant'

const videoStore = useVideoStore()
const { qualitiesData, selectedQuality } = storeToRefs(videoStore)

const localUrl = ref('')

// Computed property to get available qualities from qualitiesData
const availableQualities = computed(() => {
  const qualities = qualitiesData.value?.available_qualities || []

  // Sort qualities from best to worst
  const qualityOrder = ['best', 'best[height<=1350]', 'best[height<=900]', 'best[height<=800]', 'worst']

  return qualities.sort((a, b) => {
    const indexA = qualityOrder.indexOf(a.quality)
    const indexB = qualityOrder.indexOf(b.quality)

    // If quality not found in predefined order, put it at the end
    if (indexA === -1 && indexB === -1) return 0
    if (indexA === -1) return 1
    if (indexB === -1) return -1

    return indexA - indexB
  })
})

watch(localUrl, (newUrl) => {
  videoStore.url = newUrl
  // Reset data when URL changes
  videoStore.reset()
})

const handleGetQualities = async () => {
  if (localUrl.value) {
    try {
      await videoStore.getQualities(localUrl.value)
      // selectedQuality is auto-selected in the store
    } catch (err) {
      // Error is already handled in the store with toast notifications
      console.error('Error getting qualities:', err)
    }
  }
}

const handleDownload = async () => {
  if (localUrl.value && selectedQuality.value) {
    try {
      await videoStore.downloadVideo(localUrl.value, selectedQuality.value)
      // After getting the video URL, VideoPreview component will show with action buttons
    } catch (err) {
      // Error is already handled in the store with toast notifications
      console.error('Error downloading video:', err)
    }
  }
}

const onInput = (value: string) => {
  if (value === '') {
    videoStore.url = ''
    videoStore.reset()
  }
}

const onModelValueUpdate = (value: string) => {
  if (value === '') {
    videoStore.url = ''
    videoStore.reset()
  }
}

const pasteFromClipboard = async () => {
  try {
    // Check if the Clipboard API is available
    if (!navigator.clipboard) {
      showFailToast({message: 'Clipboard API not available in this browser', wordBreak: 'break-word',})
      return
    }

    // Request clipboard permission and read text
    const text = await navigator.clipboard.readText()

    if (text && text.trim()) {
      localUrl.value = text.trim()
      videoStore.url = text.trim()
      showSuccessToast('URL pasted successfully!')
    } else {
      showFailToast({message: 'Clipboard is empty or contains no text', wordBreak: 'break-word',})
    }
  } catch (err) {
    console.error('Failed to read clipboard:', err)

    // Handle specific permission errors
    if (err instanceof Error) {
      if (err.name === 'NotAllowedError') {
        showFailToast({
          message: 'Clipboard access denied. Please allow clipboard permissions in your browser.',
          wordBreak: 'break-word',
        })
      } else if (err.name === 'NotFoundError') {
        showFailToast({message: 'No text found in clipboard', wordBreak: 'break-word',})
      } else {
        showFailToast({
          message: 'Failed to access clipboard. Try copying the URL again.',
          wordBreak: 'break-word',
        })
      }
    } else {
      showFailToast({ message: 'Failed to access clipboard', wordBreak: 'break-word', })
    }
  }
}

const resetForm = () => {
  localUrl.value = ''
  videoStore.url = ''
  videoStore.reset()
}
</script>

<style scoped>
.video-input {
  padding: 16px;
}

.input-row {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.input-field {
  flex: 1;
}

.field-buttons {
  display: flex;
  gap: 8px;
  flex-direction: row;
  align-items: flex-end;
}

.field-buttons .van-button {
  min-width: 70px;
  white-space: nowrap;
}

.button-container {
  margin-top: 16px;
}

.quality-section {
  margin-top: 24px;
}

.download-container {
  margin-top: 16px;
}

.quality-field :deep(.van-field__left-icon) {
  align-self: center;
}

.quality-field :deep(.van-field__label) {
  width: auto;
  align-self: center;
}

.van-icon-video {
  color: #667eea;
}
</style>
