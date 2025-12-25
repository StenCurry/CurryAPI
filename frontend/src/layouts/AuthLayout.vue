<script setup lang="ts"></script>

<template>
  <div class="auth-layout">
    <div class="hero">
      <div>
        <p class="hero-tag">Curry2API · 控制台</p>
        <h1>统一管理你的 API 能力</h1>
        <p class="hero-desc">
          简洁 UI · 一站式令牌管理 · 实时看板，让 AI 接入体验更顺滑。
        </p>
      </div>
    </div>
    <div class="content">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  background: linear-gradient(135deg, #f5f7fa 0%, #e8edf5 100%);
  padding: 48px;
  gap: 32px;
}

.hero {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 24px;
  padding: 48px;
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.08);
  display: flex;
  align-items: flex-end;
}

.hero-tag {
  font-size: 14px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: #5c6b8a;
}

.hero h1 {
  margin: 8px 0;
  font-size: 40px;
  color: #111827;
}

.hero-desc {
  margin: 0;
  color: #5c667a;
  font-size: 16px;
  line-height: 1.8;
}

.content {
  display: flex;
  align-items: center;
  justify-content: center;
}

@media (max-width: 960px) {
  .auth-layout {
    padding: 24px;
  }

  .hero {
    padding: 32px;
  }
}
</style>
