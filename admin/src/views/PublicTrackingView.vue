<template>
  <div class="min-h-screen bg-subtle flex flex-col">
    <!-- Public Header -->
    <header class="bg-surface border-b sticky top-0 z-30 shadow-xs">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <router-link to="/tracking" class="flex items-center gap-3 group">
          <div class="w-9 h-9 rounded-lg bg-gradient-to-br from-sky-500 to-teal-500 flex items-center justify-center text-white shadow-sm">
            <Package class="w-5 h-5" />
          </div>
          <div>
            <div class="text-base font-black tracking-wider text-strong group-hover:text-primary transition-colors">
              OCEAN EXPRESS
            </div>
            <div class="text-[10px] text-meta tracking-wider uppercase font-medium">Cổng Tra Cứu Vận Đơn</div>
          </div>
        </router-link>

        <div class="flex items-center gap-3">
          <router-link :to="{ name: 'ShopLogin' }" class="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-semibold text-body hover:text-strong hover:bg-subtle border transition-colors">
            <Store class="w-4 h-4 text-primary" />
            <span>Cổng Đối Tác (Shop)</span>
          </router-link>
          <router-link :to="{ name: 'Login' }" class="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-md text-xs font-semibold bg-primary hover:bg-primary-hover text-white transition-colors shadow-xs">
            <LogIn class="w-4 h-4" />
            <span>Đăng nhập</span>
          </router-link>
        </div>
      </div>
    </header>

    <!-- Main Content Area -->
    <main class="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 py-8 space-y-6">
      
      <!-- Hero Search Card -->
      <div class="bg-surface rounded-[var(--r-lg)] shadow-e1 border border-[var(--border)] p-6 sm:p-8 text-center relative overflow-hidden">
        <div class="max-w-xl mx-auto space-y-3">
          <h1 class="text-2xl sm:text-3xl font-extrabold text-strong tracking-tight">
            Tra cứu hành trình vận đơn
          </h1>
          <p class="text-meta text-sm">
            Nhập mã vận đơn Ocean Express để theo dõi vị trí kiện hàng và thời gian giao dự kiến theo thời gian thực.
          </p>

          <form @submit.prevent="searchTracking" class="mt-4 flex items-center gap-2">
            <div class="relative flex-1">
              <Search class="w-5 h-5 absolute left-3.5 top-1/2 -translate-y-1/2 text-meta" />
              <input
                v-model="trackingNumber"
                type="text"
                required
                class="w-full h-12 pl-11 pr-4 bg-subtle border rounded-[var(--r-md)] text-sm font-mono font-semibold text-strong uppercase outline-none focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40 transition-all placeholder:text-meta/70"
                placeholder="VD: OE-DAK-HN-001 hoặc PKG-..."
              />
            </div>
            <button
              type="submit"
              :disabled="loading || !trackingNumber.trim()"
              class="h-12 px-6 bg-primary hover:bg-primary-hover disabled:opacity-50 text-white rounded-[var(--r-md)] text-sm font-bold flex items-center gap-2 transition-colors shrink-0 shadow-sm"
            >
              <Loader2 v-if="loading" class="w-4 h-4 animate-spin" />
              <Search v-else class="w-4 h-4" />
              <span>Tra cứu</span>
            </button>
          </form>

          <!-- Sample code suggestions -->
          <div class="pt-2 flex items-center justify-center gap-2 flex-wrap text-xs text-meta">
            <span>Thử nhanh mã mẫu:</span>
            <button
              v-for="code in sampleCodes"
              :key="code"
              type="button"
              @click="quickSearch(code)"
              class="font-mono px-2 py-0.5 rounded bg-subtle hover:bg-primary/10 hover:text-primary border transition-colors cursor-pointer"
            >
              {{ code }}
            </button>
          </div>
        </div>
      </div>

      <!-- Result Card -->
      <div v-if="order" class="space-y-6">
        
        <!-- Header Banner: Tracking number + Stepper -->
        <div class="bg-surface rounded-[var(--r-lg)] shadow-e1 border border-[var(--border)] p-6 space-y-6">
          <div class="flex items-center justify-between flex-wrap gap-4 border-b pb-4">
            <div>
              <div class="text-xs text-meta font-medium">Mã vận đơn</div>
              <div class="flex items-center gap-3 mt-0.5">
                <span class="font-mono text-xl sm:text-2xl font-black text-strong">{{ order.tracking_number }}</span>
                <button
                  @click="copyTrackingNumber"
                  class="p-1.5 text-meta hover:text-primary rounded-md hover:bg-subtle transition-colors"
                  title="Sao chép mã đơn"
                >
                  <Copy class="w-4 h-4" />
                </button>
              </div>
            </div>

            <div class="flex items-center gap-3 flex-wrap">
              <StatusBadge :status="order.status" />
              <button
                @click="openPrintModal"
                class="px-3 py-1.5 bg-subtle hover:bg-subtle-hover text-strong text-xs font-semibold rounded-md border transition-colors flex items-center gap-1.5"
              >
                <Printer class="w-4 h-4 text-primary" />
                <span>In Tem Vận Đơn</span>
              </button>
              <button
                @click="copyShareLink"
                class="px-3 py-1.5 bg-subtle hover:bg-subtle-hover text-strong text-xs font-semibold rounded-md border transition-colors flex items-center gap-1.5"
              >
                <Share2 class="w-4 h-4 text-primary" />
                <span>Chia sẻ link</span>
              </button>
            </div>
          </div>

          <!-- 5-Step Visual Delivery Progress Stepper -->
          <div class="pt-2">
            <div class="text-xs font-bold text-strong uppercase tracking-wider mb-4">Tiến trình vận chuyển</div>
            <div class="grid grid-cols-2 sm:grid-cols-5 gap-3">
              <div
                v-for="(step, sIdx) in deliverySteps"
                :key="step.key"
                class="p-3 rounded-lg border flex flex-col justify-between transition-all"
                :class="getStepClass(step, sIdx)"
              >
                <div class="flex items-center justify-between mb-2">
                  <span class="text-[11px] font-bold">0{{ sIdx + 1 }}</span>
                  <component :is="step.icon" class="w-4 h-4" />
                </div>
                <div>
                  <div class="font-bold text-xs">{{ step.label }}</div>
                  <div class="text-[10px] opacity-80 mt-0.5">{{ getStepTime(step.key) || step.desc }}</div>
                </div>
              </div>
            </div>
          </div>
          <!-- Return or Failed Alert Banner -->
          <div v-if="['delivery_failed', 'return_requested', 'returning', 'return_hub', 'returned'].includes(order.status)" class="rounded-lg p-3.5 border flex items-start gap-3 text-xs" :class="order.status === 'returned' ? 'bg-slate-100 text-slate-800 border-slate-300 dark:bg-slate-900/60 dark:text-slate-200 dark:border-slate-700' : 'bg-amber-50 text-amber-900 border-amber-200 dark:bg-amber-950/40 dark:text-amber-200 dark:border-amber-800'">
            <AlertCircle class="w-4 h-4 mt-0.5 shrink-0 text-amber-600 dark:text-amber-400" />
            <div class="space-y-0.5">
              <div class="font-bold text-sm">
                {{ order.status === 'returned' ? 'Đơn hàng đã hoàn trả về Shop' : order.status === 'delivery_failed' ? 'Giao hàng chưa thành công (Chờ phát lại)' : 'Đơn hàng đang trong quy trình chuyển hoàn về người gửi' }}
              </div>
              <p v-if="order.failure_reason" class="opacity-90">Lý do: <span class="font-semibold">{{ order.failure_reason }}</span></p>
              <p v-if="order.delivery_attempts" class="opacity-80">Số lần phát: {{ order.delivery_attempts }}/3 lần</p>
            </div>
          </div>
        </div>

        <!-- 2-Column Details: Info (Left) + Map & Timeline (Right) -->
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          
          <!-- Left Column: Package & Customer (5 cols) -->
          <div class="lg:col-span-5 space-y-6">
            
            <!-- Receiver & Sender Card -->
            <div class="bg-surface rounded-[var(--r-lg)] shadow-e1 border border-[var(--border)] p-5 space-y-4">
              <h3 class="text-sm font-bold text-strong flex items-center gap-2">
                <MapPin class="w-4 h-4 text-primary" />
                <span>Thông tin giao nhận</span>
              </h3>

              <div class="space-y-3 text-xs">
                <!-- Receiver -->
                <div class="p-3 bg-subtle rounded-md border space-y-1">
                  <div class="font-bold text-slate-500 uppercase text-[10px]">Người nhận (Đã bảo mật)</div>
                  <div class="font-bold text-strong text-sm">{{ maskName(order.receiver_name) }}</div>
                  <div class="text-primary font-mono font-semibold">{{ maskPhone(order.receiver_phone) }}</div>
                  <div class="text-body mt-1">{{ order.receiver_address_detail }}</div>
                </div>

                <!-- Sender -->
                <div class="p-3 bg-subtle rounded-md border space-y-1">
                  <div class="font-bold text-slate-500 uppercase text-[10px]">Điểm gửi hàng</div>
                  <div class="font-semibold text-strong">{{ order.sender_address_detail || 'Bưu cục Ocean Express' }}</div>
                </div>
              </div>
            </div>

            <!-- Package Specs & Financials -->
            <div class="bg-surface rounded-[var(--r-lg)] shadow-e1 border border-[var(--border)] p-5 space-y-4">
              <h3 class="text-sm font-bold text-strong flex items-center gap-2">
                <Box class="w-4 h-4 text-primary" />
                <span>Thông tin kiện hàng</span>
              </h3>

              <div class="divide-y text-xs">
                <div class="py-2.5 flex items-center justify-between">
                  <span class="text-meta">Khối lượng:</span>
                  <span class="font-bold text-strong font-mono">{{ order.weight || 500 }} g</span>
                </div>
                <div class="py-2.5 flex items-center justify-between" v-if="order.length && order.width && order.height">
                  <span class="text-meta">Kích thước:</span>
                  <span class="font-mono text-strong">{{ order.length }} x {{ order.width }} x {{ order.height }} cm</span>
                </div>
                <div class="py-2.5 flex items-center justify-between">
                  <span class="text-meta">Thu hộ COD:</span>
                  <span class="font-black text-sm" :class="order.cod_amount > 0 ? 'text-red-600' : 'text-strong'">
                    {{ order.cod_amount > 0 ? formatMoney(order.cod_amount) + 'đ' : 'Không thu COD' }}
                  </span>
                </div>
                <div class="py-2.5 flex items-center justify-between">
                  <span class="text-meta">Ngày tạo đơn:</span>
                  <span class="text-strong tabular">{{ formatDate(order.created_at) }}</span>
                </div>
                <div class="py-2.5 flex items-center justify-between" v-if="order.estimated_delivery_time">
                  <span class="text-meta">Dự kiến giao:</span>
                  <span class="font-semibold text-primary tabular">{{ formatDate(order.estimated_delivery_time) }}</span>
                </div>
              </div>
            </div>

          </div>

          <!-- Right Column: Map Tracking & Timeline (7 cols) -->
          <div class="lg:col-span-7 space-y-6">
            
            <!-- Interactive Map Card -->
            <div class="bg-surface rounded-[var(--r-lg)] shadow-e1 border border-[var(--border)] p-5 space-y-3">
              <div class="flex items-center justify-between">
                <h3 class="text-sm font-bold text-strong flex items-center gap-2">
                  <Navigation class="w-4 h-4 text-primary" />
                  <span>Bản đồ hành trình trực quan (Map Tracking)</span>
                </h3>
                <span class="text-xs text-meta">{{ (order.tracking_logs || []).length }} sự kiện</span>
              </div>

              <!-- MapTracking Component -->
              <div class="w-full rounded-md overflow-hidden border border-[var(--border)]">
                <MapTracking :logs="order.tracking_logs || []" :order="order" />
              </div>
            </div>

            <!-- Detailed Activity Timeline Card -->
            <div class="bg-surface rounded-[var(--r-lg)] shadow-e1 border border-[var(--border)] p-5 space-y-4">
              <h3 class="text-sm font-bold text-strong flex items-center gap-2">
                <Clock class="w-4 h-4 text-primary" />
                <span>Nhật ký hành trình chi tiết</span>
              </h3>

              <div v-if="sortedLogs.length === 0" class="py-8 text-center text-meta text-xs">
                Chưa có nhật ký di chuyển.
              </div>

              <div v-else class="relative pl-6 space-y-6">
                <!-- Timeline vertical line -->
                <div class="absolute top-2 bottom-2 left-[11px] w-0.5 bg-[var(--border)]"></div>

                <div v-for="(log, idx) in sortedLogs" :key="log.id || idx" class="relative">
                  <!-- Timeline dot -->
                  <div
                    class="absolute -left-[30px] top-1 w-3.5 h-3.5 rounded-full border-2 border-surface"
                    :class="idx === 0 ? 'bg-primary ring-4 ring-primary/20' : 'bg-slate-400'"
                  ></div>

                  <div class="bg-subtle p-3.5 rounded-lg border space-y-1">
                    <div class="flex items-center justify-between flex-wrap gap-2">
                      <span class="font-bold text-strong text-xs">{{ statusConfig(log.status).label }}</span>
                      <span class="text-[11px] text-meta font-mono">{{ formatDate(log.created_at) }}</span>
                    </div>
                    <p v-if="log.note" class="text-xs text-body leading-relaxed">{{ log.note }}</p>
                  </div>
                </div>
              </div>
            </div>

          </div>

        </div>

      </div>

      <!-- Not Found State -->
      <div v-else-if="searched && !loading" class="bg-surface rounded-[var(--r-lg)] shadow-e1 border border-[var(--border)] p-12 text-center space-y-3">
        <PackageX class="w-12 h-12 mx-auto text-meta/40" />
        <h3 class="text-base font-bold text-strong">Không tìm thấy vận đơn</h3>
        <p class="text-sm text-meta max-w-md mx-auto">
          Không tìm thấy đơn hàng với mã <strong class="text-strong">{{ trackingNumber }}</strong>. Vui lòng kiểm tra lại mã hoặc liên hệ hotline để được hỗ trợ.
        </p>
      </div>

    </main>

    <!-- Footer -->
    <footer class="bg-surface border-t mt-auto py-6">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 text-center text-xs text-meta space-y-1">
        <p>© 2026 OCEAN EXPRESS — Nền tảng Điều phối & Vận chuyển Hàng hóa Toàn quốc</p>
        <p>Hotline Hỗ Trợ 24/7: <span class="font-semibold text-strong">1900 8888</span> · Email: support@oceanexpress.com</p>
      </div>
    </footer>

    <!-- Shipping Label Modal -->
    <ShippingLabelModal
      v-model="showPrintModal"
      :order="order"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import api from '../services/api';
