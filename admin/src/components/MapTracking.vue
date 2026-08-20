<template>
  <div class="rounded-[var(--r-md)] overflow-hidden border relative" style="height: 320px; width: 100%; z-index: 0; background: #dbeafe;">
    <!-- Google Maps: nền chi tiết + đường đi bám đường thật (Directions API). -->
    <GoogleRouteMap v-if="useGoogleMap" :points="mapPoints" :api-key="googleKey" />

    <!-- Fallback tự chủ (GeoJSON) khi không cấu hình Google key. -->
    <div v-else class="relative h-full w-full">
      <l-map ref="map" v-model:zoom="zoom" :center="center" :bounds="routeBounds" :use-global-leaflet="true" :attribution-control="false">
        <l-tile-layer
          url="https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}&hl=vi"
          layer-type="base"
          name="Google Maps"
          :max-zoom="20"
        />

        <template v-if="detailedRoutePath.length > 1">
          <!-- Lớp 1: Viền trắng mỏng (Border) -->
          <l-polyline
            :lat-lngs="detailedRoutePath"
            :color="'#ffffff'"
            :weight="7"
            :opacity="1"
          />
          <!-- Lớp 2: Lõi xanh dương (Core) -->
          <l-polyline
            :lat-lngs="detailedRoutePath"
            :color="'#0055ff'"
            :weight="4"
            :opacity="1"
          />
        </template>
        <l-marker
          v-for="(log, idx) in routePoints"
          :key="log.id"
          :lat-lng="snappedPoints.length === routePoints.length ? snappedPoints[idx] : [log.latitude, log.longitude]"
          :icon="pointIcon(idx, log)"
        >
          <l-tooltip v-if="idx === currentIdx" :options="{ permanent: true, direction: 'top', className: 'premium-tooltip', offset: [0, -28] }">
            <div class="text-[11px] text-gray-300 font-normal mb-1">Hiện đang ở</div>
            <div class="text-[#06b6d4] font-medium text-[13px] leading-tight">{{ log.note || statusLabel(log.status) }}</div>
          </l-tooltip>
          <l-tooltip v-else>{{ pointBadge(idx) }} — {{ statusLabel(log.status) }} · {{ formatDate(log.created_at) }}</l-tooltip>
        </l-marker>
      </l-map>

      <!-- Panel tổng quan (đường chim bay / đường bộ) khi dùng fallback. -->
      <div v-if="routingDistance || fallbackSummary" class="fb-summary">
        <div class="fb-item">
          <span class="fb-label">Quãng đường</span>
          <span class="fb-value">{{ routingDistance || fallbackSummary }}</span>
        </div>
      </div>

      <!-- Legend chú thích. -->
      <div v-if="routePoints.length" class="fb-legend">
        <div class="fb-row"><span class="fb-dot fb-start"></span> Điểm xuất phát</div>
        <div class="fb-row"><span class="fb-dot fb-mid"></span> Điểm trung gian</div>
        <div class="fb-row"><Truck class="w-3 h-3 text-[var(--primary)]" style="margin: 0 1px;" /> Vị trí hiện tại</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue';
import { Truck } from 'lucide-vue-next';
import "leaflet/dist/leaflet.css";
import L from "leaflet";
import { LMap, LTileLayer, LMarker, LTooltip, LPolyline } from "@vue-leaflet/vue-leaflet";
import GoogleRouteMap from './ui/GoogleRouteMap.vue';
import { statusLabel } from '../composables/useStatus';

const props = defineProps({
  logs: {
    type: Array,
    required: true
  },
  order: {
    type: Object,
    required: false,
    default: null
  }
});

const map = ref(null);
const zoom = ref(5);
const center = ref([16.047079, 108.206230]);

const routePoints = computed(() => {
  const points = [];
  if (props.order && props.order.sender_latitude && props.order.sender_longitude) {
    points.push({
      id: 'sender',
      latitude: props.order.sender_latitude,
      longitude: props.order.sender_longitude,
      status: 'ready_to_pick',
      note: 'Điểm lấy hàng',
      created_at: props.order.created_at
    });
  }

  const validLogs = [...props.logs]
    .filter(l => l.latitude && l.longitude)
    .sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
  
  points.push(...validLogs);

  if (props.order && props.order.receiver_latitude && props.order.receiver_longitude) {
    points.push({
      id: 'receiver',
      latitude: props.order.receiver_latitude,
      longitude: props.order.receiver_longitude,
      status: 'delivered',
      note: 'Điểm giao hàng',
      created_at: new Date().toISOString()
    });
  }

  return points;
});

const logsWithGps = computed(() => routePoints.value);

