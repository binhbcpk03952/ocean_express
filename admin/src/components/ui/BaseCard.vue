<template>
  <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--r-lg)] shadow-e1">
    <div v-if="$slots.header || title" class="flex items-center justify-between px-6 py-4 border-b border-[var(--border)]">
      <div>
        <h2 v-if="title" class="text-[18px] leading-7 font-semibold text-[var(--text-strong)]">{{ title }}</h2>
        <p v-if="subtitle" class="text-[13px] text-[var(--text-meta)] mt-0.5">{{ subtitle }}</p>
        <slot name="header" />
      </div>
      <slot name="actions" />
    </div>
    <div :class="bodyClass">
      <slot />
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
  bodyClass: { type: String, default: 'p-6' },
});
</script>
