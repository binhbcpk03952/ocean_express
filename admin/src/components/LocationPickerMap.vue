<template>
  <div class="h-[400px] w-full rounded-md overflow-hidden relative border border-input">
    <div class="absolute top-2 left-2 right-2 z-[400] bg-white p-2 rounded shadow flex items-center gap-2">
      <Search class="w-4 h-4 text-meta shrink-0" />
      <input 
        v-model="searchQuery" 
        @keydown.enter.prevent="searchLocation"
        type="text" 
        placeholder="Tìm kiếm địa chỉ (vd: 300 Hà Huy Tập)..."
        class="w-full text-sm outline-none border-none bg-transparent focus:ring-0"
      />
      <BaseButton variant="secondary" size="sm" :loading="isSearching" @click.prevent="searchLocation">Tìm</BaseButton>
    </div>

    <l-map ref="mapRef" :zoom="zoom" :center="center" @click="onMapClick" :useGlobalLeaflet="false">
      <l-tile-layer
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        layer-type="base"
        name="OpenStreetMap"
      />
      <l-marker v-if="selectedLocation" :lat-lng="selectedLocation" draggable @dragend="onMarkerDrag" />
    </l-map>

    <div class="absolute bottom-2 left-2 z-[400]">
      <BaseButton variant="secondary" size="sm" @click.prevent="locateMe">
        <MapPin class="w-4 h-4 mr-1" /> Vị trí hiện tại
      </BaseButton>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue';
import 'leaflet/dist/leaflet.css';
import { LMap, LTileLayer, LMarker } from '@vue-leaflet/vue-leaflet';
import { Search, MapPin } from 'lucide-vue-next';
import BaseButton from './ui/BaseButton.vue';
import { useToastStore as useToast } from '../stores/toastStore';
import axios from 'axios';

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => null // { lat, lng }
  },
  defaultCenter: {
    type: Object,
    default: () => ({ lat: 14.0583, lng: 108.2772 }) // Trung tâm Việt Nam
  }
});

const emit = defineEmits(['update:modelValue']);
const toast = useToast();

const mapRef = ref(null);
const zoom = ref(5);
const center = ref([props.defaultCenter.lat, props.defaultCenter.lng]);
const selectedLocation = ref(null);

const searchQuery = ref('');
const isSearching = ref(false);

onMounted(() => {
  if (props.modelValue && props.modelValue.lat && props.modelValue.lng) {
    selectedLocation.value = [props.modelValue.lat, props.modelValue.lng];
    center.value = [props.modelValue.lat, props.modelValue.lng];
    zoom.value = 15;
  }
});

watch(() => props.modelValue, (newVal) => {
  if (newVal && newVal.lat && newVal.lng) {
    selectedLocation.value = [newVal.lat, newVal.lng];
    center.value = [newVal.lat, newVal.lng];
  } else {
    selectedLocation.value = null;
  }
}, { deep: true });

const onMapClick = (e) => {
  const { lat, lng } = e.latlng;
  updateLocation(lat, lng);
};

const onMarkerDrag = (e) => {
  const { lat, lng } = e.target.getLatLng();
  updateLocation(lat, lng);
};

const updateLocation = (lat, lng) => {
  selectedLocation.value = [lat, lng];
  emit('update:modelValue', { lat, lng });
};

const locateMe = () => {
  if (!navigator.geolocation) {
    toast.error('Trình duyệt của bạn không hỗ trợ định vị');
    return;
  }
  
  navigator.geolocation.getCurrentPosition(
    (position) => {
      const lat = position.coords.latitude;
      const lng = position.coords.longitude;
      updateLocation(lat, lng);
      center.value = [lat, lng];
      zoom.value = 16;
    },
    (error) => {
      toast.error('Không thể lấy vị trí: ' + error.message);
    }
  );
};

const searchLocation = async () => {
  if (!searchQuery.value) return;
  isSearching.value = true;
  try {
    const q = encodeURIComponent(searchQuery.value + ', Việt Nam');
    const res = await axios.get(`https://nominatim.openstreetmap.org/search?q=${q}&format=json&limit=1&countrycodes=vn`);
    if (res.data && res.data.length > 0) {
      const lat = parseFloat(res.data[0].lat);
      const lng = parseFloat(res.data[0].lon);
      updateLocation(lat, lng);
      center.value = [lat, lng];
      zoom.value = 16;
    } else {
      toast.warning('Không tìm thấy địa chỉ này');
    }
  } catch (error) {
    toast.error('Lỗi khi tìm kiếm địa chỉ');
  } finally {
    isSearching.value = false;
  }
};
</script>
