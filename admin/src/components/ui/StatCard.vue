<template>
  <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--r-lg)] shadow-e1 p-5 flex items-center justify-between hover:shadow-e2 transition-shadow duration-200">
    <div class="min-w-0">
      <p class="text-[12px] leading-4 font-medium text-[var(--text-meta)] uppercase tracking-wide">{{ label }}</p>
      <p class="text-[30px] leading-9 font-bold text-[var(--text-strong)] mt-2 tabular">
        <span v-if="loading" class="skeleton inline-block w-20 h-8 rounded"></span>
        <template v-else>{{ formattedValue }}</template>
      </p>
      <p v-if="hint" class="text-[12px] text-[var(--text-meta)] mt-1">{{ hint }}</p>
    </div>
    <div
      class="w-12 h-12 rounded-full flex items-center justify-center shrink-0"
      :style="{ backgroundColor: tintBg, color: tint }"
    >
      <component :is="icon" :size="22" :stroke-width="2" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  label: { type: String, required: true },
  value: { type: [Number, String], default: 0 },
  icon: { type: [Object, Function], default: null },
  hint: { type: String, default: '' },
  tint: { type: String, default: 'var(--primary)' },
  tintBg: { type: String, default: 'var(--primary-soft)' },
  loading: { type: Boolean, default: false },
  format: { type: String, default: 'number' }, // number | plain
});

const formattedValue = computed(() => {
  if (props.format === 'number' && typeof props.value === 'number') {
    return new Intl.NumberFormat('vi-VN').format(props.value);
  }
  return props.value;
});
</script>
