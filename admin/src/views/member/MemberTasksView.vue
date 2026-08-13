<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-lg font-semibold text-strong">Nhiệm vụ của tôi</h1>
        <p class="text-xs text-meta mt-0.5">{{ orders.length }} đơn đang xử lý</p>
      </div>
      <div class="flex items-center gap-2">
        <BaseButton
          v-if="['first_mile_driver', 'last_mile_driver'].includes(authStore.user?.role)"
          variant="outline" size="sm" @click="optimizeRoute" :loading="optimizing"
        >
          <Map class="w-4 h-4 mr-1" /> Tối ưu lộ trình
        </BaseButton>
        <BaseButton
          v-if="['first_mile_driver', 'last_mile_driver'].includes(authStore.user?.role)"
          variant="primary" size="sm" @click="submitCOD" :loading="submittingCod"
        >
          <Wallet class="w-4 h-4 mr-1" /> Nộp COD
        </BaseButton>
        <button
          class="rounded-md p-2 text-meta hover:bg-subtle transition-colors ml-1"
          :class="{ 'animate-spin': loading }"
          @click="fetchOrders"
          aria-label="Làm mới"
        >
          <RefreshCw class="h-5 w-5" />
        </button>
      </div>
    </div>

    <!-- Filter chips -->
    <div class="flex items-center gap-2 overflow-x-auto pb-1 -mx-1 px-1">
      <button
        @click="statusFilter = ''"
        class="px-3 h-8 rounded-full text-xs font-medium whitespace-nowrap transition-colors border shrink-0"
        :class="statusFilter === '' ? 'bg-primary text-white border-transparent' : 'bg-surface text-body border-[var(--border)]'"
      >Tất cả</button>
      <button
        v-for="s in availableStatuses"
        :key="s"
        @click="statusFilter = s"
        class="px-3 h-8 rounded-full text-xs font-medium whitespace-nowrap transition-all border shrink-0"
        :style="statusFilter === s ? { color: statusConfig(s).fg, background: statusConfig(s).bg, borderColor: statusConfig(s).fg } : {}"
        :class="statusFilter === s ? 'font-semibold' : 'bg-surface text-body border-[var(--border)]'"
      >{{ statusConfig(s).label }}</button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 4" :key="i" class="skeleton h-28 w-full rounded-[var(--r-lg)]"></div>
    </div>

    <!-- Empty -->
    <div v-else-if="filteredOrders.length === 0" class="py-16 text-center">
      <PackageCheck class="w-12 h-12 mx-auto text-meta/40 mb-3" />
      <p class="text-meta text-sm">{{ orders.length === 0 ? 'Chưa có nhiệm vụ nào.' : 'Không có đơn khớp bộ lọc.' }}</p>
    </div>

    <!-- Task cards -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 md:gap-4 lg:gap-5">
      <div
        v-for="order in filteredOrders"
        :key="order.id"
        class="flex flex-col rounded-[var(--r-lg)] border bg-surface p-4 shadow-e1 transition-all hover:shadow-e2"
      >
        <div class="flex items-start justify-between gap-3">
          <router-link :to="{ name: 'MemberOrderDetail', params: { id: order.id } }" class="font-mono text-sm font-semibold text-[var(--primary)] hover:underline">
            {{ order.tracking_number }}
          </router-link>
          <StatusBadge :status="order.status" />
        </div>
        <div class="mt-3 flex items-start gap-2.5 text-sm">
          <User class="w-4 h-4 text-meta mt-0.5 shrink-0" />
          <div class="min-w-0">
            <div class="font-medium text-strong truncate">{{ order.receiver_name }}</div>
            <a :href="`tel:${order.receiver_phone}`" class="text-[var(--primary)] text-[13px] font-medium hover:underline">{{ order.receiver_phone }}</a>
          </div>
        </div>
        <div class="mt-2 flex items-start gap-2.5 text-sm">
          <MapPin class="w-4 h-4 text-meta mt-0.5 shrink-0" />
          <div class="text-body text-[13px] leading-relaxed line-clamp-2">{{ order.receiver_address_detail }}</div>
        </div>
        <div class="mt-3 flex items-center justify-between border-t pt-2.5">
          <span class="text-xs text-meta uppercase font-medium tracking-wider">Tiền COD thu</span>
          <span class="font-bold text-[var(--primary)] tabular text-base">{{ formatMoney(order.cod_amount) }}đ</span>
        </div>

        <!-- Quick actions for drivers -->
        <div v-if="getQuickActions(order).length" class="mt-3 pt-3 border-t flex items-center gap-2">
          <button
            v-for="act in getQuickActions(order)"
            :key="act.to"
            @click="handleQuickAction(order, act)"
            class="flex-1 py-1.5 px-2 text-xs font-medium rounded-md transition-colors border text-center"
            :class="act.variant === 'primary' ? 'bg-primary text-white hover:bg-primary-hover border-transparent' : act.variant === 'danger' ? 'bg-rose-50 text-rose-600 border-rose-200 hover:bg-rose-100' : 'bg-subtle text-strong hover:bg-subtle-hover border-[var(--border)]'"
          >
            {{ act.label }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Thu COD khi giao hàng thành công -->
    <BaseModal v-model="showCodModal" title="Xác nhận giao hàng & Thu tiền COD">
      <div v-if="selectedOrder" class="space-y-4">
        <div class="p-3 bg-amber-50 border border-amber-200 rounded-md text-amber-900 text-sm">
          <p class="font-medium">Số tiền COD cần thu từ người nhận:</p>
          <p class="text-2xl font-bold text-amber-700 mt-1 tabular">{{ formatMoney(selectedOrder.cod_amount) }}đ</p>
        </div>
        <div class="text-xs text-meta space-y-1">
          <p>Mã đơn: <span class="font-mono font-medium text-strong">{{ selectedOrder.tracking_number }}</span></p>
          <p>Người nhận: <span class="font-medium text-strong">{{ selectedOrder.receiver_name }}</span> ({{ selectedOrder.receiver_phone }})</p>
          <p>Địa chỉ: {{ selectedOrder.receiver_address_detail }}</p>
        </div>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="showCodModal = false">Hủy</BaseButton>
        <BaseButton variant="primary" :loading="updatingStatus" @click="confirmDelivery">
          Xác nhận đã thu {{ formatMoney(selectedOrder?.cod_amount) }}đ
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { useAuthStore } from '../../stores/authStore';
import { statusConfig } from '../../composables/useStatus';
import { actionsFor } from '../../composables/useMemberActions';
import { RefreshCw, User, MapPin, PackageCheck, Map, Wallet } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import BaseModal from '../../components/ui/BaseModal.vue';

const toast = useToastStore();
const authStore = useAuthStore();

const orders = ref([]);
const loading = ref(true);
const statusFilter = ref('');
const statusConfigObj = statusConfig;

const optimizing = ref(false);
const submittingCod = ref(false);
const showCodModal = ref(false);
const selectedOrder = ref(null);
const updatingStatus = ref(false);

const fetchOrders = async () => {
  loading.value = true;
  try {
    const res = await api.get('/orders');
    if (res.success) orders.value = res.data || [];
  } catch (error) {
    toast.error('Không thể tải danh sách nhiệm vụ');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const getQuickActions = (order) => {
  return actionsFor(authStore.user?.role, order.status) || [];
};

const handleQuickAction = (order, action) => {
  if (action.to === 'delivered') {
    selectedOrder.value = order;
    showCodModal.value = true;
  } else {
    updateOrderStatus(order.id, action.to);
  }
};

const confirmDelivery = async () => {
  if (!selectedOrder.value) return;
  updatingStatus.value = true;
  try {
    await updateOrderStatus(selectedOrder.value.id, 'delivered');
    showCodModal.value = false;
  } finally {
    updatingStatus.value = false;
  }
};

const updateOrderStatus = async (orderId, newStatus) => {
  try {
    await api.put(`/orders/${orderId}/status`, { status: newStatus });
    toast.success('Đã cập nhật trạng thái đơn hàng');
    fetchOrders();
  } catch (err) {
    toast.error(err.response?.data?.error || 'Lỗi cập nhật trạng thái');
  }
};

const optimizeRoute = async () => {
  if (orders.value.length < 2) {
    toast.info('Cần ít nhất 2 đơn để tối ưu');
    return;
  }
  optimizing.value = true;
  try {
    const locations = orders.value.map(o => ({
      id: o.id,
      lat: o.receiver_latitude || 10.762622,
      lng: o.receiver_longitude || 106.660172
    }));
    
    const res = await api.post('/routes/optimize', { locations });
    if (res.success) {
      const optimizedIds = (res.data || []).map(l => l.id);
      orders.value.sort((a, b) => optimizedIds.indexOf(a.id) - optimizedIds.indexOf(b.id));
      toast.success('Đã tối ưu lộ trình');
    }
  } catch (err) {
    toast.error('Lỗi khi tối ưu lộ trình');
  } finally {
    optimizing.value = false;
  }
};

const submitCOD = async () => {
  submittingCod.value = true;
  try {
    const res = await api.post('/orders/submit-cod');
    if (res.success) {
      toast.success(`Nộp thành công ${new Intl.NumberFormat('vi-VN').format(res.data.amount_submitted)}đ`);
      fetchOrders();
    }
  } catch (err) {
    toast.error('Lỗi nộp COD');
  } finally {
    submittingCod.value = false;
  }
};

onMounted(fetchOrders);

const availableStatuses = computed(() => {
  const set = new Set(orders.value.map((o) => o.status));
  return [...set];
});

const filteredOrders = computed(() => {
  if (!statusFilter.value) return orders.value;
  return orders.value.filter((o) => o.status === statusFilter.value);
});

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
</script>