import { useToastStore } from '../stores/toastStore';
import { statusConfig } from '../composables/useStatus';
import {
  Package, Search, Loader2, Clock, Copy, Share2, Printer, MapPin, Box, Navigation,
  PackageCheck, Truck, CheckCircle2, AlertCircle, PackageX, Store, LogIn
} from 'lucide-vue-next';
import StatusBadge from '../components/ui/StatusBadge.vue';
import MapTracking from '../components/MapTracking.vue';
import ShippingLabelModal from '../components/ShippingLabelModal.vue';

const route = useRoute();
const router = useRouter();
const toast = useToastStore();

const trackingNumber = ref(route.query.code || '');
const loading = ref(false);
const searched = ref(false);
const order = ref(null);
const showPrintModal = ref(false);

const sampleCodes = ['OE-DAK-HN-001', 'OE-TEST-001'];

const deliverySteps = [
  { key: 'ready_to_pick', label: 'Đã tạo đơn', desc: 'Chờ lấy hàng', icon: Box },
  { key: 'picked_up', label: 'Đã lấy hàng', desc: 'Shipper đã nhận', icon: PackageCheck },
  { key: 'in_transit', label: 'Trung chuyển', desc: 'Đang luân chuyển', icon: Truck },
  { key: 'delivering', label: 'Đang giao', desc: 'Shipper đang phát', icon: Navigation },
  { key: 'delivered', label: 'Giao thành công', desc: 'Hoàn tất', icon: CheckCircle2 },
];

