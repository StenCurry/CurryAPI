<template>
  <div class="documentation">
    <!-- 动态背景 -->
    <div class="animated-background">
      <div class="gradient-orb orb-1"></div>
      <div class="gradient-orb orb-2"></div>
    </div>

    <div class="doc-card glass-card">
      <n-space vertical align="center" :size="40">
        <div class="icon-wrapper">
          <n-icon size="120">
            <BookmarkOutline />
          </n-icon>
        </div>
        <h1 class="main-title gradient-text">
          不想写文档
        </h1>
        <h2 class="subtitle">
          不会自己上网去学，很多教程
        </h2>
        <div class="divider"></div>
        <n-button type="primary" size="large" @click="router.push('/dashboard')" class="back-btn">
          返回首页
        </n-button>
      </n-space>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { BookmarkOutline } from '@vicons/ionicons5'

const router = useRouter()
</script>

<style scoped>
.documentation {
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 2rem;
  overflow: hidden;
}

/* 动态背景 - 红色警告主题 */
.animated-background {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: -1;
  overflow: hidden;
  background: #1a0a0a;
}

.animated-background::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    45deg,
    #1a0a0a 0%,
    #2a0f0f 25%,
    #1f0a0a 50%,
    #2d1010 75%,
    #1a0a0a 100%
  );
  background-size: 400% 400%;
  animation: gradientShift 15s ease infinite;
}

@keyframes gradientShift {
  0%, 100% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
}

.gradient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
  animation: float 20s infinite ease-in-out;
}

.orb-1 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(239, 68, 68, 0.6) 0%, transparent 70%);
  top: -10%;
  left: -10%;
  animation-delay: 0s;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(248, 113, 113, 0.5) 0%, transparent 70%);
  bottom: -10%;
  right: -10%;
  animation-delay: 10s;
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  50% {
    transform: translate(30px, -30px) scale(1.1);
  }
}

/* Glassmorphism 卡片 */
.glass-card {
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 32px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.doc-card {
  max-width: 800px;
  width: 100%;
  padding: 4rem 3rem;
  text-align: center;
  animation: cardSlideIn 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  background: rgba(30, 10, 10, 0.92);
  border: 1px solid rgba(239, 68, 68, 0.4);
  box-shadow: 
    0 20px 60px rgba(0, 0, 0, 0.5),
    0 0 40px rgba(239, 68, 68, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

@keyframes cardSlideIn {
  from {
    opacity: 0;
    transform: translateY(50px) scale(0.9);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.doc-card:hover {
  transform: translateY(-8px);
  box-shadow: 
    0 24px 72px rgba(0, 0, 0, 0.6),
    0 0 60px rgba(239, 68, 68, 0.25),
    inset 0 1px 0 rgba(255, 255, 255, 0.15);
  border-color: rgba(239, 68, 68, 0.6);
}

.icon-wrapper {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 180px;
  height: 180px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  animation: iconPulse 2s infinite ease-in-out;
}

@keyframes iconPulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

.icon-wrapper :deep(.n-icon) {
  color: white;
  filter: drop-shadow(0 8px 16px rgba(0, 0, 0, 0.3));
}

.main-title {
  font-size: 3rem;
  font-weight: 800;
  margin: 0;
  letter-spacing: -1px;
}

.gradient-text {
  background: linear-gradient(135deg, #ffffff 0%, #fef2f2 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  text-shadow: 0 4px 12px rgba(255, 255, 255, 0.3);
}

.subtitle {
  color: rgba(255, 255, 255, 0.9);
  font-size: 1.5rem;
  font-weight: 500;
  margin: 0;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.divider {
  width: 200px;
  height: 2px;
  background: linear-gradient(90deg, transparent 0%, rgba(255, 255, 255, 0.5) 50%, transparent 100%);
  margin: 1rem 0;
}

.back-btn {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.3) 0%, rgba(255, 255, 255, 0.2) 100%) !important;
  border: 1px solid rgba(255, 255, 255, 0.4) !important;
  color: white !important;
  font-weight: 700;
  font-size: 1.1rem;
  padding: 0.75rem 2.5rem !important;
  border-radius: 12px !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.back-btn:hover {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.4) 0%, rgba(255, 255, 255, 0.3) 100%) !important;
  border-color: white !important;
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.25);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .documentation {
    padding: 1rem;
  }

  .doc-card {
    padding: 3rem 2rem;
  }

  .icon-wrapper {
    width: 140px;
    height: 140px;
  }

  .icon-wrapper :deep(.n-icon) {
    font-size: 80px !important;
  }
  
  .main-title {
    font-size: 2rem;
  }
  
  .subtitle {
    font-size: 1.2rem;
  }
}
</style>
