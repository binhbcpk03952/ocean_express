<template>
  <div class="route-map-wrap">
    <div ref="mapEl" class="route-map-canvas"></div>

    <!-- Panel tổng quan lộ trình: quãng đường + thời gian ước tính (từ Directions API). -->
    <div v-if="summary.distance" class="route-map-summary">
      <div class="rms-item">
        <span class="rms-label">Quãng đường</span>
        <span class="rms-value">{{ summary.distance }}</span>
      </div>
      <div class="rms-divider"></div>
      <div class="rms-item">
        <span class="rms-label">Thời gian di chuyển</span>
        <span class="rms-value">{{ summary.duration }}</span>
      </div>
    </div>

    <!-- Chú thích các loại điểm trên lộ trình. -->
    <div v-if="hasPoints" class="route-map-legend">
      <div class="rml-row"><span class="rml-dot rml-start"></span> Điểm xuất phát</div>
      <div class="rml-row"><span class="rml-dot rml-mid"></span> Điểm trung gian</div>
      <div class="rml-row"><span class="rml-dot rml-end"></span> Vị trí hiện tại</div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch, onBeforeUnmount } from 'vue';
import { setOptions, importLibrary } from '@googlemaps/js-api-loader';

const props = defineProps({
  // Danh sách điểm hành trình theo thứ tự thời gian (cũ -> mới).
  // Mỗi điểm: { lat, lng, label } — label hiện trong tooltip.
  points: { type: Array, default: () => [] },
  apiKey: { type: String, required: true },
});

const mapEl = ref(null);
let map = null;
let markers = [];
let directionsRenderer = null;
let fallbackPolyline = null;
let arrowPolyline = null;
let sovereigntyMarkers = [];
let google = null;
let openInfoWindow = null;

// Tổng quan lộ trình hiển thị trên panel overlay.
const summary = reactive({ distance: '', duration: '' });

const hasPoints = computed(() => (props.points || []).length > 0);

// Tâm mặc định: miền Trung VN, zoom vừa đủ thấy cả nước.
const DEFAULT_CENTER = { lat: 16.047079, lng: 108.20623 };

// Tông màu lộ trình đồng bộ brand teal + điểm đầu/cuối nổi bật.
const COLOR_ROUTE = '#0891b2';
const COLOR_START = '#16a34a'; // xanh lá — xuất phát
const COLOR_END = '#dc2626'; // đỏ — vị trí hiện tại / điểm đến

// 2 quần đảo khẳng định chủ quyền — chủ động vẽ đè, nhãn tiếng Việt.
const SOVEREIGNTY = [
  { name: 'Quần đảo Hoàng Sa (Việt Nam)', pos: { lat: 16.5, lng: 112.0 } },
  { name: 'Quần đảo Trường Sa (Việt Nam)', pos: { lat: 9.5, lng: 114.0 } },
];

function clearOverlays() {
  markers.forEach((m) => m.setMap(null));
  markers = [];
  if (openInfoWindow) {
    openInfoWindow.close();
    openInfoWindow = null;
  }
  if (directionsRenderer) {
    directionsRenderer.setMap(null);
    directionsRenderer = null;
  }
  if (fallbackPolyline) {
    fallbackPolyline.setMap(null);
    fallbackPolyline = null;
  }
  if (arrowPolyline) {
    arrowPolyline.setMap(null);
    arrowPolyline = null;
  }
  summary.distance = '';
  summary.duration = '';
}

// Marker điểm xuất phát / điểm đến: giọt nước có ký hiệu (A / cờ).
function endpointIcon(kind) {
  const color = kind === 'start' ? COLOR_START : COLOR_END;
  const glyph = kind === 'start'
    ? '<circle cx="18" cy="15" r="4" fill="#fff"/>'
    : '<path d="M14 9h9l-2 3 2 3h-9z" fill="#fff"/><rect x="13" y="9" width="1.6" height="9" fill="#fff"/>';
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="36" height="46" viewBox="0 0 36 46">
    <path d="M18 45C18 45 33 27 33 15A15 15 0 1 0 3 15C3 27 18 45 18 45Z" fill="${color}" stroke="#fff" stroke-width="2.5"/>
    ${glyph}
  </svg>`;
  return {
    url: 'data:image/svg+xml;charset=UTF-8,' + encodeURIComponent(svg),
    scaledSize: new google.maps.Size(36, 46),
    anchor: new google.maps.Point(18, 45),
    labelOrigin: new google.maps.Point(18, 15),
  };
}

// Marker tròn đánh số thứ tự cho điểm trung gian — tông teal của app.
function numberedIcon(idx) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="26" height="26">
    <circle cx="13" cy="13" r="11" fill="${COLOR_ROUTE}" stroke="#fff" stroke-width="2"/>
    <text x="13" y="17" font-size="11" font-weight="bold" fill="#fff" text-anchor="middle" font-family="sans-serif">${idx}</text>
  </svg>`;
  return {
    url: 'data:image/svg+xml;charset=UTF-8,' + encodeURIComponent(svg),
    scaledSize: new google.maps.Size(26, 26),
    anchor: new google.maps.Point(13, 13),
  };
}