const statusProgressIndex = {
  'ready_to_pick': 0,
  'picked_up': 1,
  'hub_inbound': 2,
  'in_transit': 2,
  'hub_outbound': 2,
  'delivering': 3,
  'delivery_failed': 3,
  'return_requested': 3,
  'returning': 3,
  'return_hub': 3,
  'delivered': 4,
  'returned': 4,
};

const currentStepIdx = computed(() => {
  if (!order.value?.status) return 0;
  return statusProgressIndex[order.value.status] ?? 0;
});

const getStepClass = (step, idx) => {
  const current = currentStepIdx.value;
  if (idx < current) {
    return 'bg-emerald-50 border-emerald-300 text-emerald-800 dark:bg-emerald-950/40 dark:border-emerald-700 dark:text-emerald-300';
  }
  if (idx === current) {
    return 'bg-blue-50 border-blue-500 text-blue-800 dark:bg-blue-950/40 dark:border-blue-500 dark:text-blue-200 ring-2 ring-blue-400/30';
  }
  return 'bg-subtle border-[var(--border)] text-meta opacity-60';
};

const getStepTime = (stepKey) => {
  if (!order.value?.tracking_logs) return null;
  const match = order.value.tracking_logs.find(l => {
    if (stepKey === 'ready_to_pick') return l.status === 'ready_to_pick';
    if (stepKey === 'picked_up') return l.status === 'picked_up';
    if (stepKey === 'in_transit') return ['hub_inbound', 'in_transit', 'hub_outbound'].includes(l.status);
    if (stepKey === 'delivering') return l.status === 'delivering';
    if (stepKey === 'delivered') return l.status === 'delivered';
    return false;
  });
  if (match) return formatDateShort(match.created_at);
  return null;
};

