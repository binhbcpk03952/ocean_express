<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Vận đơn</h1>
        <p class="text-meta text-sm mt-1">Theo dõi và cập nhật trạng thái đơn hàng</p>
      </div>
      <BaseButton variant="secondary" size="md" :loading="loading" @click="fetchOrders">
        <RefreshCw class="w-4 h-4" /> Làm mới
      </BaseButton>
    </div>

    <BaseCard body-class="">
      <!-- Toolbar -->
      <div class="p-4 border-b flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-[220px]">
          <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-meta" />
          <input
            v-model="search"
            type="text"
            placeholder="Tìm mã VĐ, người nhận, SĐT..."
            class="w-full h-10 pl-9 pr-3 bg-surface border rounded-[var(--r-md)] text-sm text-strong placeholder:text-meta outline-none focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40 transition-shadow"
          />
        </div>
        <div class="flex items-center gap-2 overflow-x-auto">
          <button
            @click="statusFilter = ''"
            class="px-3 h-8 rounded-full text-xs font-medium whitespace-nowrap transition-colors border"
            :class="statusFilter === '' ? 'bg-primary text-white border-transparent' : 'bg-surface text-body border-[var(--border)] hover:bg-subtle'"
          >Tất cả</button>
          <button
            v-for="s in STATUS_ORDER"
            :key="s"
            @click="statusFilter = s"
            class="px-3 h-8 rounded-full text-xs font-medium whitespace-nowrap transition-all border"
            :style="statusFilter === s
              ? { color: statusConfig(s).fg, background: statusConfig(s).bg, borderColor: statusConfig(s).fg }
              : {}"
            :class="statusFilter === s ? 'font-semibold' : 'bg-surface text-body border-[var(--border)] hover:bg-subtle'"
          >{{ statusConfig(s).label }}</button>
        </div>
      </div>

      <!-- Table -->
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="bg-subtle text-meta text-[12px] uppercase tracking-wide">
              <th class="px-5 py-3 font-medium">Mã vận đơn</th>
              <th class="px-5 py-3 font-medium">Người nhận</th>
              <th class="px-5 py-3 font-medium">COD / Phí</th>
              <th class="px-5 py-3 font-medium">Trạng thái</th>
              <th class="px-5 py-3 font-medium text-right">Thao tác</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loading">
              <tr v-for="i in 6" :key="i" class="border-t">
                <td v-for="c in 5" :key="c" class="px-5 py-4"><div class="skeleton h-4 w-full"></div></td>
              </tr>
            </template>
            <tr v-else-if="filteredOrders.length === 0">
              <td colspan="5" class="px-5 py-16 text-center">
                <PackageX class="w-10 h-10 mx-auto text-meta/40 mb-3" />
                <p class="text-meta text-sm">{{ orders.length === 0 ? 'Chưa có vận đơn nào.' : 'Không tìm thấy vận đơn khớp bộ lọc.' }}</p>
              </td>
            </tr>
            <tr
              v-else
              v-for="order in filteredOrders"
              :key="order.id"
              class="border-t hover:bg-subtle transition-colors"
            >
              <td class="px-5 py-4">
                <span class="font-mono text-[13px] font-medium text-strong">{{ order.tracking_number }}</span>
              </td>
              <td class="px-5 py-4">
                <div class="font-medium text-strong">{{ order.receiver_name }}</div>
                <div class="text-meta text-xs">{{ order.receiver_phone }}</div>
              </td>
              <td class="px-5 py-4">
                <div class="font-semibold text-strong tabular">{{ formatMoney(order.cod_amount) }}đ</div>
                <div class="text-meta text-xs tabular">Phí: {{ formatMoney(order.shipping_fee) }}đ</div>
              </td>
              <td class="px-5 py-4"><StatusBadge :status="order.status" /></td>
              <td class="px-5 py-4">
                <div class="flex items-center justify-end gap-2">
                  <router-link :to="`/orders/${order.id}`">
                    <BaseButton variant="ghost" size="sm"><Eye class="w-4 h-4" /> Chi tiết</BaseButton>
                  </router-link>
                  <BaseButton v-if="canUpdateStatus(order)" variant="secondary" size="sm" @click="openStatusModal(order)">
                    <PencilLine class="w-4 h-4" /> Cập nhật
                  </BaseButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Footer count -->
      <div v-if="totalPages > 1" class="px-5 py-3 border-t bg-gray-50 flex items-center justify-between text-xs text-meta">
        <div>
          Trang <span class="font-medium text-strong">{{ page }}</span> / <span class="font-medium text-strong">{{ totalPages }}</span>
        </div>
        <div class="flex gap-2">
          <button @click="changePage(page - 1)" :disabled="page <= 1" class="px-3 py-1.5 rounded border bg-white disabled:opacity-50">
            Trước
          </button>
          <button @click="changePage(page + 1)" :disabled="page >= totalPages" class="px-3 py-1.5 rounded border bg-white disabled:opacity-50">
            Sau
          </button>
        </div>
      </div>
    </BaseCard>

    <!-- Modal cập nhật trạng thái -->
    <BaseModal v-model="showModal" :title="`Cập nhật ${selectedOrder?.tracking_number || ''}`">
      <div class="space-y-4">
        <div class="flex items-center gap-2 text-sm">
          <span class="text-meta">Hiện tại:</span>
          <StatusBadge v-if="selectedOrder" :status="selectedOrder.status" />
        </div>
        <FormSelect v-model="form.status" label="Trạng thái mới">
          <option v-for="s in STATUS_ORDER" :key="s" :value="s">{{ statusConfig(s).label }}</option>
        </FormSelect>
        <div>
          <label class="block text-[13px] leading-4 font-medium text-body mb-1.5">Ghi chú</label>
          <textarea
            v-model="form.note"
            rows="3"
            placeholder="Lý do hoàn hàng, khách hẹn lại..."
            class="w-full px-3.5 py-2.5 bg-surface border rounded-[var(--r-md)] text-sm text-strong placeholder:text-meta outline-none focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40 transition-shadow resize-none"
          ></textarea>
        </div>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="showModal = false">Hủy</BaseButton>
        <BaseButton variant="primary" :loading="isUpdating" @click="updateStatus">Lưu thay đổi</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../services/api';
