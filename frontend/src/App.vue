<script setup lang="ts">
import { computed, watch } from 'vue'
import {
  NConfigProvider,
  NDialogProvider,
  NLoadingBarProvider,
  NMessageProvider,
  NNotificationProvider,
  darkTheme,
  dateZhCN,
  zhCN,
  type GlobalThemeOverrides,
} from 'naive-ui'
import { useDark } from '@vueuse/core'

const isDark = useDark()

const theme = computed(() => (isDark.value ? darkTheme : null))

// 监听主题变化，更新 body class
watch(isDark, (dark) => {
  if (dark) {
    document.body.classList.add('dark-theme')
    document.body.classList.remove('light-theme')
  } else {
    document.body.classList.add('light-theme')
    document.body.classList.remove('dark-theme')
  }
}, { immediate: true })

// 简约风格主题配置 - 匹配 style.css 中的 CSS 变量
const lightThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#3b82f6',
    primaryColorHover: '#2563eb',
    primaryColorPressed: '#1d4ed8',
    primaryColorSuppl: '#60a5fa',
    infoColor: '#06b6d4',
    successColor: '#10b981',
    warningColor: '#f59e0b',
    errorColor: '#ef4444',
    textColorBase: '#1e293b',
    textColor1: '#1e293b',
    textColor2: '#64748b',
    textColor3: '#94a3b8',
    textColorDisabled: '#cbd5e1',
    placeholderColor: '#94a3b8',
    placeholderColorDisabled: '#cbd5e1',
    borderColor: '#e2e8f0',
    dividerColor: '#e2e8f0',
    inputColor: '#ffffff',
    inputColorDisabled: '#f1f5f9',
    actionColor: '#f8fafc',
    hoverColor: '#f1f5f9',
    pressedColor: '#e2e8f0',
    tableColor: '#ffffff',
    cardColor: '#ffffff',
    modalColor: '#ffffff',
    bodyColor: '#f8fafc',
    popoverColor: '#ffffff',
    tagColor: 'rgba(59, 130, 246, 0.1)',
    avatarColor: 'rgba(59, 130, 246, 0.1)',
    invertedColor: '#1e293b',
    scrollbarColor: '#cbd5e1',
    scrollbarColorHover: '#94a3b8',
    borderRadius: '8px',
    borderRadiusSmall: '6px',
    fontFamily: "'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif",
    fontFamilyMono: "'JetBrains Mono', 'SF Mono', 'Monaco', 'Menlo', monospace",
    fontSizeMini: '12px',
    fontSizeTiny: '12px',
    fontSizeSmall: '14px',
    fontSizeMedium: '14px',
    fontSizeLarge: '15px',
    fontSizeHuge: '16px',
    lineHeight: '1.6',
    heightTiny: '28px',
    heightSmall: '34px',
    heightMedium: '40px',
    heightLarge: '46px',
    heightHuge: '52px',
    // 简化阴影 - 移除发光效果
    boxShadow1: '0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04)',
    boxShadow2: '0 4px 6px -1px rgba(0, 0, 0, 0.08), 0 2px 4px -1px rgba(0, 0, 0, 0.04)',
    boxShadow3: '0 10px 15px -3px rgba(0, 0, 0, 0.08), 0 4px 6px -2px rgba(0, 0, 0, 0.04)',
  },
  Button: {
    borderRadiusMedium: '8px',
    borderRadiusSmall: '6px',
    borderRadiusLarge: '10px',
    fontWeightStrong: '500',
  },
  Card: {
    borderRadius: '12px',
    color: '#ffffff',
    colorModal: '#ffffff',
    borderColor: '#e2e8f0',
    titleTextColor: '#1e293b',
    textColor: '#64748b',
    actionColor: '#f8fafc',
  },
  Input: {
    color: '#ffffff',
    colorFocus: '#ffffff',
    border: '1px solid #e2e8f0',
    borderHover: '1px solid #cbd5e1',
    borderFocus: '1px solid #3b82f6',
    borderRadius: '8px',
    // 简化聚焦阴影
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.1)',
    textColor: '#1e293b',
    placeholderColor: '#94a3b8',
    caretColor: '#3b82f6',
  },
  Tag: {
    borderRadius: '6px',
    fontWeightStrong: '500',
  },
  DataTable: {
    borderRadius: '12px',
    thColor: '#f8fafc',
    thTextColor: '#1e293b',
    tdColor: '#ffffff',
    tdColorHover: '#f8fafc',
    tdTextColor: '#64748b',
    borderColor: '#e2e8f0',
  },
  Dialog: {
    color: '#ffffff',
    textColor: '#64748b',
    borderRadius: '12px',
    titleTextColor: '#1e293b',
  },
  Modal: {
    color: '#ffffff',
    textColor: '#64748b',
    borderRadius: '12px',
  },
  Menu: {
    borderRadius: '8px',
    itemTextColor: '#64748b',
    itemTextColorHover: '#1e293b',
    itemTextColorActive: '#ffffff',
    itemTextColorActiveHover: '#ffffff',
    itemTextColorChildActive: '#3b82f6',
    itemTextColorChildActiveHover: '#3b82f6',
    itemIconColor: '#94a3b8',
    itemIconColorHover: '#64748b',
    itemIconColorActive: '#ffffff',
    itemIconColorActiveHover: '#ffffff',
    itemIconColorChildActive: '#3b82f6',
    itemIconColorChildActiveHover: '#3b82f6',
    itemColorHover: '#f1f5f9',
    itemColorActive: '#3b82f6',
    itemColorActiveHover: '#2563eb',
    arrowColor: '#94a3b8',
    arrowColorHover: '#64748b',
    arrowColorActive: '#ffffff',
    arrowColorActiveHover: '#ffffff',
    arrowColorChildActive: '#3b82f6',
    arrowColorChildActiveHover: '#3b82f6',
    groupTextColor: '#94a3b8',
    color: 'transparent',
  },
  Tabs: {
    tabTextColorLine: '#64748b',
    tabTextColorActiveLine: '#1e293b',
    tabTextColorHoverLine: '#1e293b',
    tabTextColorDisabledLine: '#cbd5e1',
    barColor: '#3b82f6',
    tabColorSegment: '#f1f5f9',
    tabBorderColor: '#e2e8f0',
    tabTextColorSegment: '#64748b',
    tabTextColorActiveSegment: '#1e293b',
    tabTextColorHoverSegment: '#1e293b',
    tabTextColorDisabledSegment: '#cbd5e1',
    tabFontWeightActive: '500',
    colorSegment: '#f8fafc',
    tabBorderRadius: '8px',
    paneTextColor: '#64748b',
  },
  Dropdown: {
    optionColorHover: '#f1f5f9',
    optionTextColor: '#64748b',
    optionTextColorHover: '#1e293b',
    optionTextColorActive: '#3b82f6',
    color: '#ffffff',
    borderRadius: '8px',
    dividerColor: '#e2e8f0',
  },
  Select: {
    peers: {
      InternalSelection: {
        textColor: '#1e293b',
        placeholderColor: '#94a3b8',
        borderRadius: '8px',
        color: '#ffffff',
        border: '1px solid #e2e8f0',
        borderHover: '1px solid #cbd5e1',
        borderActive: '1px solid #3b82f6',
        borderFocus: '1px solid #3b82f6',
        boxShadowActive: '0 0 0 2px rgba(59, 130, 246, 0.1)',
        boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.1)',
      },
      InternalSelectMenu: {
        optionTextColor: '#64748b',
        optionTextColorActive: '#3b82f6',
        optionColorPending: '#f1f5f9',
        optionColorActive: 'rgba(59, 130, 246, 0.1)',
        color: '#ffffff',
        borderRadius: '8px',
      },
    },
  },
  Popover: {
    color: '#ffffff',
    textColor: '#64748b',
    borderRadius: '8px',
    spaceArrow: '10px',
    arrowHeight: '8px',
    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.08), 0 2px 4px -1px rgba(0, 0, 0, 0.04)',
  },
  Alert: {
    borderRadius: '8px',
  },
  Message: {
    borderRadius: '8px',
    color: '#ffffff',
    textColor: '#1e293b',
    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.08), 0 2px 4px -1px rgba(0, 0, 0, 0.04)',
  },
  Notification: {
    borderRadius: '12px',
    color: '#ffffff',
    textColor: '#64748b',
    titleTextColor: '#1e293b',
    boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.08), 0 4px 6px -2px rgba(0, 0, 0, 0.04)',
  },
  Progress: {
    fillColor: '#3b82f6',
    railColor: '#e2e8f0',
    fontSizeCircle: '24px',
    fontWeightCircle: '600',
    textColorCircle: '#1e293b',
    railHeight: '8px',
  },
  Form: {
    labelTextColor: '#64748b',
    asteriskColor: '#ef4444',
    feedbackTextColorError: '#ef4444',
    feedbackTextColorWarning: '#f59e0b',
  },
  Descriptions: {
    thColor: '#f8fafc',
    tdColor: '#ffffff',
    thTextColor: '#64748b',
    tdTextColor: '#1e293b',
    thFontWeight: '500',
    borderColor: '#e2e8f0',
    borderRadius: '8px',
  },
  Pagination: {
    itemTextColor: '#64748b',
    itemTextColorHover: '#1e293b',
    itemTextColorActive: '#ffffff',
    itemTextColorPressed: '#ffffff',
    itemColorHover: '#f1f5f9',
    itemColorActive: '#3b82f6',
    itemColorActiveHover: '#2563eb',
    itemColorPressed: '#1d4ed8',
    itemColor: '#ffffff',
    itemBorder: '1px solid #e2e8f0',
    itemBorderHover: '1px solid #cbd5e1',
    itemBorderActive: '1px solid #3b82f6',
    itemBorderPressed: '1px solid #1d4ed8',
    itemBorderRadius: '6px',
  },
  Spin: {
    color: '#3b82f6',
  },
  Avatar: {
    color: 'rgba(59, 130, 246, 0.1)',
    borderRadius: '50%',
  },
  Badge: {
    color: '#ef4444',
  },
  Checkbox: {
    colorChecked: '#3b82f6',
    borderChecked: '1px solid #3b82f6',
    borderFocus: '1px solid #3b82f6',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.1)',
  },
  Radio: {
    buttonColorActive: '#3b82f6',
    buttonTextColorActive: '#ffffff',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.1)',
  },
  Switch: {
    railColorActive: '#3b82f6',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.1)',
  },
  DatePicker: {
    panelColor: '#ffffff',
    calendarTitleTextColor: '#1e293b',
    arrowColor: '#64748b',
    itemColorHover: '#f1f5f9',
    itemColorActive: '#3b82f6',
    itemTextColor: '#64748b',
    itemTextColorActive: '#ffffff',
    calendarDaysDividerColor: '#e2e8f0',
    panelHeaderDividerColor: '#e2e8f0',
    panelBorderRadius: '12px',
    calendarTitleFontWeight: '500',
    itemBorderRadius: '6px',
  },
  Empty: {
    textColor: '#94a3b8',
    iconColor: '#cbd5e1',
  },
  Tooltip: {
    color: '#1e293b',
    textColor: '#ffffff',
    borderRadius: '6px',
  },
  LoadingBar: {
    colorLoading: '#3b82f6',
  },
}