const sortedLogs = computed(() => {
  if (!order.value?.tracking_logs) return [];
  return [...order.value.tracking_logs].sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
});

const searchTracking = async () => {
  const code = trackingNumber.value.trim().toUpperCase();
  if (!code) return;

  router.replace({ query: { code } });
  loading.value = true;
  searched.value = true;
  order.value = null;

  try {
    const res = await api.get(`/public/tracking/${encodeURIComponent(code)}`);
    if (res.success) {
      order.value = res.data;
    }
  } catch (error) {
    if (error.response?.status === 404) {
      toast.error('Không tìm thấy mã vận đơn ' + code);
    } else {
      toast.error('Có lỗi xảy ra khi tra cứu vận đơn');
    }
  } finally {
    loading.value = false;
  }
};

const quickSearch = (code) => {
  trackingNumber.value = code;
  searchTracking();
};

const maskName = (name) => {
  if (!name) return 'Khách hàng';
  const parts = name.trim().split(' ');
  if (parts.length === 1) return name[0] + '***';
  return parts[0] + ' ' + parts[parts.length - 1][0] + '***';
};

const maskPhone = (phone) => {
  if (!phone || phone.length < 10) return phone || '';
  return phone.substring(0, 3) + '****' + phone.substring(phone.length - 3);
};

