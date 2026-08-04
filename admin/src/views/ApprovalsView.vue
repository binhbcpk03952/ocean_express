<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Duyệt tài khoản</h1>
        <p class="text-meta text-sm mt-1">Xét duyệt đối tác và tài xế đăng ký tự phục vụ</p>
      </div>
      <BaseButton variant="secondary" size="md" :loading="loading" @click="fetchData">
        <RefreshCw class="w-4 h-4" /> Làm mới
      </BaseButton>
    </div>

    <!-- Tabs -->
    <div class="inline-flex p-1 rounded-[var(--r-md)] bg-subtle border border-[var(--border)]">
      <button
        @click="activeTab = 'shops'"
        class="px-4 h-9 rounded-[var(--r-sm)] text-sm font-medium transition-colors flex items-center gap-2"
        :class="activeTab === 'shops' ? 'bg-surface text-strong shadow-e1' : 'text-meta hover:text-body'"
      >
        <Store class="w-4 h-4" /> Đối tác
        <span v-if="pendingShops.length" class="ml-1 rounded-full bg-[var(--danger)] text-white text-[10px] font-semibold px-1.5 py-0.5 leading-none">{{ pendingShops.length }}</span>
      </button>
      <button
        @click="activeTab = 'shippers'"
        class="px-4 h-9 rounded-[var(--r-sm)] text-sm font-medium transition-colors flex items-center gap-2"
        :class="activeTab === 'shippers' ? 'bg-surface text-strong shadow-e1' : 'text-meta hover:text-body'"
      >
        <Truck class="w-4 h-4" /> Tài xế
        <span v-if="pendingShippers.length" class="ml-1 rounded-full bg-[var(--danger)] text-white text-[10px] font-semibold px-1.5 py-0.5 leading-none">{{ pendingShippers.length }}</span>
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 3" :key="i" class="skeleton h-20 w-full rounded-[var(--r-lg)]"></div>
    </div>

    <!-- Shops tab -->
    <template v-else-if="activeTab === 'shops'">
      <div v-if="pendingShops.length === 0" class="py-16 text-center">
        <Store class="w-12 h-12 mx-auto text-meta/40 mb-3" />
        <p class="text-meta text-sm">Không có đối tác nào đang chờ duyệt.</p>
      </div>
      <div v-else class="space-y-3">
        <div v-for="shop in pendingShops" :key="shop.id" class="rounded-[var(--r-lg)] border bg-surface p-4 shadow-e1">
          <div class="flex items-start justify-between gap-4 flex-wrap">
            <div class="min-w-0">
              <div class="font-semibold text-strong">{{ shop.name }}</div>
              <div class="text-sm text-meta mt-0.5">{{ shop.email }}</div>
              <div class="text-sm text-body mt-1">{{ shop.address_detail }}</div>
              <div v-if="shop.location_id" class="text-xs text-meta mt-1">Khu vực: {{ shop.location_id }}</div>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <BaseButton variant="danger" size="sm" :loading="acting === shop.id + ':reject'" @click="review('shop', shop, false)">
                <X class="w-4 h-4" /> Từ chối
              </BaseButton>
              <BaseButton variant="primary" size="sm" :loading="acting === shop.id + ':approve'" @click="review('shop', shop, true)">
                <Check class="w-4 h-4" /> Duyệt
              </BaseButton>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Shippers tab -->
    <template v-else>
      <div v-if="pendingShippers.length === 0" class="py-16 text-center">
        <Truck class="w-12 h-12 mx-auto text-meta/40 mb-3" />
        <p class="text-meta text-sm">Không có tài xế nào đang chờ duyệt.</p>
      </div>
      <div v-else class="space-y-3">
        <div v-for="emp in pendingShippers" :key="emp.id" class="rounded-[var(--r-lg)] border bg-surface p-4 shadow-e1">
          <div class="flex items-start justify-between gap-4 flex-wrap">
            <div class="min-w-0">
              <div class="font-semibold text-strong">{{ emp.name }}</div>
              <div class="text-sm text-meta mt-0.5">{{ emp.phone }}</div>
              <div class="text-sm text-body mt-1">{{ roleLabel(emp.role) }}</div>
              <div class="text-xs text-meta mt-1">Bưu cục: {{ hubName(emp.hub_id) }}</div>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <BaseButton variant="danger" size="sm" :loading="acting === emp.id + ':reject'" @click="review('shipper', emp, false)">
                <X class="w-4 h-4" /> Từ chối
              </BaseButton>
              <BaseButton variant="primary" size="sm" :loading="acting === emp.id + ':approve'" @click="review('shipper', emp, true)">
                <Check class="w-4 h-4" /> Duyệt
              </BaseButton>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Modal API key (hiện 1 lần khi duyệt shop lần đầu) -->
    <BaseModal v-model="showKeyModal" title="API Key của đối tác" :close-on-backdrop="false">
      <div class="space-y-4">
        <div class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-warning/10 text-[var(--warning)]">
          <TriangleAlert class="w-4 h-4 mt-0.5 shrink-0" />
          <span>Gửi key này cho đối tác. Vì lý do bảo mật, key sẽ <b>không hiển thị lại</b> lần nào nữa.</span>
        </div>
        <div class="bg-subtle border rounded-[var(--r-md)] p-3 font-mono text-sm text-strong break-all select-all">
          {{ createdApiKey }}
        </div>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="copyKey">
          <Check v-if="copied" class="w-4 h-4" /><Copy v-else class="w-4 h-4" />
          {{ copied ? 'Đã copy' : 'Copy' }}
        </BaseButton>
        <BaseButton variant="primary" @click="showKeyModal = false">Đã lưu, đóng</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../services/api';
