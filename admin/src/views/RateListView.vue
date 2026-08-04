<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Bảng giá cước</h1>
        <p class="text-meta text-sm mt-1">Cấu hình tính cước tự động theo tuyến và khối lượng</p>
      </div>
      <BaseButton
        v-if="authStore.user?.role === 'admin'"
        variant="primary"
        size="md"
        @click="openCreateModal"
      >
        <Plus class="w-4 h-4" /> Thêm cấu hình
      </BaseButton>
    </div>

    <BaseCard body-class="">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="bg-subtle text-meta text-[12px] uppercase tracking-wide">
              <th class="px-5 py-3 font-medium">Tuyến đường</th>
              <th class="px-5 py-3 font-medium">Khối lượng chuẩn</th>
              <th class="px-5 py-3 font-medium">Phí cơ bản</th>
              <th class="px-5 py-3 font-medium">Bước cộng thêm</th>
              <th class="px-5 py-3 font-medium">Phí cộng thêm</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loading">
              <tr v-for="i in 4" :key="i" class="border-t">
                <td v-for="c in 5" :key="c" class="px-5 py-4"><div class="skeleton h-4 w-full"></div></td>
              </tr>
            </template>
            <tr v-else-if="rates.length === 0">
              <td colspan="5" class="px-5 py-16 text-center">
                <Receipt class="w-10 h-10 mx-auto text-meta/40 mb-3" />
                <p class="text-meta text-sm">Chưa có cấu hình giá cước nào.</p>
              </td>
            </tr>
            <tr v-else v-for="rate in rates" :key="rate.id" class="border-t hover:bg-subtle transition-colors">
              <td class="px-5 py-4">
                <span
                  v-if="!rate.from_location_id && !rate.to_location_id"
                  class="inline-flex items-center gap-1.5 rounded-sm px-2.5 py-1 text-xs font-medium"
                  :style="{ color: 'var(--primary)', background: 'var(--primary-soft)' }"
                >
                  <Globe class="w-3.5 h-3.5" /> Toàn quốc (mặc định)
                </span>
                <span v-else class="flex items-center gap-2 text-body font-medium">
                  <span>{{ rate.from_location_id || 'Mọi nơi' }}</span>
                  <ArrowRight class="w-3.5 h-3.5 text-meta" />
                  <span>{{ rate.to_location_id || 'Mọi nơi' }}</span>
                </span>
              </td>
              <td class="px-5 py-4 tabular text-body">{{ formatNumber(rate.base_weight) }} g</td>
              <td class="px-5 py-4 tabular font-semibold text-strong">{{ formatMoney(rate.base_fee) }}đ</td>
              <td class="px-5 py-4 tabular text-body">mỗi {{ formatNumber(rate.extra_weight_step) }} g</td>
              <td class="px-5 py-4 tabular font-semibold text-strong">{{ formatMoney(rate.extra_fee) }}đ</td>
            </tr>
          </tbody>
        </table>
      </div>
    </BaseCard>

    <!-- Modal Create Rate -->
    <BaseModal v-model="showModal" title="Thêm cấu hình giá cước" size="lg">
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <FormSelect v-model="form.from_location_id" label="Gửi từ">
            <option value="">-- Mọi nơi --</option>
            <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }} ({{ l.id }})</option>
          </FormSelect>
          <FormSelect v-model="form.to_location_id" label="Gửi đến">
            <option value="">-- Mọi nơi --</option>
            <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }} ({{ l.id }})</option>
          </FormSelect>
        </div>
        <div class="flex items-start gap-2 px-3 py-2.5 rounded-md text-xs bg-primary-soft text-primary">
          <Info class="w-4 h-4 shrink-0 mt-0.5" />
          <span>Để trống cả hai ô tương đương với giá "mặc định toàn quốc".</span>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <FormField v-model="form.base_weight" label="Khối lượng cơ bản (gram)" type="number" />
          <FormField v-model="form.base_fee" label="Phí cơ bản (VNĐ)" type="number" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <FormField v-model="form.extra_weight_step" label="Bước cộng thêm (gram)" type="number" />
          <FormField v-model="form.extra_fee" label="Phí cộng thêm mỗi bước (VNĐ)" type="number" />
        </div>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="showModal = false">Hủy</BaseButton>
        <BaseButton variant="primary" :loading="isSaving" @click="submitCreate">Lưu cấu hình</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../services/api';
import { useAuthStore } from '../stores/authStore';
import { useToastStore } from '../stores/toastStore';
import { Plus, Receipt, Globe, ArrowRight, Info } from 'lucide-vue-next';
import BaseCard from '../components/ui/BaseCard.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import BaseModal from '../components/ui/BaseModal.vue';
import FormField from '../components/ui/FormField.vue';
import FormSelect from '../components/ui/FormSelect.vue';

const authStore = useAuthStore();
const toast = useToastStore();

const loading = ref(false);
const isSaving = ref(false);
const showModal = ref(false);

const rates = ref([]);
const locations = ref([]);

const form = ref({
  from_location_id: '',
  to_location_id: '',
  base_weight: 1000,
  base_fee: 30000,
  extra_weight_step: 500,
  extra_fee: 5000,
});

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await api.get('/rates');
    if (res.success) rates.value = res.data || [];
  } catch (error) {
    toast.error('Không thể tải bảng giá cước');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const fetchLocations = async () => {
  if (locations.value.length > 0) return;
  try {
    const res = await api.get('/locations');
    if (res.success) locations.value = res.data || [];
  } catch (error) {
    console.error(error);
  }
};

onMounted(fetchData);

const openCreateModal = () => {
  fetchLocations();
  form.value = {
    from_location_id: '',
    to_location_id: '',
    base_weight: 1000,
    base_fee: 30000,
    extra_weight_step: 500,
    extra_fee: 5000,
  };
  showModal.value = true;
};

const submitCreate = async () => {
  isSaving.value = true;
  try {
    const payload = {
      from_location_id: form.value.from_location_id || null,
      to_location_id: form.value.to_location_id || null,
      base_weight: Number(form.value.base_weight),
      base_fee: Number(form.value.base_fee),
      extra_weight_step: Number(form.value.extra_weight_step),
      extra_fee: Number(form.value.extra_fee),
    };
    const res = await api.post('/rates', payload);
    if (res.success) {
      toast.success('Thêm cấu hình giá cước thành công');
      showModal.value = false;
      fetchData();
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Có lỗi xảy ra');
  } finally {
    isSaving.value = false;
  }
};

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
const formatNumber = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
</script>
