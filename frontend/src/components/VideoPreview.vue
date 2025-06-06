<template>
  <div v-if="videoData" class="video-preview">
    <van-card
      :title="videoData.title"
      :desc="`${videoData.platform} • ${formatDuration(videoData.duration)}`"
      :thumb="videoData.metadata?.thumbnail"
    >
      <template #tags>
        <van-tag type="primary">{{ videoData.quality }}</van-tag>
      </template>
      
      <template #footer>
        <van-row gutter="8">
          <van-col span="8">
            <van-button 
              type="primary" 
              size="small"
              block
              @click="shareVideo"
            >
              <van-icon name="share" />
              Share
            </van-button>
          </van-col>
          <van-col span="8">
            <van-button 
              plain
              type="primary"
              size="small"
              block
              @click="saveToGallery"
            >
              <van-icon name="photo" />
               Gallery
            </van-button>
          </van-col>
          <van-col span="8">
            <van-button 
              plain
              type="primary"
              size="small"
              block
              @click="downloadFile"
            >
              <van-icon name="down" />
              Download
            </van-button>
          </van-col>
        </van-row>
      </template>
    </van-card>
    
    <div class="video-container">
      <video 
        ref="videoElement"
        :src="videoData.video_url"
        controls
        playsinline
        webkit-playsinline
        class="video-player"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useVideoStore } from '@/stores/video'

const videoStore = useVideoStore()
const { videoData } = storeToRefs(videoStore)

const formatDuration = (seconds: number): string => {
  if (!seconds) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const shareVideo = async () => {
  await videoStore.shareVideo()
}

const saveToGallery = async () => {
  try {
    await videoStore.saveToGallery()
  } catch (error) {
    console.error('Error saving to gallery:', error)
  }
}

const downloadFile = async () => {
  try {
    await videoStore.downloadFile()
  } catch (error) {
    console.error('Error downloading file:', error)
  }
}
</script>

<style scoped>
.video-preview {
  padding: 16px;
}

.video-container {
  margin-top: 16px;
  border-radius: 8px;
  overflow: hidden;
  width: fit-content;
  max-width: 100%;
  margin-left: auto;
  margin-right: auto;
}

.video-player {
  width: 100%;
  height: auto;
  max-height: 50vh;
  background: #000;
  display: block;
}
</style>
