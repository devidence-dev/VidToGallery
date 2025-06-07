<script setup lang="ts">
import VideoInput from '@/components/VideoInput.vue'
import VideoPreview from '@/components/VideoPreview.vue'
import { useVideoStore } from '@/stores/video'

const videoStore = useVideoStore()
const { error } = videoStore
</script>

<template>
  <main class="home">
    <!-- Header with gradient background -->
    <header class="hero-header">
      <div class="hero-content">
        <div class="hero-icon">
          <van-icon name="video" size="48" color="#ffffff" />
        </div>
        <h1 class="hero-title">VidToGallery</h1>
        <p class="hero-subtitle">Download your videos directly to your iOS gallery</p>
      </div>
      <div class="hero-decoration"></div>
    </header>
    
    <!-- Main content with cards -->
    <div class="content">
      <div class="card-container">
        <div class="input-card">
          <div class="card-header">
            <van-icon name="plus" size="20" />
            <h3>Add Video</h3>
          </div>
          <VideoInput />
        </div>
        
        <van-notice-bar
          v-if="error"
          type="danger"
          :text="error"
          closeable
          @close="videoStore.error = ''"
          class="error-notice"
        />
        
        <div class="preview-card">
          <div class="card-header">
            <van-icon name="photo" size="20" />
            <h3>Preview</h3>
          </div>
          <VideoPreview />
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.home {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
}

/* Custom button colors to match purple gradient */
.home :deep(.van-button--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: #667eea;
}

.home :deep(.van-button--primary:hover) {
  background: linear-gradient(135deg, #5a6fd8 0%, #6a4190 100%);
  border-color: #5a6fd8;
}

.home :deep(.van-button--primary:active) {
  background: linear-gradient(135deg, #5566c6 0%, #5e3a7e 100%);
  border-color: #5566c6;
}

.home :deep(.van-button--primary.van-button--plain) {
  background: transparent;
  color: #667eea;
  border-color: #667eea;
}

.home :deep(.van-button--primary.van-button--plain:hover) {
  background: rgba(102, 126, 234, 0.1);
  color: #5a6fd8;
  border-color: #5a6fd8;
}

.home :deep(.van-radio__icon--checked .van-icon) {
  color: var(--van-white);
  background-color: #667eea;
  border-color: #667eea;
}

.home :deep(.van-tag--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.hero-header {
  position: relative;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 40px 20px 60px;
  text-align: center;
  overflow: hidden;
}

.hero-content {
  position: relative;
  z-index: 2;
  animation: fadeInUp 0.8s ease-out;
}

.hero-icon {
  margin-bottom: 16px;
  animation: pulse 2s infinite;
}

.hero-title {
  font-size: 32px;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 8px 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.hero-subtitle {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.9);
  margin: 0 0 16px 0;
  font-weight: 400;
}

.install-button {
  margin-top: 16px;
}

.install-button .van-button {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.3);
  color: white;
}

.install-button .van-button:hover {
  background: rgba(255, 255, 255, 0.3);
  border-color: rgba(255, 255, 255, 0.5);
}

.hero-decoration {
  position: absolute;
  top: -50%;
  right: -20%;
  width: 300px;
  height: 300px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  animation: float 6s ease-in-out infinite;
}

.hero-decoration::before {
  content: '';
  position: absolute;
  top: 50px;
  left: 50px;
  width: 150px;
  height: 150px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 50%;
  animation: float 4s ease-in-out infinite reverse;
}

.content {
  background: #f8f9fa;
  border-radius: 24px 24px 0 0;
  margin-top: -24px;
  position: relative;
  z-index: 1;
  min-height: calc(100vh - 200px);
  padding: 24px 16px;
  width: 100%;
}

.card-container {
  max-width: 600px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
}

.input-card,
.preview-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.2);
  transition: all 0.3s ease;
  animation: slideInUp 0.6s ease-out;
}

.input-card:hover,
.preview-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.12);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.card-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.card-header .van-icon {
  color: #667eea;
}

.error-notice {
  animation: slideInDown 0.3s ease-out;
}

/* Animations */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideInDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-20px);
  }
}

/* Responsive design */
@media (max-width: 768px) {
  .hero-header {
    padding: 30px 16px 50px;
  }
  
  .hero-title {
    font-size: 28px;
  }
  
  .hero-subtitle {
    font-size: 14px;
  }
  
  .content {
    padding: 20px 12px;
  }
  
  .input-card,
  .preview-card {
    padding: 16px;
  }
}

@media (min-width: 769px) {
  .home {
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  
  .hero-header {
    width: 100%;
    max-width: 100%;
  }
  
  .content {
    width: 100%;
    max-width: 1200px;
    padding: 40px 24px;
    border-radius: 24px;
    margin-top: -24px;
  }
  
  .card-container {
    max-width: 800px;
  }
  
  .input-card,
  .preview-card {
    padding: 32px;
  }
}

@media (min-width: 1024px) {
  .content {
    max-width: 1400px;
    padding: 48px 32px;
  }
  
  .card-container {
    max-width: 1000px;
  }
}
</style>
