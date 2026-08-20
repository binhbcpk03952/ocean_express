<template>
  <div class="space-y-4">
    <!-- Back + header -->
    <div class="flex items-center gap-3">
      <button
        class="rounded-md border p-2 text-body hover:bg-subtle transition-colors"
        @click="router.back()"
        aria-label="Quay lại"
      >
        <ArrowLeft class="h-5 w-5" />
      </button>
      <div class="min-w-0">
        <span class="font-mono text-base font-semibold text-strong">{{ order?.tracking_number || '—' }}</span>
      </div>
      <BaseButton v-if="order" variant="secondary" size="sm" @click="showPrintModal = true">
        <Printer class="w-4 h-4 mr-1" /> In vận đơn
      </BaseButton>
      <StatusBadge v-if="order" :status="order.status" class="ml-auto" />
    </div>

    <!-- Shipping Label Modal -->
    <ShippingLabelModal
      v-model="showPrintModal"
      :order="order"
    />

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div class="skeleton h-40 w-full rounded-[var(--r-lg)]"></div>
      <div class="skeleton h-56 w-full rounded-[var(--r-lg)]"></div>
    </div>

    <template v-else-if="order">
      <!-- Người nhận -->
      <div class="rounded-[var(--r-lg)] border bg-surface p-4 shadow-e1 space-y-3">
        <div class="flex items-start gap-3">
          <User class="w-4 h-4 text-meta mt-0.5 shrink-0" />
          <div>
            <div class="font-medium text-strong">{{ order.receiver_name }}</div>
            <a :href="`tel:${order.receiver_phone}`" class="text-[var(--primary)] text-sm">{{ order.receiver_phone }}</a>
          </div>
        </div>
        <div class="flex items-start gap-3">
          <MapPin class="w-4 h-4 text-meta mt-0.5 shrink-0" />
          <div class="text-body text-sm">{{ order.receiver_address_detail }}</div>
        </div>
        <div class="flex items-center justify-between border-t pt-3 text-sm">
          <span class="text-meta flex items-center gap-2"><Wallet class="w-4 h-4" /> Thu hộ (COD)</span>
          <span class="font-bold text-[var(--primary)] tabular">{{ formatMoney(order.cod_amount) }}đ</span>
        </div>
        <div class="flex items-center justify-between text-sm">
          <span class="text-meta flex items-center gap-2"><Weight class="w-4 h-4" /> Khối lượng</span>
          <span class="text-strong tabular">{{ order.weight }}g</span>
        </div>
      </div>

      <!-- Bản đồ Hành trình -->
      <div class="rounded-[var(--r-lg)] border bg-surface p-4 shadow-e1">
        <h2 class="text-sm font-semibold text-strong mb-4">Bản đồ Hành trình</h2>
        <MapTracking :logs="logs" :order="order" />
      </div>

      <!-- Timeline -->
      <div class="rounded-[var(--r-lg)] border bg-surface p-4 shadow-e1">
        <h2 class="text-sm font-semibold text-strong mb-4">Hành trình</h2>
        <div v-if="sortedLogs.length === 0" class="py-6 text-center text-meta text-sm">Chưa có dữ liệu hành trình.</div>
        <ol v-else class="relative border-l-2 border-[var(--border)] ml-1 space-y-5">
          <li v-for="(log, idx) in sortedLogs" :key="log.id" class="relative pl-5">
            <span
              class="absolute -left-[7px] top-1 w-3 h-3 rounded-full border-2 border-[var(--bg-surface)]"
              :style="{ background: statusColor(log.status) }"
            ></span>
            <StatusBadge :status="log.status" />
            <p v-if="log.note" class="text-sm text-body mt-1.5">{{ log.note }}</p>
            <span class="text-xs text-meta tabular block mt-1">{{ formatDate(log.created_at) }}</span>
          </li>
        </ol>
      </div>
    </template>

    <div v-else class="py-20 text-center">
      <PackageX class="w-12 h-12 mx-auto text-meta/40 mb-3" />
      <p class="text-meta">Không tìm thấy vận đơn.</p>
    </div>

    <!-- Action bar (sticky bottom, trên tab bar) -->
    <div
      v-if="order && actions.length"
      class="fixed inset-x-0 bottom-[68px] z-10 border-t bg-surface/95 p-3 backdrop-blur-md space-y-2"
    >
      <BaseButton
        v-for="act in actions"
        :key="act.to"
        :variant="act.variant"
        size="md"
        class="w-full"
        :loading="submitting === act.to"
        :disabled="submitting !== null"
        @click="doAction(act)"
      >
        {{ act.label }}
      </BaseButton>
    </div>
    <!-- Modal Thu COD khi giao hàng thành công -->
    <BaseModal v-model="showCodModal" title="Xác nhận giao hàng & Thu tiền COD">
      <div v-if="order" class="space-y-4">
        <div class="p-3 bg-amber-50 border border-amber-200 rounded-md text-amber-900 text-sm">
          <p class="font-medium">Số tiền COD cần thu từ người nhận:</p>
          <p class="text-2xl font-bold text-amber-700 mt-1 tabular">{{ formatMoney(order.cod_amount) }}đ</p>
        </div>
        <div class="text-xs text-meta space-y-1">
          <p>Mã đơn: <span class="font-mono font-medium text-strong">{{ order.tracking_number }}</span></p>
          <p>Người nhận: <span class="font-medium text-strong">{{ order.receiver_name }}</span> ({{ order.receiver_phone }})</p>
          <p>Địa chỉ: {{ order.receiver_address_detail }}</p>
        </div>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="showCodModal = false">Hủy</BaseButton>
        <BaseButton variant="primary" :loading="submitting !== null" @click="confirmDelivery">
          Xác nhận đã thu {{ formatMoney(order?.cod_amount) }}đ
        </BaseButton>
      </template>
    </BaseModal>

    <!-- Modal Giao hàng thất bại -->
    <BaseModal v-model="showFailModal" title="Báo cáo giao hàng không thành công">
      <div v-if="order" class="space-y-4">
        <div class="text-xs text-meta">
          Mã đơn: <span class="font-mono font-bold text-strong">{{ order.tracking_number }}</span> · Người nhận: <span class="font-bold text-strong">{{ order.receiver_name }}</span>
        </div>

        <div>
          <label class="block text-xs font-semibold text-strong mb-1.5">Lý do không giao được *</label>
          <select
            v-model="failReason"
            class="w-full h-10 px-3 bg-surface border rounded-[var(--r-md)] text-sm text-strong outline-none focus:border-[var(--primary)]"
          >
            <option value="Khách không nghe máy / Thuê bao">Khách không nghe máy / Thuê bao</option>
            <option value="Khách hẹn giao lại ngày khác">Khách hẹn giao lại ngày khác</option>
            <option value="Địa chỉ không rõ / Không tìm thấy người nhận">Địa chỉ không rõ / Không tìm thấy người nhận</option>
            <option value="Khách từ chối nhận hàng (Hủy đơn)">Khách từ chối nhận hàng (Hủy đơn)</option>
            <option value="Lý do khác">Lý do khác</option>
          </select>
        </div>

        <div>
          <label class="block text-xs font-semibold text-strong mb-1.5">Ghi chú chi tiết</label>
          <textarea
            v-model="failNote"
            rows="2"
            placeholder="VD: Gọi 3 cuộc không nghe máy lúc 10h15..."
            class="w-full p-2.5 bg-surface border rounded-[var(--r-md)] text-xs text-strong outline-none focus:border-[var(--primary)] resize-none"
          ></textarea>
        </div>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="showFailModal = false">Hủy</BaseButton>
        <BaseButton variant="danger" :loading="submitting !== null" @click="confirmFailedDelivery">
          Xác nhận giao thất bại
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import api from '../../services/api';
import { useAuthStore } from '../../stores/authStore';
import { useToastStore } from '../../stores/toastStore';
import { statusColor } from '../../composables/useStatus';
import { actionsFor } from '../../composables/useMemberActions';
import { ArrowLeft, User, MapPin, Wallet, Weight, PackageX, Printer } from 'lucide-vue-next';
import BaseButton from '../../components/ui/BaseButton.vue';
import BaseModal from '../../components/ui/BaseModal.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import MapTracking from '../../components/MapTracking.vue';
import ShippingLabelModal from '../../components/ShippingLabelModal.vue';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const toast = useToastStore();
const orderId = route.params.id;

