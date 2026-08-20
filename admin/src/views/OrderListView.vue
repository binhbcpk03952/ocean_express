<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Vận đơn</h1>
        <p class="text-meta text-sm mt-1">Theo dõi và cập nhật trạng thái đơn hàng</p>
      </div>
      <div class="flex gap-2">
        <input type="file" ref="fileInput" @change="handleImport" class="hidden" accept=".csv" />
        <BaseButton v-if="authStore.user?.role === 'shop'" variant="outline" size="md" :loading="isImporting" @click="$refs.fileInput.click()">
          <Upload class="w-4 h-4" /> Nhập CSV
        </BaseButton>
        <BaseButton variant="secondary" size="md" :loading="loading" @click="fetchOrders">
          <RefreshCw class="w-4 h-4" /> Làm mới
        </BaseButton>
      </div>
    </div>

    <!-- Floating Batch Action Bar -->
    <div
      v-if="selectedOrderIds.length > 0"
      class="bg-blue-900 text-white rounded-[var(--r-md)] p-3 px-5 shadow-lg flex items-center justify-between gap-4 animate-fade-in"
    >
      <div class="flex items-center gap-2 text-sm font-semibold">
        <CheckSquare class="w-4 h-4 text-teal-300" />
        <span>Đã chọn <strong class="text-teal-300">{{ selectedOrderIds.length }}</strong> vận đơn</span>
      </div>

      <div class="flex items-center gap-3">
        <button
          type="button"
          @click="openBatchPrintModal"
          class="px-4 py-1.5 bg-teal-500 hover:bg-teal-400 text-slate-900 font-bold text-xs rounded-md shadow-xs transition-colors flex items-center gap-1.5 cursor-pointer"
        >
          <Printer class="w-4 h-4" />
          <span>In Hàng Loạt ({{ selectedOrderIds.length }} tem)</span>
        </button>
        <button
          type="button"
          @click="clearSelection"
          class="text-xs text-sky-200 hover:text-white underline cursor-pointer"
        >
          Bỏ chọn
        </button>
      </div>
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
              <th class="px-4 py-3 w-10 text-center">
                <input
                  type="checkbox"
                  :checked="isAllSelected"
                  :indeterminate="isIndeterminate"
                  @change="toggleSelectAll"
                  class="rounded border-slate-300 text-primary focus:ring-primary h-4 w-4 cursor-pointer"
                />
              </th>
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
                <td v-for="c in 6" :key="c" class="px-5 py-4"><div class="skeleton h-4 w-full"></div></td>
              </tr>
            </template>
            <tr v-else-if="filteredOrders.length === 0">
              <td colspan="6" class="px-5 py-16 text-center">
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
              <td class="px-4 py-4 text-center" @click.stop>
                <input
                  type="checkbox"
                  :value="order.id"
                  v-model="selectedOrderIds"
                  class="rounded border-slate-300 text-primary focus:ring-primary h-4 w-4 cursor-pointer"
                />
              </td>
              <td class="px-5 py-4">
                <router-link :to="`/orders/${order.id}`" class="font-mono text-[13px] font-bold text-primary hover:underline">
                  {{ order.tracking_number }}
                </router-link>
              </td>
              <td class="px-5 py-4">
                <div class="font-medium text-strong">{{ order.receiver_name }}</div>
                <div class="text-meta text-xs">{{ order.receiver_phone }}</div>
              </td>
              <td class="px-5 py-4">
                <div class="font-semibold text-strong tabular">{{ formatMoney(order.cod_amount) }}đ</div>
                <div class="text-meta text-xs tabular">Phí: {{ formatMoney(order.shipping_fee) }}đ</div>
              </td>
              <td class="px-5 py-4">
                <StatusBadge :status="order.status" />
                <div v-if="order.delivery_attempts > 0" class="text-meta text-[10px] mt-1">
                  Giao lại: {{ order.delivery_attempts }} lần
                </div>
                <div v-if="isSlaBreached(order)" class="text-red-500 font-bold text-[10px] mt-1 flex items-center gap-1">
                  <AlertCircle class="w-3 h-3" /> QUÁ HẠN SLA
                </div>
              </td>
              <td class="px-5 py-4">
                <div class="flex items-center justify-end gap-1.5" @click.stop>
                  <router-link :to="`/orders/${order.id}`">
                    <BaseButton variant="ghost" size="sm" title="Chi tiết"><Eye class="w-4 h-4" /></BaseButton>
                  </router-link>
                  <button
                    @click="openSinglePrintModal(order)"
                    class="inline-flex items-center gap-1 px-2.5 py-1.5 bg-subtle hover:bg-subtle-hover text-strong text-xs font-semibold rounded-md border transition-colors cursor-pointer"
                    title="In tem vận đơn"
                  >
                    <Printer class="w-3.5 h-3.5 text-primary" />
                    <span>In Tem</span>
                  </button>
                  <BaseButton v-if="canAssign(order)" variant="secondary" size="sm" @click="openAssignModal(order)" title="Phân công">
                    <UserPlus class="w-4 h-4" />
                  </BaseButton>
                  <BaseButton v-if="canUpdateStatus(order)" variant="secondary" size="sm" @click="openStatusModal(order)" title="Cập nhật">
                    <PencilLine class="w-4 h-4" />
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

    <!-- Modal Phân công Shipper -->
    <BaseModal v-model="showAssignModal" :title="`Phân công Shipper cho ${selectedOrder?.tracking_number || ''}`">
      <div class="space-y-4">
        <div class="flex items-center gap-2 text-sm">
          <span class="text-meta">Trạng thái:</span>
          <StatusBadge v-if="selectedOrder" :status="selectedOrder.status" />
        </div>
        <div v-if="loadingShippers" class="text-sm text-meta">Đang tải danh sách Shipper...</div>
        <FormSelect v-else v-model="assignForm.shipper_id" label="Chọn Shipper">
          <option value="" disabled>-- Chọn Shipper --</option>
          <option v-for="s in shippers" :key="s.id" :value="s.id">{{ s.full_name }} ({{ s.phone }})</option>
        </FormSelect>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="showAssignModal = false">Hủy</BaseButton>
        <BaseButton variant="primary" :loading="isAssigning" @click="assignOrder" :disabled="!assignForm.shipper_id">Phân công</BaseButton>
      </template>
    </BaseModal>

    <!-- Shipping Label Modal -->
    <ShippingLabelModal
      v-model="showPrintModal"
      :order="selectedOrderForPrint"
      :orders="selectedOrdersList"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../services/api';
