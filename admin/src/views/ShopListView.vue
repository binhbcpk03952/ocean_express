<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Đối tác E-commerce</h1>
        <p class="text-meta text-sm mt-1">Quản lý shop tích hợp và cấp phát API key</p>
      </div>
      <BaseButton
        v-if="authStore.user?.role === 'admin'"
        variant="primary"
        size="md"
        @click="openCreateModal"
      >
        <Plus class="w-4 h-4" /> Thêm đối tác
      </BaseButton>
    </div>

    <BaseCard body-class="p-0">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="bg-subtle text-meta text-[12px] uppercase tracking-wide">
              <th class="px-5 py-3 font-medium">Tên shop</th>
              <th class="px-5 py-3 font-medium">SĐT</th>
              <th class="px-5 py-3 font-medium">Webhook URL</th>
              <th class="px-5 py-3 font-medium">Khu vực</th>
              <th class="px-5 py-3 font-medium">Địa chỉ</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loading">
              <tr v-for="i in 4" :key="i" class="border-t">
                <td v-for="c in 4" :key="c" class="px-5 py-4"><div class="skeleton h-4 w-full"></div></td>
              </tr>
            </template>
            <tr v-else-if="shops.length === 0">
              <td colspan="5" class="px-5 py-16 text-center">
                <Store class="w-10 h-10 mx-auto text-meta/40 mb-3" />
                <p class="text-meta text-sm">Chưa có đối tác nào.</p>
              </td>
            </tr>
            <tr v-else v-for="shop in shops" :key="shop.id" class="border-t hover:bg-subtle transition-colors">
              <td class="px-5 py-4 font-medium text-strong">{{ shop.name }}</td>
              <td class="px-5 py-4 font-medium text-strong">{{ shop.phone || '—' }}</td>
              <td class="px-5 py-4">
                <span class="font-mono text-xs text-meta break-all">{{ shop.webhook_url }}</span>
              </td>
              <td class="px-5 py-4">
                <span class="px-2 py-1 rounded-[var(--r-sm)] text-xs bg-primary-soft text-primary-hover font-medium">
                  {{ shop.location_id || '—' }}
                </span>
              </td>
              <td class="px-5 py-4 text-body">{{ shop.address_detail }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </BaseCard>

    <!-- Modal tạo đối tác -->
    <BaseModal v-model="showModal" title="Thêm đối tác mới">
      <div class="space-y-4">
        <FormField v-model="form.name" label="Tên shop" placeholder="VD: BC Sport" required />
        <FormField v-model="form.phone" label="Số điện thoại" placeholder="VD: 0987654321" required />
        <FormField v-model="form.webhook_url" label="Webhook URL" placeholder="https://shop.com/webhook" required />
        <FormSelect v-model="form.location_id" label="Khu vực gửi hàng mặc định">
          <option value="">— Chọn khu vực —</option>
          <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }} ({{ l.id }})</option>
        </FormSelect>
        <FormField v-model="form.address_detail" label="Địa chỉ chi tiết" required />
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="showModal = false">Hủy</BaseButton>
        <BaseButton variant="primary" :loading="isSaving" @click="submitCreate">Tạo đối tác</BaseButton>
      </template>
    </BaseModal>

    <!-- Modal API key (chỉ hiển thị 1 lần) -->
    <BaseModal v-model="showKeyModal" title="API Key của đối tác" :close-on-backdrop="false" @update:model-value="(v) => !v && closeKeyModal()">
      <div class="space-y-4">
        <div class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-warning/10 text-[var(--warning)]">
          <TriangleAlert class="w-4 h-4 mt-0.5 shrink-0" />
          <span>Lưu lại ngay. Vì lý do bảo mật, key này sẽ <b>không hiển thị lại</b> lần nào nữa.</span>
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
        <BaseButton variant="primary" @click="closeKeyModal">Đã lưu, đóng</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../services/api';
import { useAuthStore } from '../stores/authStore';
import { useToastStore } from '../stores/toastStore';
import { Plus, Store, Copy, Check, TriangleAlert } from 'lucide-vue-next';
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
const showKeyModal = ref(false);
const createdApiKey = ref('');
const copied = ref(false);

const shops = ref([]);
const locations = ref([]);
const form = ref({ name: '', phone: '', webhook_url: '', location_id: '', address_detail: '' });

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await api.get('/shops');
    if (res.success) shops.value = res.data || [];
  } catch (error) {
    toast.error('Không thể tải danh sách đối tác');
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
  form.value = { name: '', phone: '', webhook_url: '', location_id: '', address_detail: '' };
  showModal.value = true;
};

const submitCreate = async () => {
  isSaving.value = true;
  try {
    const payload = { ...form.value };
    if (!payload.location_id) payload.location_id = null;
    const res = await api.post('/shops', payload);
    showModal.value = false;
    if (res.success && res.data.api_key) {
      createdApiKey.value = res.data.api_key;
      copied.value = false;
      showKeyModal.value = true;
    }
    toast.success('Đã tạo đối tác mới');
    fetchData();
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Có lỗi xảy ra');
  } finally {
    isSaving.value = false;
  }
};

const copyKey = async () => {
  try {
    await navigator.clipboard.writeText(createdApiKey.value);
    copied.value = true;
  } catch {
    // clipboard có thể bị chặn; người dùng vẫn select-all copy thủ công được
  }
};

const closeKeyModal = () => {
  showKeyModal.value = false;
  createdApiKey.value = '';
  copied.value = false;
};
</script>