function drawSovereignty() {
  SOVEREIGNTY.forEach((s) => {
    const marker = new google.maps.Marker({
      position: s.pos,
      map,
      label: { text: s.name, fontSize: '11px', fontWeight: '600', color: '#0f172a' },
      icon: {
        path: google.maps.SymbolPath.CIRCLE,
        scale: 5,
        fillColor: '#dc2626',
        fillOpacity: 1,
        strokeColor: '#fff',
        strokeWeight: 1.5,
      },
    });
    sovereigntyMarkers.push(marker);
  });
}

// Mũi tên chỉ hướng chạy dọc theo path để thể hiện chiều di chuyển của lộ trình.
function arrowSymbol() {
  return {
    icon: {
      path: google.maps.SymbolPath.FORWARD_CLOSED_ARROW,
      scale: 2.6,
      strokeColor: '#fff',
      strokeWeight: 1,
      fillColor: COLOR_ROUTE,
      fillOpacity: 1,
    },
    offset: '0%',
    repeat: '90px',
  };
}

// Vẽ đường đi bám đường thật qua Directions API. Fallback nối thẳng nếu thất bại
// (điểm ngoài mạng đường, quá nhiều waypoint, hết quota...).
function drawRoute(pts) {
  const path = pts.map((p) => ({ lat: Number(p.lat), lng: Number(p.lng) }));

  if (path.length < 2) return;

  const service = new google.maps.DirectionsService();
  directionsRenderer = new google.maps.DirectionsRenderer({
    map,
    suppressMarkers: true, // dùng marker đánh số riêng
    polylineOptions: {
      strokeColor: COLOR_ROUTE,
      strokeWeight: 5,
      strokeOpacity: 0.9,
      icons: [arrowSymbol()],
    },
    preserveViewport: true, // tự fit bounds thủ công bên dưới
  });

  // Directions: tối đa 25 waypoint giữa origin & destination.
  const origin = path[0];
  const destination = path[path.length - 1];
  const waypoints = path.slice(1, -1).slice(0, 25).map((location) => ({ location, stopover: true }));

  service.route(
    {
      origin,
      destination,
      waypoints,
      travelMode: google.maps.TravelMode.DRIVING,
    },
    (result, status) => {
      if (status === 'OK' && result) {
        directionsRenderer.setDirections(result);
        computeSummary(result);
      } else {
        // Fallback: nối thẳng các điểm + mũi tên hướng.
        directionsRenderer.setMap(null);
        directionsRenderer = null;
        fallbackPolyline = new google.maps.Polyline({
          path,
          map,
          strokeColor: COLOR_ROUTE,
          strokeWeight: 4,
          strokeOpacity: 0.75,
        });
        arrowPolyline = new google.maps.Polyline({
          path,
          map,
          strokeOpacity: 0,
          icons: [arrowSymbol()],
        });
        computeHaversineSummary(path);
      }
    }
  );
}

// Tổng hợp quãng đường + thời gian từ kết quả Directions (cộng dồn các leg).
function computeSummary(result) {
  const legs = result.routes?.[0]?.legs || [];
  let meters = 0;
  let seconds = 0;
  legs.forEach((leg) => {
    meters += leg.distance?.value || 0;
    seconds += leg.duration?.value || 0;
  });
  summary.distance = formatDistance(meters);
  summary.duration = formatDuration(seconds);
}

// Ước tính quãng đường theo đường chim bay khi Directions không khả dụng.
function computeHaversineSummary(path) {
  let meters = 0;
  for (let i = 1; i < path.length; i++) {
    meters += haversine(path[i - 1], path[i]);
  }
  summary.distance = formatDistance(meters) + ' (đường chim bay)';
  summary.duration = '—';
}

function haversine(a, b) {
  const R = 6371000;
  const toRad = (d) => (d * Math.PI) / 180;
  const dLat = toRad(b.lat - a.lat);
  const dLng = toRad(b.lng - a.lng);
  const lat1 = toRad(a.lat);
  const lat2 = toRad(b.lat);
  const h = Math.sin(dLat / 2) ** 2 + Math.cos(lat1) * Math.cos(lat2) * Math.sin(dLng / 2) ** 2;
  return 2 * R * Math.asin(Math.sqrt(h));
}