const loading = ref(true);
const order = ref(null);
const logs = ref([]);
const showPrintModal = ref(false);
const showCodModal = ref(false);
const showFailModal = ref(false);
const failReason = ref('Khách không nghe máy / Thuê bao');
const failNote = ref('');
const submitting = ref(null);

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await api.get(`/orders/${orderId}`);
    if (res.success) {
      order.value = res.data.order;
      logs.value = res.data.logs || [];
    }
  } catch (error) {
    toast.error('Không thể tải chi tiết đơn hàng');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchData);

const sortedLogs = computed(() =>
  [...logs.value].reverse()
);

const actions = computed(() => {
  if (!order.value) return [];
  return actionsFor(authStore.user?.role, order.value.status);
});

// Lấy GPS của thiết bị (nếu user cho phép). Không chặn thao tác nếu bị từ chối.
const getPosition = () =>
  new Promise((resolve) => {
    if (!navigator.geolocation) return resolve({ latitude: 0, longitude: 0 });
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve({ latitude: pos.coords.latitude, longitude: pos.coords.longitude }),
      () => resolve({ latitude: 0, longitude: 0 }),
      { timeout: 5000 }
    );
  });

const doAction = async (act) => {
  if (act.to === 'delivered') {
    showCodModal.value = true;
  } else if (act.to === 'delivery_failed') {
    failReason.value = 'Khách không nghe máy / Thuê bao';
    failNote.value = '';
    showFailModal.value = true;
  } else {
    executeStatusUpdate(act.to, act.note || '');
  }
};

const confirmDelivery = async () => {
  await executeStatusUpdate('delivered');
  showCodModal.value = false;
};

const confirmFailedDelivery = async () => {
  await executeStatusUpdate('delivery_failed', failNote.value, failReason.value);
  showFailModal.value = false;
};

const executeStatusUpdate = async (status, note = '', failureReason = '') => {
  submitting.value = status;
  try {
    const { latitude, longitude } = await getPosition();
    const res = await api.put(`/orders/${orderId}/status`, {
      status,
      note,
      failure_reason: failureReason,
      latitude,
      longitude,
    });
    if (res.success) {
      toast.success('Cập nhật trạng thái thành công');
      await fetchData();
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Cập nhật thất bại');
  } finally {
    submitting.value = null;
  }
};

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
const formatDate = (s) => {
  if (!s) return '';
  return new Date(s).toLocaleString('vi-VN', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit' });
};
</script>