// Xác định vị trí hiện tại (index của xe tải)
const currentIdx = computed(() => {
  if (routePoints.value.length === 0) return -1;
  // Nếu đã giao thành công, xe tải ở điểm cuối cùng
  if (props.order && props.order.status === 'delivered') {
    return routePoints.value.length - 1;
  }
  // Nếu chưa giao, tìm điểm log cuối cùng không phải là receiver
  for (let i = routePoints.value.length - 1; i >= 0; i--) {
    if (routePoints.value[i].id !== 'receiver') {
      return i;
    }
  }
  return 0;
});

const refreshMapSize = () => {
  nextTick(() => {
    setTimeout(() => {
      if (map.value?.leafletObject) {
        map.value.leafletObject.invalidateSize();
      }
    }, 200);
  });
};

onMounted(() => {
  if (routePoints.value.length > 0) {
    const idx = currentIdx.value !== -1 ? currentIdx.value : 0;
    const current = routePoints.value[idx];
    center.value = [current.latitude, current.longitude];
    zoom.value = 12;
  }
  refreshMapSize();
  window.addEventListener('resize', refreshMapSize);
});

// Mảng [lat, lng] cho polyline nối các điểm theo trình tự hành trình (đường chim bay cơ bản).
const routePath = computed(() => routePoints.value.map((l) => [l.latitude, l.longitude]));

// Mảng [lat, lng] chi tiết đường bộ từ OSRM.
const detailedRoutePath = ref([]);
const routingDistance = ref('');
// Điểm đã được OSRM "snap" (kéo) vào đúng mặt đường.
const snappedPoints = ref([]);

// Theo dõi routePath, gọi OpenRouteService API để lấy đường bộ thật
watch(routePath, async (path) => {
  if (path.length < 2) {
    detailedRoutePath.value = path;
    snappedPoints.value = [];
    routingDistance.value = '';
    return;
  }

  const orsKey = import.meta.env.VITE_ORS_API_KEY || '';

  if (!orsKey) {
    // Fallback: đường chim bay nếu không cấu hình key
    detailedRoutePath.value = path;
    snappedPoints.value = [];
    routingDistance.value = `${haversineDistance(path)} km (đường chim bay)`;
    return;
  }

  try {
    // ORS nhận [lng, lat]
    const coordinates = path.map(p => [p[1], p[0]]);
    const res = await fetch('https://api.openrouteservice.org/v2/directions/driving-car/geojson', {
      method: 'POST',
      headers: {
        'Authorization': orsKey,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ coordinates }),
    });
    const data = await res.json();

    if (data.features && data.features.length > 0) {
      const feature = data.features[0];
      // ORS GeoJSON trả về [lng, lat], Leaflet cần [lat, lng]
      detailedRoutePath.value = feature.geometry.coordinates.map(c => [c[1], c[0]]);
      snappedPoints.value = [];
      const distM = feature.properties?.summary?.distance || 0;
      const distKm = (distM / 1000).toFixed(1);
      const durationMin = Math.round((feature.properties?.summary?.duration || 0) / 60);
      routingDistance.value = `${distKm} km · ~${durationMin} phút (đường bộ)`;
    } else {
      detailedRoutePath.value = path;
      snappedPoints.value = [];
      routingDistance.value = `${haversineDistance(path)} km (đường chim bay)`;
    }
  } catch (err) {
    console.error('Lỗi lấy đường bộ ORS:', err);
    detailedRoutePath.value = path;
    snappedPoints.value = [];
    routingDistance.value = `${haversineDistance(path)} km (đường chim bay)`;
  }
}, { immediate: true, deep: true });

// Tính khoảng cách chim bay theo công thức Haversine
function haversineDistance(path) {
  if (path.length < 2) return '0.0';
  let total = 0;
  for (let i = 0; i < path.length - 1; i++) {
    const [lat1, lng1] = path[i];
    const [lat2, lng2] = path[i + 1];
    const R = 6371;
    const dLat = (lat2 - lat1) * Math.PI / 180;
    const dLng = (lng2 - lng1) * Math.PI / 180;
    const a = Math.sin(dLat/2)**2 + Math.cos(lat1*Math.PI/180) * Math.cos(lat2*Math.PI/180) * Math.sin(dLng/2)**2;
    total += R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
  }
  return total.toFixed(1);
}

// Khung bao trọn lộ trình
const routeBounds = computed(() => {
  if (routePath.value.length < 2) return null;
  const lats = routePath.value.map((p) => p[0]);
  const lngs = routePath.value.map((p) => p[1]);
  return [
    [Math.min(...lats), Math.min(...lngs)],
    [Math.max(...lats), Math.max(...lngs)],
  ];
});

