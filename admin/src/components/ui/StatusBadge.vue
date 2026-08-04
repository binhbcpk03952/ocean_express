<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-sm px-2.5 py-1 text-xs font-medium whitespace-nowrap"
    :style="{ color: cfg.fg, background: cfg.bg }"
  >
    <span class="h-1.5 w-1.5 rounded-full" :style="{ background: cfg.fg }"></span>
    {{ cfg.label }}
  </span>
</template>

<script setup>
import { computed } from 'vue';
import { statusConfig } from '../../composables/useStatus';

const props = defineProps({
  status: { type: String, required: true },
});

const cfg = computed(() => statusConfig(props.status));
</script>