function formatDistance(meters) {
  if (meters >= 1000) return (meters / 1000).toFixed(1) + ' km';
  return Math.round(meters) + ' m';
}

function formatDuration(seconds) {
  if (!seconds) return '—';
  const h = Math.floor(seconds / 3600);
  const m = Math.round((seconds % 3600) / 60);
  if (h > 0) return `${h} giờ ${m} phút`;
  return `${m} phút`;
}

function render() {
  if (!map || !google) return;
  clearOverlays();

  const pts = props.points || [];
  const bounds = new google.maps.LatLngBounds();
  const lastIdx = pts.length - 1;

  pts.forEach((p, idx) => {
    const position = { lat: Number(p.lat), lng: Number(p.lng) };
    let icon;
    let zIndex = 10;
    if (idx === 0 && pts.length > 1) {
      icon = endpointIcon('start');
      zIndex = 30;
    } else if (idx === lastIdx) {
      icon = endpointIcon('end');
      zIndex = 40; // vị trí hiện tại nổi trên cùng
    } else {
      icon = numberedIcon(idx + 1);
    }

    const marker = new google.maps.Marker({
      position,
      map,
      icon,
      zIndex,
      title: p.label || '',
    });
    if (p.label) {
      const badge = idx === 0 && pts.length > 1
        ? 'Xuất phát'
        : idx === lastIdx
          ? 'Vị trí hiện tại'
          : `Điểm ${idx + 1}`;
      const info = new google.maps.InfoWindow({
        content: `<div style="font-size:12px;max-width:220px;line-height:1.4">
          <div style="font-weight:700;color:${COLOR_ROUTE};margin-bottom:2px">${badge}</div>
          <div style="color:#334155">${escapeHtml(p.label)}</div>
        </div>`,
      });
      marker.addListener('click', () => {
        if (openInfoWindow) openInfoWindow.close();
        info.open(map, marker);
        openInfoWindow = info;
      });
    }
    markers.push(marker);
    bounds.extend(position);
  });

  drawRoute(pts);

  if (pts.length === 1) {
    map.setCenter(bounds.getCenter());
    map.setZoom(14);
  } else if (pts.length > 1) {
    map.fitBounds(bounds, 56);
  }
}

function escapeHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

onMounted(async () => {
  // API loader v2: functional API thay cho class Loader cũ. Nạp các library cần
  // dùng (maps, marker, routes cho Directions) rồi lấy namespace từ window.google.
  setOptions({ key: props.apiKey, v: 'weekly' });
  await Promise.all([
    importLibrary('maps'),
    importLibrary('marker'),
    importLibrary('routes'),
  ]);
  google = window.google;

  map = new google.maps.Map(mapEl.value, {
    center: DEFAULT_CENTER,
    zoom: 5,
    mapTypeControl: false,
    streetViewControl: false,
    fullscreenControl: true,
    gestureHandling: 'cooperative',
    styles: [
      { featureType: 'poi', elementType: 'labels', stylers: [{ visibility: 'off' }] },
      { featureType: 'transit', elementType: 'labels', stylers: [{ visibility: 'off' }] },
    ],
  });

  drawSovereignty();
  render();
});

watch(() => props.points, render, { deep: true });

onBeforeUnmount(() => {
  clearOverlays();
  sovereigntyMarkers.forEach((m) => m.setMap(null));
  sovereigntyMarkers = [];
});
</script>

<style scoped>
.route-map-wrap {
  position: relative;
  height: 100%;
  width: 100%;
}

.route-map-canvas {
  height: 100%;
  width: 100%;
}

/* Panel tổng quan — góc trên trái, nổi trên bản đồ. */
.route-map-summary {
  position: absolute;
  top: 12px;
  left: 12px;
  display: flex;
  align-items: center;
  gap: 14px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(6px);
  border: 1px solid rgba(8, 145, 178, 0.2);
  border-radius: 12px;
  padding: 8px 14px;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.12);
  z-index: 5;
}

.rms-item {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.rms-label {
  font-size: 10px;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.rms-value {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
  font-variant-numeric: tabular-nums;
}

.rms-divider {
  width: 1px;
  align-self: stretch;
  background: rgba(15, 23, 42, 0.1);
}

/* Legend chú thích — góc dưới trái. */
.route-map-legend {
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
  z-index: 5;
}

.rml-row {
  display: flex;
  align-items: center;
  gap: 7px;
}

.rml-row + .rml-row {
  margin-top: 4px;
}

.rml-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 2px solid #fff;
  box-shadow: 0 0 0 1px rgba(15, 23, 42, 0.15);
  flex-shrink: 0;
}

.rml-start { background: #16a34a; }
.rml-mid { background: #0891b2; }
.rml-end { background: #dc2626; }
</style>
