<template>
  <div>
    <label v-if="label" class="block text-[13px] leading-4 font-medium text-[var(--text-body)] mb-1.5">
      {{ label }}
      <span v-if="required" class="text-[var(--danger)]">*</span>
    </label>
    <div class="relative">
      <component
        :is="icon"
        v-if="icon"
        class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-meta)] pointer-events-none"
      />
      <input
        :value="modelValue"
        :type="type"
        :placeholder="placeholder"
        :disabled="disabled"
        :autocomplete="autocomplete"
        class="w-full h-10 bg-[var(--bg-surface)] border rounded-[var(--r-md)] text-[14px] text-[var(--text-strong)] placeholder:text-[var(--text-meta)] outline-none transition-shadow disabled:opacity-50 disabled:cursor-not-allowed"
        :class="[
          icon ? 'pl-9 pr-3.5' : 'px-3.5',
          error
            ? 'border-[var(--danger)] focus:ring-2 focus:ring-[var(--danger)]/30'
            : 'border-[var(--border)] focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40',
        ]"
        @input="$emit('update:modelValue', $event.target.value)"
      />
    </div>
    <p v-if="error" class="text-[12px] text-[var(--danger)] mt-1">{{ error }}</p>
    <p v-else-if="hint" class="text-[12px] text-[var(--text-meta)] mt-1">{{ hint }}</p>
  </div>
</template>

<script setup>
defineProps({
  modelValue: { type: [String, Number], default: '' },
  label: { type: String, default: '' },
  type: { type: String, default: 'text' },
  placeholder: { type: String, default: '' },
  hint: { type: String, default: '' },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  icon: { type: [Object, Function], default: null },
  autocomplete: { type: String, default: 'off' },
});
defineEmits(['update:modelValue']);
</script>
