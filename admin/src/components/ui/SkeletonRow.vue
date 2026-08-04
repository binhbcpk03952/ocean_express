<template>
  <div class="flex items-center gap-4 py-3.5 px-4">
    <div v-for="(w, i) in widths" :key="i" class="skeleton h-4" :style="{ width: w }"></div>
  </div>
</template>

<script setup>
defineProps({
  widths: { type: Array, default: () => ['20%', '25%', '15%', '20%', '15%'] },
});
</script>
