<template>
  <div>
    <label v-if="label" class="block text-[13px] leading-4 font-medium text-[var(--text-body)] mb-1.5">
      {{ label }}
      <span v-if="required" class="text-[var(--danger)]">*</span>
    </label>
    <div class="relative">
      <select
        :value="modelValue"
        :disabled="disabled"
        class="w-full h-10 pl-3.5 pr-10 bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--r-md)] text-[14px] text-[var(--text-strong)] outline-none appearance-none cursor-pointer transition-shadow focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40 disabled:opacity-50"
        @change="$emit('update:modelValue', $event.target.value)"
      >
        <slot />
      </select>
      <ChevronDown :size="16" class="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--text-meta)] pointer-events-none" />
    </div>
    <p v-if="hint" class="text-[12px] text-[var(--text-meta)] mt-1">{{ hint }}</p>
  </div>
</template>

<script setup>
import { ChevronDown } from 'lucide-vue-next';
defineProps({
  modelValue: { type: [String, Number], default: '' },
  label: { type: String, default: '' },
  hint: { type: String, default: '' },
  required: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
});
defineEmits(['update:modelValue']);
</script>