import { useToastStore } from '../stores/toastStore';
import { Store, Truck, RefreshCw, Check, X, Copy, TriangleAlert } from 'lucide-vue-next';
import BaseButton from '../components/ui/BaseButton.vue';
import BaseModal from '../components/ui/BaseModal.vue';

const toast = useToastStore();

const activeTab = ref('shops');
const loading = ref(false);
const acting = ref('');

const pendingShops = ref([]);
const pendingShippers = ref([]);
const hubs = ref([]);

const showKeyModal = ref(false);
const createdApiKey = ref('');
const copied = ref(false);

const ROLE_LABELS = {
  first_mile_driver: 'Tài xế lấy hàng',
  last_mile_driver: 'Tài xế giao hàng',
};
const roleLabel = (r) => ROLE_LABELS[r] || r;

const fetchData = async () => {
  loading.value = true;
  try {
    const [shopRes, empRes, hubRes] = await Promise.all([
      api.get('/shops', { params: { status: 'pending' } }),
      api.get('/employees', { params: { status: 'pending' } }),
      api.get('/hubs'),
    ]);
    if (shopRes.success) pendingShops.value = shopRes.data || [];
    if (empRes.success) pendingShippers.value = empRes.data || [];
    if (hubRes.success) hubs.value = hubRes.data || [];
  } catch (error) {
    toast.error('Không thể tải danh sách chờ duyệt');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchData);

const hubName = (hubId) => {
  if (!hubId) return '—';
  const hub = hubs.value.find((h) => h.id === hubId);
  return hub ? hub.name : hubId;
};

const review = async (kind, item, approve) => {
  acting.value = item.id + (approve ? ':approve' : ':reject');
  try {
    const path = kind === 'shop' ? `/shops/${item.id}/review` : `/employees/${item.id}/review`;
    const res = await api.patch(path, { approve });
    if (res.success) {
      toast.success(approve ? 'Đã duyệt tài khoản' : 'Đã từ chối tài khoản');
      // Shop duyệt lần đầu trả về api_key một lần -> hiện modal.
      if (kind === 'shop' && approve && res.data?.api_key) {
        createdApiKey.value = res.data.api_key;
        copied.value = false;
        showKeyModal.value = true;
      }
      fetchData();
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Thao tác thất bại');
  } finally {
    acting.value = '';
  }
};

const copyKey = async () => {
  try {
    await navigator.clipboard.writeText(createdApiKey.value);
    copied.value = true;
  } catch {
    // clipboard có thể bị chặn; user vẫn select-all copy thủ công được
  }
};
</script>
