<template>
  <div class="space-y-3">
    <!-- Tỉnh / Thành phố -->
    <FormSelect 
      v-model="selectedProvince" 
      :label="label ? `${label} (Tỉnh/Thành)` : 'Tỉnh/Thành phố'" 
      :required="required" 
      @change="onProvinceChange"
    >
      <option value="">— Chọn Tỉnh/Thành —</option>
      <option v-for="p in provinces" :key="p.id" :value="p.id">{{ p.name }}</option>
    </FormSelect>

    <!-- Quận / Huyện / Phường / Xã -->
    <FormSelect 
      v-model="selectedChild" 
      label="Quận/Huyện/Phường/Xã" 
      :required="required" 
      :disabled="!selectedProvince" 
      :hint="hint" 
      @change="onChildChange"
    >
      <option value="">— Chọn Khu vực —</option>
      <option v-for="c in children" :key="c.id" :value="c.id">{{ c.name }}</option>
    </FormSelect>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import api from '../../services/api';
import FormSelect from './FormSelect.vue';

const props = defineProps({
  modelValue: { type: String, default: '' },
  label: { type: String, default: 'Khu vực' },
  required: { type: Boolean, default: false },
  hint: { type: String, default: '' },
});

const emit = defineEmits(['update:modelValue', 'change']);

const allLocations = ref([]);
const loading = ref(false);

const selectedProvince = ref('');
const selectedChild = ref('');

// Computed lists
const provinces = computed(() => {
  return allLocations.value.filter(l => !l.parent_id);
});

const children = computed(() => {
  if (!selectedProvince.value) return [];
  return allLocations.value.filter(l => l.parent_id === selectedProvince.value);
});

// Setup Initial Value logic if modelValue exists
const applyModelValue = () => {
  if (!props.modelValue || allLocations.value.length === 0) return;
  
  const loc = allLocations.value.find(l => l.id === props.modelValue);
  if (!loc) return;

  if (!loc.parent_id) {
    // Là Tỉnh
    selectedProvince.value = loc.id;
    selectedChild.value = '';
  } else {
    // Là Cấp con (Xã/Phường/Quận/Huyện)
    selectedProvince.value = loc.parent_id;
    selectedChild.value = loc.id;
  }
};

const fetchAll = async () => {
  loading.value = true;
  try {
    const res = await api.get('/locations');
    if (res.success && res.data) {
      // Sắp xếp theo tên để dễ tìm
      allLocations.value = res.data.sort((a, b) => a.name.localeCompare(b.name));
      applyModelValue();
    }
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchAll);

watch(() => props.modelValue, (newVal) => {
  if (newVal !== selectedChild.value && newVal !== selectedProvince.value) {
    applyModelValue();
  }
});

const onProvinceChange = () => {
  selectedChild.value = '';
  emitValue(selectedProvince.value);
};

const onChildChange = () => {
  emitValue(selectedChild.value);
};

const emitValue = (val) => {
  emit('update:modelValue', val);
  emit('change', val);
};
</script>
