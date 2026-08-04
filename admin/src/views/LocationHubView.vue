<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Khu vực & Bưu cục</h1>
        <p class="text-meta text-sm mt-1">Quản lý sơ đồ hành chính và mạng lưới bưu cục</p>
      </div>
      <BaseButton
        v-if="authStore.user?.role === 'admin'"
        variant="primary"
        size="md"
        @click="openCreateModal"
      >
        <Plus class="w-4 h-4" /> Thêm {{ activeTab === 'locations' ? 'khu vực' : 'bưu cục' }}
      </BaseButton>
    </div>

    <!-- Actions Row -->
    <div class="flex items-center justify-between gap-4 flex-wrap">
      <!-- Tabs -->
      <div class="inline-flex p-1 rounded-[var(--r-md)] bg-subtle border border-[var(--border)]">
        <button
          @click="activeTab = 'locations'"
          class="px-4 h-9 rounded-[var(--r-sm)] text-sm font-medium transition-colors flex items-center gap-2"
          :class="activeTab === 'locations' ? 'bg-surface text-strong shadow-e1' : 'text-meta hover:text-body'"
        >
          <MapPin class="w-4 h-4" /> Khu vực
        </button>
        <button
          @click="activeTab = 'hubs'"
          class="px-4 h-9 rounded-[var(--r-sm)] text-sm font-medium transition-colors flex items-center gap-2"
          :class="activeTab === 'hubs' ? 'bg-surface text-strong shadow-e1' : 'text-meta hover:text-body'"
        >
          <Warehouse class="w-4 h-4" /> Bưu cục
        </button>
      </div>

      <!-- Filter (Only for locations) -->
      <div v-if="activeTab === 'locations'" class="w-full sm:w-64">
        <FormSelect v-model="selectedProvinceFilter">
          <option value="">-- Tất cả Tỉnh/Thành (Ẩn cấp Xã) --</option>
          <option v-for="p in provincesOnly" :key="p.id" :value="p.id">{{ p.name }}</option>
        </FormSelect>
      </div>
    </div>

    <!-- Locations tab -->
    <BaseCard v-if="activeTab === 'locations'" body-class="">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="bg-subtle text-meta text-[12px] uppercase tracking-wide">
              <th class="px-5 py-3 font-medium">Mã khu vực</th>
              <th class="px-5 py-3 font-medium">Tên</th>
              <th class="px-5 py-3 font-medium">Cấp bậc</th>
              <th class="px-5 py-3 font-medium">Trực thuộc</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loading">
              <tr v-for="i in 5" :key="i" class="border-t">
                <td v-for="c in 4" :key="c" class="px-5 py-4"><div class="skeleton h-4 w-full"></div></td>
              </tr>
            </template>
            <tr v-else-if="locations.length === 0">
              <td colspan="4" class="px-5 py-16 text-center">
                <MapPin class="w-10 h-10 mx-auto text-meta/40 mb-3" />
                <p class="text-meta text-sm">Chưa có khu vực nào.</p>
              </td>
            </tr>
            <tr v-else v-for="loc in filteredLocations" :key="loc.id" class="border-t hover:bg-subtle transition-colors">
              <td class="px-5 py-4"><span class="font-mono text-[13px] font-medium text-strong">{{ loc.id }}</span></td>
              <td class="px-5 py-4 font-medium text-body">{{ loc.name }}</td>
              <td class="px-5 py-4">
                <span class="px-2 py-0.5 rounded-sm text-xs bg-subtle text-meta uppercase tracking-wide">{{ typeLabel(loc.type) }}</span>
              </td>
              <td class="px-5 py-4 text-meta">{{ loc.parent_id || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </BaseCard>

    <!-- Hubs tab -->
    <BaseCard v-else body-class="">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="bg-subtle text-meta text-[12px] uppercase tracking-wide">
              <th class="px-5 py-3 font-medium">Tên bưu cục</th>
              <th class="px-5 py-3 font-medium">Loại kho</th>
              <th class="px-5 py-3 font-medium">Khu vực</th>
              <th class="px-5 py-3 font-medium">Địa chỉ chi tiết</th>
              <th class="px-5 py-3 font-medium">Tọa độ (Lat, Lng)</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loading">
              <tr v-for="i in 5" :key="i" class="border-t">
                <td v-for="c in 5" :key="c" class="px-5 py-4"><div class="skeleton h-4 w-full"></div></td>
              </tr>
            </template>
            <tr v-else-if="hubs.length === 0">
              <td colspan="5" class="px-5 py-16 text-center">
                <Warehouse class="w-10 h-10 mx-auto text-meta/40 mb-3" />
                <p class="text-meta text-sm">Chưa có bưu cục nào.</p>
              </td>
            </tr>
            <tr v-else v-for="hub in hubs" :key="hub.id" class="border-t hover:bg-subtle transition-colors">
              <td class="px-5 py-4 font-medium text-strong">{{ hub.name }}</td>
              <td class="px-5 py-4">
                <span class="px-2 py-0.5 rounded-sm text-xs uppercase tracking-wide" 
                      :class="hub.type === 'soc' ? 'bg-[var(--primary-soft)] text-[var(--primary)]' : 'bg-subtle text-meta'">
                  {{ hub.type === 'soc' ? 'Kho Tổng (SOC)' : 'Bưu cục (LM Hub)' }}
                </span>
              </td>
              <td class="px-5 py-4">
                <span class="px-2 py-0.5 rounded-sm text-xs" :style="{ color: 'var(--primary)', background: 'var(--primary-soft)' }">{{ hub.location_id || '—' }}</span>
              </td>
              <td class="px-5 py-4 text-body max-w-xs truncate" :title="hub.address_detail">{{ hub.address_detail }}</td>
              <td class="px-5 py-4 text-meta font-mono text-xs">
                {{ hub.latitude ? `${hub.latitude.toFixed(4)}, ${hub.longitude.toFixed(4)}` : '—' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </BaseCard>

    <!-- Modal Create -->
    <BaseModal v-model="showModal" :title="`Thêm ${activeTab === 'locations' ? 'khu vực' : 'bưu cục'} mới`">
      <!-- Form Location -->
      <div v-if="activeTab === 'locations'" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <FormField v-model="locForm.id" label="Mã khu vực" placeholder="VD: VN-HN" required />
          <FormField v-model="locForm.name" label="Tên" placeholder="VD: Hà Nội" required />
        </div>
        <FormSelect v-model="locForm.type" label="Cấp bậc">
          <option value="province">Tỉnh/Thành phố (Province)</option>
          <option value="district">Quận/Huyện (District)</option>
          <option value="ward">Phường/Xã (Ward)</option>
        </FormSelect>
        <FormSelect v-model="locForm.parent_id" label="Khu vực cha (tùy chọn)">
          <option value="">-- Không có --</option>
          <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }} ({{ l.id }})</option>
        </FormSelect>
      </div>

      <!-- Form Hub -->
      <div v-else class="space-y-4">
        <FormField v-model="hubForm.name" label="Tên bưu cục" required />
        <FormSelect v-model="hubForm.type" label="Loại kho" required>
          <option value="soc">Kho Tổng (SOC)</option>
          <option value="lm_hub">Bưu cục địa phương (LM Hub)</option>
        </FormSelect>
        <LocationCascader v-model="hubForm.location_id" label="Khu vực" />
        <FormField v-model="hubForm.address_detail" label="Địa chỉ chi tiết" required />
        <div class="grid grid-cols-2 gap-4">
          <FormField v-model="hubForm.latitude" type="number" step="0.000001" label="Vĩ độ (Latitude)" />
          <FormField v-model="hubForm.longitude" type="number" step="0.000001" label="Kinh độ (Longitude)" />
        </div>
      </div>

      <template #footer>
        <BaseButton variant="secondary" @click="showModal = false">Hủy</BaseButton>
        <BaseButton variant="primary" :loading="isSaving" @click="submitCreate">Lưu lại</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue';
import api from '../services/api';
import { useAuthStore } from '../stores/authStore';
import { useToastStore } from '../stores/toastStore';
import { Plus, MapPin, Warehouse } from 'lucide-vue-next';
import BaseCard from '../components/ui/BaseCard.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import BaseModal from '../components/ui/BaseModal.vue';
import FormField from '../components/ui/FormField.vue';
import FormSelect from '../components/ui/FormSelect.vue';
import LocationCascader from '../components/ui/LocationCascader.vue';

const authStore = useAuthStore();
const toast = useToastStore();

const activeTab = ref('locations');
const loading = ref(false);
const isSaving = ref(false);
const showModal = ref(false);
const selectedProvinceFilter = ref('');

const locations = ref([]);
const hubs = ref([]);

const locForm = ref({ id: '', name: '', type: 'province', parent_id: '' });
const hubForm = ref({ name: '', type: 'lm_hub', location_id: '', address_detail: '', latitude: null, longitude: null });

const provincesOnly = computed(() => {
  return locations.value.filter(l => !l.parent_id).sort((a, b) => a.name.localeCompare(b.name));
});

const filteredLocations = computed(() => {
  if (!selectedProvinceFilter.value) {
    return locations.value.filter(l => !l.parent_id).sort((a, b) => a.name.localeCompare(b.name));
  }
  return locations.value
    .filter(l => l.id === selectedProvinceFilter.value || l.parent_id === selectedProvinceFilter.value)
    .sort((a, b) => {
      if (a.id === selectedProvinceFilter.value) return -1;
      if (b.id === selectedProvinceFilter.value) return 1;
      return a.name.localeCompare(b.name);
    });
});

const fetchData = async () => {
  loading.value = true;
  try {
    if (activeTab.value === 'locations') {
      const res = await api.get('/locations');
      if (res.success) locations.value = res.data || [];
    } else {
      if (locations.value.length === 0) {
        const resLoc = await api.get('/locations');
        if (resLoc.success) locations.value = resLoc.data || [];
      }
      const resHub = await api.get('/hubs');
      if (resHub.success) hubs.value = resHub.data || [];
    }
  } catch (error) {
    toast.error('Không thể tải dữ liệu');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

watch(activeTab, fetchData);
onMounted(fetchData);

const openCreateModal = () => {
  if (activeTab.value === 'locations') {
    locForm.value = { id: '', name: '', type: 'province', parent_id: '' };
  } else {
    hubForm.value = { name: '', type: 'lm_hub', location_id: '', address_detail: '', latitude: null, longitude: null };
  }
  showModal.value = true;
};

const submitCreate = async () => {
  isSaving.value = true;
  try {
    if (activeTab.value === 'locations') {
      const payload = { ...locForm.value };
      if (!payload.parent_id) payload.parent_id = null;
      await api.post('/locations', payload);
    } else {
      const payload = { ...hubForm.value };
      if (!payload.location_id) payload.location_id = null;
      if (payload.latitude === '') payload.latitude = null;
      if (payload.longitude === '') payload.longitude = null;
      await api.post('/hubs', payload);
    }
    toast.success('Thêm mới thành công');
    showModal.value = false;
    fetchData();
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Có lỗi xảy ra');
  } finally {
    isSaving.value = false;
  }
};

const typeLabel = (type) => ({ province: 'Tỉnh/TP', district: 'Quận/Huyện', ward: 'Phường/Xã' }[type] || type);
</script>