import { useAuthStore } from '../stores/authStore';
import { useToastStore } from '../stores/toastStore';
import { STATUS_ORDER, statusConfig } from '../composables/useStatus';
import { RefreshCw, Search, Eye, PencilLine, PackageX, Upload, AlertCircle, UserPlus, Printer, CheckSquare } from 'lucide-vue-next';
import BaseCard from '../components/ui/BaseCard.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import BaseModal from '../components/ui/BaseModal.vue';
import FormSelect from '../components/ui/FormSelect.vue';
import StatusBadge from '../components/ui/StatusBadge.vue';
import ShippingLabelModal from '../components/ShippingLabelModal.vue';

const authStore = useAuthStore();
const toast = useToastStore();

const orders = ref([]);
const loading = ref(true);
const search = ref('');
const statusFilter = ref('');
const page = ref(1);
const limit = ref(10);
const totalPages = ref(1);

const selectedOrderIds = ref([]);
const showPrintModal = ref(false);
const selectedOrderForPrint = ref(null);
const selectedOrdersList = ref([]);

const showModal = ref(false);
const isUpdating = ref(false);
const selectedOrder = ref(null);
const form = ref({ status: '', note: '' });

// Phân công Shipper
const showAssignModal = ref(false);
const isAssigning = ref(false);
const assignForm = ref({ shipper_id: '' });
const shippers = ref([]);
const loadingShippers = ref(false);

const fetchShippers = async () => {
  loadingShippers.value = true;
  try {
    const res = await api.get('/employees');
    if (res.success) {
      shippers.value = (res.data || []).filter(e => e.role === 'first_mile_driver' || e.role === 'last_mile_driver');
    }
  } catch (error) {
    console.error('Không thể tải shipper', error);
  } finally {
    loadingShippers.value = false;
  }
};

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

const isImporting = ref(false);
const fileInput = ref(null);

const handleImport = async (e) => {
  const file = e.target.files[0];
  if (!file) return;

  const formData = new FormData();
  formData.append('file', file);

  isImporting.value = true;
  try {
    const res = await api.post('/shop/orders/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    if (res.success) {
      toast.success(`Nhập thành công ${res.data.success_count} đơn hàng`);
      if (res.data.errors && res.data.errors.length > 0) {
        console.warn('Lỗi nhập Excel:', res.data.errors);
        toast.error(`Có ${res.data.errors.length} dòng lỗi, xem console`);
      }
      fetchOrders();
    }
  } catch (error) {
    toast.error(error.response?.data?.error || 'Lỗi khi nhập file');
  } finally {
    isImporting.value = false;
    e.target.value = null; // reset input
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

const isAllSelected = computed(() => {
  if (filteredOrders.value.length === 0) return false;
  return filteredOrders.value.every(o => selectedOrderIds.value.includes(o.id));
});

const isIndeterminate = computed(() => {
  const count = selectedOrderIds.value.length;
  return count > 0 && !isAllSelected.value;
});

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedOrderIds.value = [];
  } else {
    selectedOrderIds.value = filteredOrders.value.map(o => o.id);
  }
};

const clearSelection = () => {
  selectedOrderIds.value = [];
};

const openSinglePrintModal = (order) => {
  selectedOrderForPrint.value = order;
  selectedOrdersList.value = [order];
  showPrintModal.value = true;
};

const openBatchPrintModal = () => {
  const list = orders.value.filter(o => selectedOrderIds.value.includes(o.id));
  if (list.length === 0) return;
  selectedOrderForPrint.value = null;
  selectedOrdersList.value = list;
  showPrintModal.value = true;
};

const isSlaBreached = (order) => {
  if (!order.sla_deadline) return false;
  if (['delivered', 'returned'].includes(order.status)) return false;
  return new Date(order.sla_deadline) < new Date();
};

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

const canAssign = (order) => {
  const role = authStore.user?.role;
  return role === 'admin' || role === 'hub_staff';
};

const openAssignModal = (order) => {
  selectedOrder.value = order;
  assignForm.value.shipper_id = order.current_driver_id || '';
  if (shippers.value.length === 0) fetchShippers();
  showAssignModal.value = true;
};

const assignOrder = async () => {
  if (!selectedOrder.value || !assignForm.value.shipper_id) return;
  isAssigning.value = true;
  try {
    const res = await api.post(`/orders/${selectedOrder.value.id}/assign`, {
      shipper_id: assignForm.value.shipper_id
    });
    if (res.success) {
      toast.success('Phân công Shipper thành công');
      showAssignModal.value = false;
      fetchOrders();
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Phân công thất bại');
  } finally {
    isAssigning.value = false;
  }
};

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
</script>
