<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-strong">Đối soát & Chi trả</h1>
      <p class="text-meta text-sm mt-1">Chốt tiền COD (đã trừ cước) và chi trả cho đối tác</p>
    </div>

    <!-- Chọn shop để xem ví + chốt phiên -->
    <BaseCard title="Đối soát cho đối tác">
      <div class="flex flex-wrap items-end gap-3">
        <div class="flex-1 min-w-[220px]">
          <FormSelect v-model="selectedShopId" label="Chọn đối tác">
            <option value="">— Chọn shop —</option>
            <option v-for="s in shops" :key="s.id" :value="s.id">{{ s.name }}</option>
          </FormSelect>
        </div>
        <BaseButton variant="secondary" size="md" :loading="loadingWallet" :disabled="!selectedShopId" @click="fetchWallet">
          <Search class="w-4 h-4" /> Xem ví
        </BaseButton>
      </div>

      <div v-if="wallet" class="mt-4 flex flex-wrap items-center justify-between gap-4 rounded-[var(--r-md)] bg-subtle px-4 py-3">
        <div>
          <div class="text-xs text-meta">Số dư chưa đối soát</div>
          <div class="text-2xl font-bold tabular" :class="wallet.available_balance >= 0 ? 'text-[var(--st-delivered-fg)]' : 'text-[var(--st-returned-fg)]'">
            {{ formatMoney(wallet.available_balance) }}đ
          </div>
        </div>
        <BaseButton
          variant="primary"
          size="md"
          :loading="creating"
          :disabled="!wallet.available_balance"
          @click="createSettlement"
        >
          <FileCheck class="w-4 h-4" /> Chốt phiên chi trả
        </BaseButton>
      </div>
    </BaseCard>

    <!-- Danh sách phiên chi trả -->
    <BaseCard body-class="p-0">
      <div class="flex items-center justify-between p-4 border-b">
        <h2 class="text-sm font-semibold text-strong">Các phiên chi trả</h2>
        <BaseButton variant="ghost" size="sm" :loading="loadingList" @click="fetchSettlements">
          <RefreshCw class="w-4 h-4" /> Làm mới
        </BaseButton>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="bg-subtle text-meta text-[12px] uppercase tracking-wide">
              <th class="px-5 py-3 font-medium">Đối tác</th>
              <th class="px-5 py-3 font-medium">Số tiền</th>
              <th class="px-5 py-3 font-medium">Chốt lúc</th>
              <th class="px-5 py-3 font-medium">Trạng thái</th>
              <th class="px-5 py-3 font-medium text-right">Thao tác</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loadingList">
              <tr v-for="i in 3" :key="i" class="border-t">
                <td v-for="c in 5" :key="c" class="px-5 py-4"><div class="skeleton h-4 w-full"></div></td>
              </tr>
            </template>
            <tr v-else-if="settlements.length === 0">
              <td colspan="5" class="px-5 py-12 text-center text-meta text-sm">Chưa có phiên chi trả nào.</td>
            </tr>
            <tr v-else v-for="s in settlements" :key="s.id" class="border-t hover:bg-subtle transition-colors">
              <td class="px-5 py-4 text-body">{{ shopName(s.shop_id) }}</td>
              <td class="px-5 py-4 font-semibold text-strong tabular">{{ formatMoney(s.total_amount) }}đ</td>
              <td class="px-5 py-4 text-meta tabular">{{ formatDate(s.created_at) }}</td>
              <td class="px-5 py-4">
                <span
                  class="inline-flex items-center gap-1.5 rounded-sm px-2.5 py-1 text-xs font-medium"
                  :style="s.status === 'paid'
                    ? { color: 'var(--st-delivered-fg)', background: 'var(--st-delivered-bg)' }
                    : { color: 'var(--st-ready-fg)', background: 'var(--st-ready-bg)' }"
                >
                  <span class="h-1.5 w-1.5 rounded-full" style="background: currentColor"></span>
                  {{ s.status === 'paid' ? 'Đã chi trả' : 'Chờ chi trả' }}
                </span>
              </td>
              <td class="px-5 py-4 text-right">
                <BaseButton
                  v-if="s.status !== 'paid'"
                  variant="secondary"
                  size="sm"
                  :loading="markingId === s.id"
                  @click="markPaid(s)"
                >
                  <Check class="w-4 h-4" /> Đánh dấu đã chi
                </BaseButton>
                <span v-else class="text-xs text-meta tabular">{{ formatDate(s.paid_at) }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </BaseCard>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../services/api';
import { useToastStore } from '../stores/toastStore';
import { Search, FileCheck, RefreshCw, Check } from 'lucide-vue-next';
import BaseCard from '../components/ui/BaseCard.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import FormSelect from '../components/ui/FormSelect.vue';

const toast = useToastStore();

const shops = ref([]);
const selectedShopId = ref('');
const wallet = ref(null);
const settlements = ref([]);

const loadingWallet = ref(false);
const loadingList = ref(false);
const creating = ref(false);
const markingId = ref(null);

const fetchShops = async () => {
  try {
    const res = await api.get('/shops', { params: { status: 'approved' } });
    if (res.success) shops.value = res.data || [];
  } catch (error) {
    console.error(error);
  }
};

const fetchWallet = async () => {
  if (!selectedShopId.value) return;
  loadingWallet.value = true;
  try {
    const res = await api.get(`/shops/${selectedShopId.value}/wallet`);
    if (res.success) wallet.value = res.data;
  } catch (error) {
    toast.error('Không thể tải ví đối tác');
    console.error(error);
  } finally {
    loadingWallet.value = false;
  }
};

const fetchSettlements = async () => {
  loadingList.value = true;
  try {
    const res = await api.get('/settlements');
    if (res.success) settlements.value = res.data || [];
  } catch (error) {
    toast.error('Không thể tải danh sách phiên chi trả');
    console.error(error);
  } finally {
    loadingList.value = false;
  }
};

onMounted(() => {
  fetchShops();
  fetchSettlements();
});

const createSettlement = async () => {
  if (!selectedShopId.value) return;
  creating.value = true;
  try {
    const res = await api.post('/settlements', { shop_id: selectedShopId.value });
    if (res.success) {
      toast.success(`Đã chốt phiên chi trả ${formatMoney(res.data.total_amount)}đ`);
      wallet.value = null;
      fetchSettlements();
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Không có giao dịch nào để chốt');
  } finally {
    creating.value = false;
  }
};

const markPaid = async (s) => {
  markingId.value = s.id;
  try {
    const res = await api.patch(`/settlements/${s.id}/paid`);
    if (res.success) {
      toast.success('Đã đánh dấu chi trả thành công');
      fetchSettlements();
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || 'Cập nhật thất bại');
  } finally {
    markingId.value = null;
  }
};

const shopName = (id) => shops.value.find((s) => s.id === id)?.name || id.slice(0, 8);
const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
const formatDate = (s) => {
  if (!s) return '';
  return new Date(s).toLocaleString('vi-VN', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit', year: 'numeric' });
};
</script>