import { useAuthStore } from '../stores/authStore';
import { useToastStore } from '../stores/toastStore';
import { STATUS_ORDER, statusConfig } from '../composables/useStatus';
import { RefreshCw, Search, Eye, PencilLine, PackageX } from 'lucide-vue-next';
import BaseCard from '../components/ui/BaseCard.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import BaseModal from '../components/ui/BaseModal.vue';
import FormSelect from '../components/ui/FormSelect.vue';
import StatusBadge from '../components/ui/StatusBadge.vue';

const authStore = useAuthStore();
const toast = useToastStore();

const orders = ref([]);
const loading = ref(true);
const search = ref('');
const statusFilter = ref('');
const page = ref(1);
const limit = ref(10);
const totalPages = ref(1);

const showModal = ref(false);
const isUpdating = ref(false);
const selectedOrder = ref(null);
const form = ref({ status: '', note: '' });

const fetchOrders = async () => {
  loading.value = true;
  try {
    const res = await api.get('/orders', { params: { page: page.value, limit: limit.value } });
    if (res.success) {
      if (res.data && res.meta) {
         orders.value = res.data;
         totalPages.value = res.meta.total_pages || 1;
      } else {
         orders.value = res.data || [];
         totalPages.value = 1;
      }
    }
  } catch (error) {
    toast.error('Không thể tải danh sách vận đơn');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const filteredOrders = computed(() => {
  const q = search.value.trim().toLowerCase();
  return orders.value.filter((o) => {
    if (statusFilter.value && o.status !== statusFilter.value) return false;
    if (!q) return true;
    return (
      (o.tracking_number || '').toLowerCase().includes(q) ||
      (o.receiver_name || '').toLowerCase().includes(q) ||
      (o.receiver_phone || '').includes(q)
    );
  });
});

const openStatusModal = (order) => {
  selectedOrder.value = order;
  form.value.status = order.status;
  form.value.note = '';
  showModal.value = true;
};

const updateStatus = async () => {
  if (!selectedOrder.value) return;
  isUpdating.value = true;
  try {
    const res = await api.put(`/orders/${selectedOrder.value.id}/status`, {
      status: form.value.status,
      note: form.value.note,
    });
    if (res.success) {
      toast.success('Cập nhật trạng thái thành công');
      showModal.value = false;
      fetchOrders();
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Cập nhật thất bại');
  } finally {
    isUpdating.value = false;
  }
};

onMounted(fetchOrders);

const changePage = (newPage) => {
  if (newPage < 1 || newPage > totalPages.value) return;
  page.value = newPage;
  fetchOrders();
};

const canUpdateStatus = (order) => {
  const role = authStore.user?.role;
  if (!role) return false;
  if (role === 'admin' || role === 'hub_staff') return true;
  if (role === 'first_mile_driver' || role === 'last_mile_driver') {
    return !order.current_driver_id || order.current_driver_id === authStore.user.id;
  }
  return false;
};

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
</script>
