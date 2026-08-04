<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="[
      'inline-flex items-center justify-center gap-2 font-medium whitespace-nowrap rounded-md transition-all duration-150 outline-none focus-visible:ring-2 focus-visible:ring-[var(--primary-ring)] focus-visible:ring-offset-1 focus-visible:ring-offset-[var(--bg-surface)] disabled:opacity-50 disabled:pointer-events-none',
      sizeCls,
      variantCls,
    ]"
  >
    <svg v-if="loading" class="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
    </svg>
    <slot v-else name="icon" />
    <slot />
  </button>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  variant: { type: String, default: 'primary' }, // primary | secondary | ghost | danger
  size: { type: String, default: 'md' },          // sm | md
  type: { type: String, default: 'button' },
  disabled: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
});

const sizeCls = computed(() => ({
  sm: 'h-9 px-3 text-[13px]',
  md: 'h-10 px-4 text-sm',
  lg: 'h-12 px-6 text-base',
}[props.size] || 'h-10 px-4 text-sm'));

const variantCls = computed(() => ({
  primary: 'bg-[var(--primary)] text-[var(--on-primary)] hover:bg-[var(--primary-hover)] shadow-e1',
  secondary: 'bg-[var(--bg-surface)] text-[var(--text-body)] border border-[var(--border-strong)] hover:bg-[var(--bg-subtle)]',
  ghost: 'text-[var(--text-body)] hover:bg-[var(--bg-subtle)]',
  danger: 'bg-[var(--danger)] text-white hover:brightness-95 shadow-e1',
}[props.variant]));
</script>