const formatDate = (s) => {
  if (!s) return '';
  return new Date(s).toLocaleString('vi-VN', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit', year: 'numeric' });
};

// Map provider check
const googleKey = import.meta.env.VITE_GOOGLE_MAPS_KEY || '';
const useGoogleMap = computed(
  () => import.meta.env.VITE_MAP_PROVIDER === 'google' && !!googleKey
);

const mapPoints = computed(() =>
  routePoints.value.map((l, idx) => ({
    lat: l.latitude,
    lng: l.longitude,
    label: `${idx + 1}. ${statusLabel(l.status)} — ${formatDate(l.created_at)}`,
  }))
);

// Tính góc quay (bearing) của xe
const vehicleAngle = computed(() => {
  const path = detailedRoutePath.value;
  if (path.length < 2) return 0;
  const [lat1, lng1] = path[path.length - 2];
  const [lat2, lng2] = path[path.length - 1];
  
  const toRad = Math.PI / 180;
  const toDeg = 180 / Math.PI;
  const dLng = (lng2 - lng1) * toRad;
  const y = Math.sin(dLng) * Math.cos(lat2 * toRad);
  const x = Math.cos(lat1 * toRad) * Math.sin(lat2 * toRad) -
            Math.sin(lat1 * toRad) * Math.cos(lat2 * toRad) * Math.cos(dLng);
  let bearing = Math.atan2(y, x) * toDeg;
  return (bearing + 360) % 360;
});

// Icon marker fallback
const COLOR_START = '#16a34a';
const COLOR_MID = '#0891b2';
const COLOR_END = '#dc2626';

const pointIcon = (idx, log) => {
  const isStart = log.id === 'sender';
  const isEnd = log.id === 'receiver';
  const isCurrent = idx === currentIdx.value;
  const color = isStart ? COLOR_START : isEnd ? COLOR_END : COLOR_MID;
  
  if (isCurrent) {
    let svg = '';
    const angle = vehicleAngle.value;
    const isMoving = ['in_transit', 'hub_inbound', 'hub_outbound', 'ready_to_pick'].includes(props.order?.status || log.status);
    
    if (isMoving && !isEnd) {
      // TRUCK
      svg = `<svg viewBox="0 0 64 64" width="64" height="64" style="overflow: visible;">
  <g style="transform-origin: 32px 32px; transform: rotate(${angle}deg);">
    <g transform="translate(20, 12)">
      <!-- Shadow -->
      <rect x="0" y="3" width="24" height="36" rx="4" fill="rgba(0,0,0,0.3)" />
      <!-- Truck Cargo Body -->
      <rect x="2" y="14" width="20" height="24" rx="2" fill="#3b82f6" />
      <!-- Cab -->
      <rect x="3" y="4" width="18" height="10" rx="3" fill="#60a5fa" />
      <!-- Windshield -->
      <path d="M4 10 Q 12 6 20 10 L 18 13 L 6 13 Z" fill="#1e293b" />
      <!-- Headlights -->
      <rect x="4" y="3" width="5" height="2" rx="1" fill="#fef08a" />
      <rect x="15" y="3" width="5" height="2" rx="1" fill="#fef08a" />
      <!-- Tail lights -->
      <rect x="4" y="36" width="5" height="2" rx="1" fill="#ef4444" />
      <rect x="15" y="36" width="5" height="2" rx="1" fill="#ef4444" />
    </g>
  </g>
</svg>`;
    } else if (log && ['delivering', 'delivered'].includes(log.status)) {
      // MOTORBIKE
      svg = `<svg viewBox="0 0 64 64" width="64" height="64" style="overflow: visible;">
  <g style="transform-origin: 32px 32px; transform: rotate(${angle}deg);">
    <g transform="translate(20, 12)">
      <!-- Shadow -->
      <rect x="4" y="2" width="16" height="36" rx="8" fill="rgba(0,0,0,0.3)" />
      <!-- Front tire -->
      <rect x="10" y="2" width="4" height="10" rx="2" fill="#1e293b" />
      <!-- Rear tire -->
      <rect x="10" y="28" width="4" height="10" rx="2" fill="#1e293b" />
      <!-- Chassis -->
      <rect x="8" y="10" width="8" height="22" rx="4" fill="#ef4444" />
      <!-- Handlebars -->
      <rect x="3" y="11" width="18" height="2.5" rx="1.25" fill="#475569" />
      <!-- Dashboard -->
      <path d="M10 9 Q 12 6 14 9" stroke="#cbd5e1" stroke-width="2.5" fill="none" />
      <!-- Headlight -->
      <circle cx="12" cy="3" r="2.5" fill="#fef08a" />
      <!-- Cargo Box -->
      <rect x="4" y="22" width="16" height="15" rx="2.5" fill="#f59e0b" stroke="#d97706" stroke-width="1.5" />
      <!-- Rider Helmet -->
      <circle cx="12" cy="16" r="5" fill="#f8fafc" stroke="#94a3b8" stroke-width="1.5" />
    </g>
  </g>
</svg>`;
    } else {
      svg = `<div style="position: relative; width: 44px; height: 44px; display: flex; align-items: center; justify-content: center;">
  <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="#dc2626" stroke="#fff" stroke-width="2" style="filter: drop-shadow(0 4px 8px rgba(0,0,0,0.3)); transform: translateY(-8px);"><path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z"/><circle cx="12" cy="9" r="2.5" fill="#fff" stroke="none"/></svg>
</div>`;
    }
    
    return L.divIcon({
      className: 'route-pin-premium',
      html: svg,
      iconSize: [64, 64],
      iconAnchor: [32, 32],
    });
  }
  
  // Normal points
  const size = isStart ? 28 : 22;
  const glyph = isStart ? '●' : String(idx + 1);
  const fontSize = isStart ? 14 : 11;
  return L.divIcon({
    className: 'route-pin',
    html: `<div style="width:${size}px;height:${size}px;border-radius:50%;background:${color};border:2px solid #fff;box-shadow:0 1px 4px rgba(15,23,42,.4);display:flex;align-items:center;justify-content:center;color:#fff;font-size:${fontSize}px;font-weight:700;font-family:sans-serif">${glyph}</div>`,
    iconSize: [size, size],
    iconAnchor: [size / 2, size / 2],
  });
};

const pointBadge = (idx) => {
  const last = routePoints.value.length - 1;
  if (idx === 0 && routePoints.value.length > 1) return 'Xuất phát';
  if (idx === last) return 'Vị trí hiện tại';
  return `Điểm ${idx + 1}`;
};

// Tổng quãng đường theo đường chim bay
const fallbackSummary = computed(() => {
  const path = routePath.value;
  if (path.length < 2) return '';
  const R = 6371000;
  const toRad = (d) => (d * Math.PI) / 180;
  let meters = 0;
  for (let i = 1; i < path.length; i++) {
    const [lat1, lng1] = path[i - 1];
    const [lat2, lng2] = path[i];
    const dLat = toRad(lat2 - lat1);
    const dLng = toRad(lng2 - lng1);
    const h = Math.sin(dLat / 2) ** 2 + Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) * Math.sin(dLng / 2) ** 2;
    meters += 2 * R * Math.asin(Math.sqrt(h));
  }
  const dist = meters >= 1000 ? (meters / 1000).toFixed(1) + ' km' : Math.round(meters) + ' m';
  return `${dist} (đường chim bay)`;
});

