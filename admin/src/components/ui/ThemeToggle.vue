<template>
  <button
    class="w-9 h-9 flex items-center justify-center rounded-[var(--r-md)] text-[var(--text-meta)] hover:bg-[var(--bg-subtle)] hover:text-[var(--text-strong)] transition-colors"
    :title="theme.isDark ? 'Chuyển sáng' : 'Chuyển tối'"
    @click="theme.toggle()"
  >
    <Moon v-if="!theme.isDark" :size="18" />
    <Sun v-else :size="18" />
  </button>
</template>

<script setup>
import { Moon, Sun } from 'lucide-vue-next';
import { useThemeStore } from '../../stores/themeStore';
const theme = useThemeStore();
</script>