// 暗色主题配置
const darkThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#3b82f6',
    primaryColorHover: '#60a5fa',
    primaryColorPressed: '#2563eb',
    primaryColorSuppl: '#93c5fd',
    infoColor: '#06b6d4',
    successColor: '#10b981',
    warningColor: '#f59e0b',
    errorColor: '#ef4444',
    textColorBase: '#f1f5f9',
    textColor1: '#f1f5f9',
    textColor2: '#94a3b8',
    textColor3: '#64748b',
    textColorDisabled: '#475569',
    placeholderColor: '#64748b',
    placeholderColorDisabled: '#475569',
    borderColor: '#334155',
    dividerColor: '#334155',
    inputColor: '#1e293b',
    inputColorDisabled: '#0f172a',
    actionColor: '#1e293b',
    hoverColor: '#334155',
    pressedColor: '#475569',
    tableColor: '#1e293b',
    cardColor: '#1e293b',
    modalColor: '#1e293b',
    bodyColor: '#0f172a',
    popoverColor: '#1e293b',
    tagColor: 'rgba(59, 130, 246, 0.2)',
    avatarColor: 'rgba(59, 130, 246, 0.2)',
    invertedColor: '#f1f5f9',
    scrollbarColor: '#475569',
    scrollbarColorHover: '#64748b',
    borderRadius: '8px',
    borderRadiusSmall: '6px',
    fontFamily: "'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif",
    fontFamilyMono: "'JetBrains Mono', 'SF Mono', 'Monaco', 'Menlo', monospace",
    fontSizeMini: '12px',
    fontSizeTiny: '12px',
    fontSizeSmall: '14px',
    fontSizeMedium: '14px',
    fontSizeLarge: '15px',
    fontSizeHuge: '16px',
    lineHeight: '1.6',
    heightTiny: '28px',
    heightSmall: '34px',
    heightMedium: '40px',
    heightLarge: '46px',
    heightHuge: '52px',
    // 简化阴影 - 暗色主题
    boxShadow1: '0 1px 3px rgba(0, 0, 0, 0.3), 0 1px 2px rgba(0, 0, 0, 0.2)',
    boxShadow2: '0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -1px rgba(0, 0, 0, 0.2)',
    boxShadow3: '0 10px 15px -3px rgba(0, 0, 0, 0.3), 0 4px 6px -2px rgba(0, 0, 0, 0.2)',
  },
  Button: {
    borderRadiusMedium: '8px',
    borderRadiusSmall: '6px',
    borderRadiusLarge: '10px',
    fontWeightStrong: '500',
  },
  Card: {
    borderRadius: '12px',
    color: '#1e293b',
    colorModal: '#1e293b',
    borderColor: '#334155',
    titleTextColor: '#f1f5f9',
    textColor: '#94a3b8',
    actionColor: '#0f172a',
  },
  Input: {
    color: '#1e293b',
    colorFocus: '#1e293b',
    border: '1px solid #334155',
    borderHover: '1px solid #475569',
    borderFocus: '1px solid #3b82f6',
    borderRadius: '8px',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
    textColor: '#f1f5f9',
    placeholderColor: '#64748b',
    caretColor: '#3b82f6',
  },
  Tag: {
    borderRadius: '6px',
    fontWeightStrong: '500',
  },
  DataTable: {
    borderRadius: '12px',
    thColor: '#0f172a',
    thTextColor: '#f1f5f9',
    tdColor: '#1e293b',
    tdColorHover: '#334155',
    tdTextColor: '#94a3b8',
    borderColor: '#334155',
  },
  Dialog: {
    color: '#1e293b',
    textColor: '#94a3b8',
    borderRadius: '12px',
    titleTextColor: '#f1f5f9',
  },
  Modal: {
    color: '#1e293b',
    textColor: '#94a3b8',
    borderRadius: '12px',
  },
  Menu: {
    borderRadius: '8px',
    itemTextColor: '#94a3b8',
    itemTextColorHover: '#f1f5f9',
    itemTextColorActive: '#ffffff',
    itemTextColorActiveHover: '#ffffff',
    itemTextColorChildActive: '#60a5fa',
    itemTextColorChildActiveHover: '#60a5fa',
    itemIconColor: '#64748b',
    itemIconColorHover: '#94a3b8',
    itemIconColorActive: '#ffffff',
    itemIconColorActiveHover: '#ffffff',
    itemIconColorChildActive: '#60a5fa',
    itemIconColorChildActiveHover: '#60a5fa',
    itemColorHover: '#334155',
    itemColorActive: '#3b82f6',
    itemColorActiveHover: '#2563eb',
    arrowColor: '#64748b',
    arrowColorHover: '#94a3b8',
    arrowColorActive: '#ffffff',
    arrowColorActiveHover: '#ffffff',
    arrowColorChildActive: '#60a5fa',
    arrowColorChildActiveHover: '#60a5fa',
    groupTextColor: '#64748b',
    color: 'transparent',
  },
  Tabs: {
    tabTextColorLine: '#94a3b8',
    tabTextColorActiveLine: '#f1f5f9',
    tabTextColorHoverLine: '#f1f5f9',
    tabTextColorDisabledLine: '#475569',
    barColor: '#3b82f6',
    tabColorSegment: '#334155',
    tabBorderColor: '#334155',
    tabTextColorSegment: '#94a3b8',
    tabTextColorActiveSegment: '#f1f5f9',
    tabTextColorHoverSegment: '#f1f5f9',
    tabTextColorDisabledSegment: '#475569',
    tabFontWeightActive: '500',
    colorSegment: '#1e293b',
    tabBorderRadius: '8px',
    paneTextColor: '#94a3b8',
  },
  Dropdown: {
    optionColorHover: '#334155',
    optionTextColor: '#94a3b8',
    optionTextColorHover: '#f1f5f9',
    optionTextColorActive: '#60a5fa',
    color: '#1e293b',
    borderRadius: '8px',
    dividerColor: '#334155',
  },
  Select: {
    peers: {
      InternalSelection: {
        textColor: '#f1f5f9',
        placeholderColor: '#64748b',
        borderRadius: '8px',
        color: '#1e293b',
        border: '1px solid #334155',
        borderHover: '1px solid #475569',
        borderActive: '1px solid #3b82f6',
        borderFocus: '1px solid #3b82f6',
        boxShadowActive: '0 0 0 2px rgba(59, 130, 246, 0.2)',
        boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
      },
      InternalSelectMenu: {
        optionTextColor: '#94a3b8',
        optionTextColorActive: '#60a5fa',
        optionColorPending: '#334155',
        optionColorActive: 'rgba(59, 130, 246, 0.2)',
        color: '#1e293b',
        borderRadius: '8px',
      },
    },
  },
  Popover: {
    color: '#1e293b',
    textColor: '#94a3b8',
    borderRadius: '8px',
    spaceArrow: '10px',
    arrowHeight: '8px',
    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -1px rgba(0, 0, 0, 0.2)',
  },
  Alert: {
    borderRadius: '8px',
  },
  Message: {
    borderRadius: '8px',
    color: '#1e293b',
    textColor: '#f1f5f9',
    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -1px rgba(0, 0, 0, 0.2)',
  },
  Notification: {
    borderRadius: '12px',
    color: '#1e293b',
    textColor: '#94a3b8',
    titleTextColor: '#f1f5f9',
    boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.3), 0 4px 6px -2px rgba(0, 0, 0, 0.2)',
  },
  Progress: {
    fillColor: '#3b82f6',
    railColor: '#334155',
    fontSizeCircle: '24px',
    fontWeightCircle: '600',
    textColorCircle: '#f1f5f9',
    railHeight: '8px',
  },
  Form: {
    labelTextColor: '#94a3b8',
    asteriskColor: '#ef4444',
    feedbackTextColorError: '#ef4444',
    feedbackTextColorWarning: '#f59e0b',
  },
  Descriptions: {
    thColor: '#0f172a',
    tdColor: '#1e293b',
    thTextColor: '#94a3b8',
    tdTextColor: '#f1f5f9',
    thFontWeight: '500',
    borderColor: '#334155',
    borderRadius: '8px',
  },
  Pagination: {
    itemTextColor: '#94a3b8',
    itemTextColorHover: '#f1f5f9',
    itemTextColorActive: '#ffffff',
    itemTextColorPressed: '#ffffff',
    itemColorHover: '#334155',
    itemColorActive: '#3b82f6',
    itemColorActiveHover: '#2563eb',
    itemColorPressed: '#1d4ed8',
    itemColor: '#1e293b',
    itemBorder: '1px solid #334155',
    itemBorderHover: '1px solid #475569',
    itemBorderActive: '1px solid #3b82f6',
    itemBorderPressed: '1px solid #1d4ed8',
    itemBorderRadius: '6px',
  },
  Spin: {
    color: '#3b82f6',
  },
  Avatar: {
    color: 'rgba(59, 130, 246, 0.2)',
    borderRadius: '50%',
  },
  Badge: {
    color: '#ef4444',
  },
  Checkbox: {
    colorChecked: '#3b82f6',
    borderChecked: '1px solid #3b82f6',
    borderFocus: '1px solid #3b82f6',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
  },
  Radio: {
    buttonColorActive: '#3b82f6',
    buttonTextColorActive: '#ffffff',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
  },
  Switch: {
    railColorActive: '#3b82f6',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
  },
  DatePicker: {
    panelColor: '#1e293b',
    calendarTitleTextColor: '#f1f5f9',
    arrowColor: '#94a3b8',
    itemColorHover: '#334155',
    itemColorActive: '#3b82f6',
    itemTextColor: '#94a3b8',
    itemTextColorActive: '#ffffff',
    calendarDaysDividerColor: '#334155',
    panelHeaderDividerColor: '#334155',
    panelBorderRadius: '12px',
    calendarTitleFontWeight: '500',
    itemBorderRadius: '6px',
  },
  Empty: {
    textColor: '#64748b',
    iconColor: '#475569',
  },
  Tooltip: {
    color: '#f1f5f9',
    textColor: '#1e293b',
    borderRadius: '6px',
  },
  LoadingBar: {
    colorLoading: '#3b82f6',
  },
}

// 根据主题选择配置
const themeOverrides = computed(() => isDark.value ? darkThemeOverrides : lightThemeOverrides)
</script>

<template>
  <n-config-provider
    :theme="theme"
    :theme-overrides="themeOverrides"
    :locale="zhCN"
    :date-locale="dateZhCN"
  >
    <n-loading-bar-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <n-message-provider>
            <router-view />
          </n-message-provider>
        </n-notification-provider>
      </n-dialog-provider>
    </n-loading-bar-provider>
  </n-config-provider>
</template>

<style>
/* 移除背景图片，使用简洁的渐变背景（由 style.css 控制） */
body {
  margin: 0;
  min-height: 100vh;
}

/* 确保内容区域透明，让背景显示 */
#app {
  min-height: 100vh;
  background: transparent !important;
}

/* 确保 naive-ui 布局组件透明 */
.n-layout,
.n-layout-content,
.n-layout-scroll-container,
.n-config-provider {
  background: transparent !important;
}
</style>