</script>

<style scoped>
/* Ẩn logo Leaflet / Attribution */
:deep(.leaflet-control-attribution) {
  display: none !important;
}

/* Panel tổng quan fallback */
.fb-summary {
  position: absolute;
  top: 12px;
  right: 12px;
  display: flex;
  align-items: center;
  gap: 14px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(6px);
  border: 1px solid rgba(8, 145, 178, 0.2);
  border-radius: 12px;
  padding: 8px 14px;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.12);
  z-index: 500;
  pointer-events: none;
}

.fb-item {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.fb-label {
  font-size: 10px;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.fb-value {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
  font-variant-numeric: tabular-nums;
}

/* Legend chú thích */
.fb-legend {
  position: absolute;
  bottom: 12px;
  left: 12px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(6px);
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 10px;
  padding: 8px 12px;
  font-size: 11px;
  color: #334155;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.1);
  z-index: 500;
  pointer-events: none;
}

.fb-row {
  display: flex;
  align-items: center;
  gap: 7px;
}

.fb-row + .fb-row {
  margin-top: 4px;
}

.fb-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 2px solid #fff;
  box-shadow: 0 0 0 1px rgba(15, 23, 42, 0.15);
  flex-shrink: 0;
}

.fb-start { background: #16a34a; }
.fb-mid { background: #0891b2; }
.fb-end { background: #dc2626; }

/* Dark Premium Tooltip */
:deep(.leaflet-tooltip.premium-tooltip) {
  background-color: #222 !important;
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  color: #fff !important;
  border-radius: 12px !important;
  padding: 10px 14px !important;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4) !important;
  font-family: inherit !important;
  white-space: nowrap !important;
}
:deep(.leaflet-tooltip.premium-tooltip::before) {
  border-top-color: #222 !important;
}
</style>