const formatDate = (dateString) => {
  if (!dateString) return '';
  const d = new Date(dateString);
  if (isNaN(d.getTime())) return dateString;
  const h = d.getHours().toString().padStart(2, '0');
  const m = d.getMinutes().toString().padStart(2, '0');
  const D = d.getDate().toString().padStart(2, '0');
  const M = (d.getMonth() + 1).toString().padStart(2, '0');
  const Y = d.getFullYear();
  return `${h}:${m} · ${D}/${M}/${Y}`;
};

const formatDateShort = (dateString) => {
  if (!dateString) return '';
  const d = new Date(dateString);
  if (isNaN(d.getTime())) return '';
  const h = d.getHours().toString().padStart(2, '0');
  const m = d.getMinutes().toString().padStart(2, '0');
  const D = d.getDate().toString().padStart(2, '0');
  const M = (d.getMonth() + 1).toString().padStart(2, '0');
  return `${h}:${m} ${D}/${M}`;
};

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);

const copyTrackingNumber = () => {
  if (!order.value?.tracking_number) return;
  navigator.clipboard.writeText(order.value.tracking_number);
  toast.success('Đã sao chép mã vận đơn');
};

const copyShareLink = () => {
  const url = window.location.origin + '/tracking?code=' + encodeURIComponent(order.value.tracking_number);
  navigator.clipboard.writeText(url);
  toast.success('Đã sao chép link tra cứu đơn hàng');
};

const openPrintModal = () => {
  showPrintModal.value = true;
};

onMounted(() => {
  if (trackingNumber.value) {
    searchTracking();
  }
});
</script>
