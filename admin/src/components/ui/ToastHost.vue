<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-[100] flex flex-col gap-2 w-[340px] max-w-[calc(100vw-2rem)]">
      <TransitionGroup name="toast">
        <div
          v-for="t in toast.toasts"
          :key="t.id"
          class="flex items-start gap-3 p-3.5 pr-3 rounded-[var(--r-md)] bg-[var(--bg-surface)] shadow-[var(--e2)] border-l-4"
          :style="{ borderLeftColor: colorOf(t.type) }"
        >
          <component :is="iconOf(t.type)" :size="18" :style="{ color: colorOf(t.type) }" class="shrink-0 mt-0.5" />
          <p class="flex-1 text-[13px] text-[var(--text-body)] leading-5">{{ t.message }}</p>
          <button class="text-[var(--text-meta)] hover:text-[var(--text-strong)] transition-colors" @click="toast.dismiss(t.id)">
            <X :size="16" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup>
import { CheckCircle2, AlertTriangle, XCircle, Info, X } from 'lucide-vue-next';
import { useToastStore } from '../../stores/toastStore';

const toast = useToastStore();

const colorOf = (type) => ({
  success: 'var(--success)',
  danger: 'var(--danger)',
  warning: 'var(--warning)',
  info: 'var(--info)',
}[type] || 'var(--info)');

const iconOf = (type) => ({
  success: CheckCircle2,
  danger: XCircle,
  warning: AlertTriangle,
  info: Info,
}[type] || Info);
</script>

<style scoped>
.toast-enter-active { transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1); }
.toast-leave-active { transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1); }
.toast-enter-from { opacity: 0; transform: translateX(20px); }
.toast-leave-to { opacity: 0; transform: translateX(20px); }
</style>
