<template>
  <Teleport to="body">
    <Transition name="oe-modal">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        @mousedown.self="closeOnBackdrop && $emit('update:modelValue', false)"
      >
        <div class="absolute inset-0 bg-[rgba(6,24,46,0.55)] backdrop-blur-sm"></div>
        <div
          class="relative bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--r-lg)] shadow-e3 w-full animate-scale-in"
          :class="sizeClass"
        >
          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-[var(--border)]">
            <div>
              <h3 class="text-[18px] leading-7 font-semibold text-[var(--text-strong)]">{{ title }}</h3>
              <p v-if="subtitle" class="text-[12px] text-[var(--text-meta)] mt-0.5">{{ subtitle }}</p>
            </div>
            <button
              class="w-8 h-8 flex items-center justify-center rounded-[var(--r-sm)] text-[var(--text-meta)] hover:bg-[var(--bg-subtle)] hover:text-[var(--text-strong)] transition-colors"
              @click="$emit('update:modelValue', false)"
            >
              <X :size="18" />
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-5 max-h-[70vh] overflow-y-auto">
            <slot />
          </div>

          <!-- Footer -->
          <div v-if="$slots.footer" class="flex items-center justify-end gap-3 px-6 py-4 border-t border-[var(--border)] bg-[var(--bg-subtle)] rounded-b-[var(--r-lg)]">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue';
import { X } from 'lucide-vue-next';

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
  size: { type: String, default: 'md' }, // sm | md | lg
  closeOnBackdrop: { type: Boolean, default: true },
});

defineEmits(['update:modelValue']);

const sizeClass = computed(() => ({
  sm: 'max-w-md',
  md: 'max-w-lg',
  lg: 'max-w-2xl',
}[props.size]));
</script>

<style scoped>
.oe-modal-enter-active,
.oe-modal-leave-active { transition: opacity 0.2s ease; }
.oe-modal-enter-from,
.oe-modal-leave-to { opacity: 0; }
</style>
