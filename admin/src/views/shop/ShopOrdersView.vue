<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-bold text-strong">Vận đơn của tôi</h1>
        <p class="text-meta text-sm mt-1">Theo dõi, điều phối và in tem vận đơn</p>
      </div>
      <div class="flex items-center gap-2">
        <BaseButton variant="secondary" size="md" :loading="loading" @click="fetchOrders">
          <RefreshCw class="w-4 h-4" /> Làm mới
        </BaseButton>
        <router-link :to="{ name: 'ShopCreateOrder' }">
          <BaseButton variant="primary" size="md"><Plus class="w-4 h-4" /> Tạo đơn</BaseButton>
        </router-link>
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
            :style="statusFilter === s ? { color: statusConfig(s).fg, background: statusConfig(s).bg, borderColor: statusConfig(s).fg } : {}"
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
              <th class="px-4 py-3 font-medium">Mã vận đơn</th>
              <th class="px-4 py-3 font-medium">Người nhận</th>
              <th class="px-4 py-3 font-medium">COD / Phí</th>
              <th class="px-4 py-3 font-medium">Trạng thái</th>
              <th class="px-4 py-3 font-medium text-right">Thao tác</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loading">
              <tr v-for="i in 6" :key="i" class="border-t">
                <td v-for="c in 6" :key="c" class="px-4 py-4"><div class="skeleton h-4 w-full"></div></td>
              </tr>
            </template>
            <tr v-else-if="filteredOrders.length === 0">
              <td colspan="6" class="px-5 py-16 text-center">
                <PackageX class="w-10 h-10 mx-auto text-meta/40 mb-3" />
                <p class="text-meta text-sm">{{ orders.length === 0 ? 'Chưa có vận đơn nào. Tạo đơn đầu tiên nhé.' : 'Không tìm thấy đơn khớp bộ lọc.' }}</p>
              </td>
            </tr>
            <tr
              v-else
              v-for="order in filteredOrders"
              :key="order.id"
              class="border-t hover:bg-subtle transition-colors group"
            >
              <!-- Checkbox -->
              <td class="px-4 py-4 text-center" @click.stop>
                <input
                  type="checkbox"
                  :value="order.id"
                  v-model="selectedOrderIds"
                  class="rounded border-slate-300 text-primary focus:ring-primary h-4 w-4 cursor-pointer"
                />
              </td>

              <!-- Tracking Number -->
              <td class="px-4 py-4">
                <router-link
                  :to="{ name: 'ShopOrderDetail', params: { id: order.id } }"
                  class="font-mono text-[13px] font-bold text-primary hover:underline"
                >
                  {{ order.tracking_number }}
                </router-link>
                <div class="text-[11px] text-meta">{{ formatDate(order.created_at) }}</div>
              </td>

              <!-- Receiver -->
              <td class="px-4 py-4">
                <div class="font-semibold text-strong">{{ order.receiver_name }}</div>
                <div class="text-meta text-xs font-mono">{{ order.receiver_phone }}</div>
                <div class="text-meta text-[11px] truncate max-w-xs">{{ order.receiver_address_detail }}</div>
              </td>

              <!-- COD / Fee -->
              <td class="px-4 py-4">
                <div class="font-bold text-strong tabular font-mono" :class="order.cod_amount > 0 ? 'text-red-600' : ''">
                  {{ formatMoney(order.cod_amount) }}đ
                </div>
                <div class="text-meta text-xs tabular font-mono">Phí: {{ formatMoney(order.shipping_fee) }}đ</div>
              </td>

              <!-- Status -->
              <td class="px-4 py-4">
                <StatusBadge :status="order.status" />
              </td>

              <!-- Actions -->
              <td class="px-4 py-4 text-right">
                <div class="flex items-center justify-end gap-1.5" @click.stop>
                  <router-link :to="{ name: 'ShopOrderDetail', params: { id: order.id } }">
                    <BaseButton variant="ghost" size="sm" title="Xem chi tiết & bản đồ lộ trình">
                      <Eye class="w-4 h-4" />
                    </BaseButton>
                  </router-link>

                  <button
                    @click="openSinglePrintModal(order)"
                    class="inline-flex items-center gap-1 px-2.5 py-1.5 bg-subtle hover:bg-subtle-hover text-strong text-xs font-semibold rounded-md border border-[var(--border)] transition-colors cursor-pointer"
                    title="In tem vận đơn"
                  >
                    <Printer class="w-3.5 h-3.5 text-primary" />
                    <span>In Tem</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination UI -->
      <div v-if="totalPages > 1" class="px-5 py-3 border-t bg-gray-50 dark:bg-slate-900/40 flex items-center justify-between text-xs text-meta">
        <div>
          Trang <span class="font-medium text-strong">{{ page }}</span> / <span class="font-medium text-strong">{{ totalPages }}</span>
        </div>
        <div class="flex gap-2">
          <button @click="changePage(page - 1)" :disabled="page <= 1" class="px-3 py-1.5 rounded border bg-white dark:bg-slate-800 disabled:opacity-50 cursor-pointer">
            Trước
          </button>
          <button @click="changePage(page + 1)" :disabled="page >= totalPages" class="px-3 py-1.5 rounded border bg-white dark:bg-slate-800 disabled:opacity-50 cursor-pointer">
            Sau
          </button>
        </div>
      </div>
    </BaseCard>

    <!-- Printable Shipping Label Modal (Single or Batch) -->
    <ShippingLabelModal
      v-model="showPrintModal"
      :order="selectedOrderForPrint"
      :orders="selectedOrdersList"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { STATUS_ORDER, statusConfig } from '../../composables/useStatus';
import { RefreshCw, Search, Plus, PackageX, Printer, Eye, CheckSquare } from 'lucide-vue-next';
import BaseCard from '../../components/ui/BaseCard.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import ShippingLabelModal from '../../components/ShippingLabelModal.vue';

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

const changePage = (newPage) => {
  if (newPage < 1 || newPage > totalPages.value) return;
  page.value = newPage;
  selectedOrderIds.value = [];
  fetchOrders();
};

onMounted(fetchOrders);

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

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);

const formatDate = (s) => {
  if (!s) return '';
  const d = new Date(s);
  return d.toLocaleDateString('vi-VN');
};
</script>
