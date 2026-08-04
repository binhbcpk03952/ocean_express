<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Nhân sự</h1>
        <p class="text-meta text-sm mt-1">Quản lý tài khoản nhân viên và phân quyền</p>
      </div>
      <BaseButton
        v-if="authStore.user?.role === 'admin'"
        variant="primary"
        size="md"
        @click="openCreateModal"
      >
        <UserPlus class="w-4 h-4" /> Thêm nhân sự
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
            placeholder="Tìm tên, số điện thoại..."
            class="w-full h-10 pl-9 pr-3 bg-surface border rounded-[var(--r-md)] text-sm text-strong placeholder:text-meta outline-none focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40 transition-shadow"
          />
        </div>
        <FormSelect v-model="roleFilter" class="min-w-[180px]">
          <option value="">Tất cả vai trò</option>
          <option v-for="r in ROLE_ORDER" :key="r" :value="r">{{ ROLE_META[r].label }}</option>
        </FormSelect>
      </div>

      <!-- Table -->
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="bg-subtle text-meta text-[12px] uppercase tracking-wide">
              <th class="px-5 py-3 font-medium">Họ tên / Email</th>
              <th class="px-5 py-3 font-medium">Số điện thoại</th>
              <th class="px-5 py-3 font-medium">Vai trò</th>
              <th class="px-5 py-3 font-medium">Bưu cục</th>
              <th class="px-5 py-3 font-medium">Trạng thái</th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr v-for="i in 5" :key="i" class="border-t">
              <td v-for="c in 5" :key="c" class="px-5 py-4"><div class="skeleton h-4 w-full"></div></td>
            </tr>
          </tbody>
          <tbody v-else-if="filteredEmployees.length === 0">
            <tr>
              <td colspan="7" class="px-6 py-12 text-center text-gray-500">
                <div class="flex flex-col items-center justify-center">
                  <svg class="w-12 h-12 text-gray-300 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
                  <span class="text-base font-medium text-gray-600">Không tìm thấy nhân viên nào</span>
                  <span class="text-sm text-gray-400 mt-1">Điều chỉnh bộ lọc hoặc thêm mới.</span>
                </div>
              </td>
            </tr>
          </tbody>
          <tbody v-else class="divide-y divide-gray-100">
            <tr
              v-for="emp in filteredEmployees"
              :key="emp.id"
              class="hover:bg-gray-50/80 transition-colors"
            >
              <td class="px-5 py-4">
                <div class="flex items-center gap-3">
                  <span
                    class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-semibold shrink-0"
                    :style="{ color: ROLE_META[emp.role]?.fg || 'var(--text-meta)', background: ROLE_META[emp.role]?.bg || 'var(--bg-subtle)' }"
                  >{{ initials(emp.name) }}</span>
                  <div class="leading-tight">
                    <span class="font-medium text-strong block">{{ emp.name }}</span>
                    <span class="text-[11px] text-meta">{{ emp.email || '—' }}</span>
                  </div>
                </div>
              </td>
              <td class="px-5 py-4 tabular text-body">{{ emp.phone }}</td>
              <td class="px-5 py-4">
                <span
                  class="inline-flex items-center rounded-sm px-2.5 py-1 text-xs font-medium"
                  :style="{ color: ROLE_META[emp.role]?.fg || 'var(--text-meta)', background: ROLE_META[emp.role]?.bg || 'var(--bg-subtle)' }"
                >{{ formatRole(emp.role) }}</span>
              </td>
              <td class="px-5 py-4 text-body">{{ hubName(emp.hub_id) }}</td>
              <td class="px-5 py-4">
                <span
                  class="inline-flex items-center gap-1.5 rounded-sm px-2.5 py-1 text-xs font-medium"
                  :style="emp.is_active
                    ? { color: 'var(--st-delivered-fg)', background: 'var(--st-delivered-bg)' }
                    : { color: 'var(--st-returned-fg)', background: 'var(--st-returned-bg)' }"
                >
                  <span class="h-1.5 w-1.5 rounded-full" :style="{ background: 'currentColor' }"></span>
                  {{ emp.is_active ? 'Hoạt động' : 'Đã khóa' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination UI -->
      <div v-if="totalPages > 1" class="px-6 py-4 border-t border-gray-100 bg-gray-50 flex items-center justify-between">
        <div class="text-sm text-gray-500">
          Trang <span class="font-medium text-gray-900">{{ page }}</span> trên <span class="font-medium text-gray-900">{{ totalPages }}</span>
        </div>
        <div class="flex gap-2">
          <button @click="changePage(page - 1)" :disabled="page <= 1" class="px-3 py-1.5 text-sm font-medium rounded-lg border border-gray-200 bg-white text-gray-600 hover:bg-gray-50 hover:text-gray-900 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            Trước
          </button>
          <button @click="changePage(page + 1)" :disabled="page >= totalPages" class="px-3 py-1.5 text-sm font-medium rounded-lg border border-gray-200 bg-white text-gray-600 hover:bg-gray-50 hover:text-gray-900 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            Sau
          </button>
        </div>
      </div>
    </BaseCard>

    <!-- Modal Create -->
    <BaseModal v-model="showModal" title="Thêm nhân sự mới" subtitle="Tạo tài khoản và phân quyền">
      <div class="space-y-4">
        <FormField v-model="form.name" label="Họ tên" required placeholder="Nguyễn Văn A" />
        <FormField v-model="form.phone" label="Số điện thoại" required placeholder="09..." />
        <FormField v-model="form.email" label="Email" placeholder="abc@email.com" />
        <FormField v-model="form.password" label="Mật khẩu" type="password" required placeholder="••••••••" />
        <FormSelect v-model="form.role" label="Vai trò" required>
          <option v-for="r in ROLE_ORDER" :key="r" :value="r">{{ ROLE_META[r].full }}</option>
        </FormSelect>
        <FormSelect
          v-if="['hub_staff', 'first_mile_driver', 'last_mile_driver'].includes(form.role)"
          v-model="form.hub_id"
          label="Bưu cục trực thuộc"
          hint="Bắt buộc chọn bưu cục cho nhân viên kho và tài xế địa phương."
          required
        >
          <option value="" disabled>-- Chọn bưu cục --</option>
          <option v-for="h in hubs" :key="h.id" :value="h.id">{{ h.name }}</option>
        </FormSelect>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="showModal = false">Hủy</BaseButton>
        <BaseButton variant="primary" :loading="isSaving" @click="submitCreate">Tạo nhân sự</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../services/api';
import { useAuthStore } from '../stores/authStore';
import { useToastStore } from '../stores/toastStore';
import { Search, UserPlus, Users } from 'lucide-vue-next';
import BaseCard from '../components/ui/BaseCard.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import BaseModal from '../components/ui/BaseModal.vue';
import FormField from '../components/ui/FormField.vue';
import FormSelect from '../components/ui/FormSelect.vue';

const authStore = useAuthStore();
const toast = useToastStore();

const ROLE_ORDER = ['admin', 'hub_staff', 'first_mile_driver', 'transit_driver', 'last_mile_driver'];
const ROLE_META = {
  admin:             { label: 'Admin',        full: 'Quản trị viên (Admin)',            fg: 'var(--st-outbound-fg)',  bg: 'var(--st-outbound-bg)' },
  hub_staff:         { label: 'NV Kho',       full: 'Nhân viên kho (Hub Staff)',        fg: 'var(--st-inbound-fg)',   bg: 'var(--st-inbound-bg)' },
  first_mile_driver: { label: 'TX Lấy hàng',  full: 'Tài xế lấy hàng (First-mile)',     fg: 'var(--st-picked-fg)',    bg: 'var(--st-picked-bg)' },
  transit_driver:    { label: 'TX Trung chuyển', full: 'Tài xế trung chuyển (Transit)',   fg: 'var(--st-hub-in-fg)',    bg: 'var(--st-hub-in-bg)' },
  last_mile_driver:  { label: 'TX Giao hàng', full: 'Tài xế giao hàng (Last-mile)',     fg: 'var(--st-delivering-fg)', bg: 'var(--st-delivering-bg)' },
};

const loading = ref(false);
const isSaving = ref(false);
const showModal = ref(false);
const search = ref('');
const roleFilter = ref('');

const employees = ref([]);
const hubs = ref([]);
const page = ref(1);
const limit = ref(10);
const totalPages = ref(1);

const form = ref({ name: '', phone: '', email: '', password: '', role: 'hub_staff', hub_id: '' });

const fetchData = async () => {
  loading.value = true;
  try {
    // Note: Since search is client-side, we would normally send it to backend. 
    // Here we just fetch the current page of employees.
    const res = await api.get('/employees', { params: { page: page.value, limit: limit.value } });
    if (res.success) {
      if (res.data && res.meta) {
         employees.value = res.data;
         totalPages.value = res.meta.total_pages || 1;
      } else {
         // Fallback if API hasn't been reloaded yet
         employees.value = res.data || [];
         totalPages.value = 1;
      }
    }
  } catch (error) {
    toast.error('Không thể tải danh sách nhân sự');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const changePage = (newPage) => {
  if (newPage < 1 || newPage > totalPages.value) return;
  page.value = newPage;
  fetchData();
};

const fetchHubs = async () => {
  try {
    const res = await api.get('/hubs');
    if (res.success) hubs.value = res.data || [];
  } catch (error) {
    console.error(error);
  }
};

onMounted(() => {
  fetchData();
  fetchHubs();
});

const filteredEmployees = computed(() => {
  const q = search.value.trim().toLowerCase();
  return employees.value.filter((e) => {
    if (roleFilter.value && e.role !== roleFilter.value) return false;
    if (!q) return true;
    return (e.name || '').toLowerCase().includes(q) || (e.phone || '').includes(q) || (e.email || '').toLowerCase().includes(q);
  });
});

const openCreateModal = () => {
  form.value = { name: '', phone: '', email: '', password: '', role: 'hub_staff', hub_id: '' };
  showModal.value = true;
};

const submitCreate = async () => {
  if (!form.value.name || !form.value.phone || !form.value.password) {
    toast.warning('Vui lòng điền đầy đủ họ tên, số điện thoại và mật khẩu');
    return;
  }
  
  const requiresHub = ['hub_staff', 'first_mile_driver', 'last_mile_driver'].includes(form.value.role);
  if (requiresHub && !form.value.hub_id) {
    toast.warning('Vui lòng chọn bưu cục trực thuộc cho vai trò này');
    return;
  }

  isSaving.value = true;
  try {
    const payload = { ...form.value };
    if (!requiresHub || !payload.hub_id) {
      payload.hub_id = null;
    }
    const res = await api.post('/employees', payload);
    if (res.success) {
      toast.success('Tạo nhân sự thành công');
      showModal.value = false;
      fetchData();
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Tạo nhân sự thất bại');
  } finally {
    isSaving.value = false;
  }
};

const formatRole = (role) => ROLE_META[role]?.label || role;

const hubName = (hubId) => {
  if (!hubId) return '—';
  const hub = hubs.value.find((h) => h.id === hubId);
  return hub ? hub.name : hubId;
};

const initials = (name) => {
  if (!name) return '?';
  const parts = name.trim().split(/\s+/);
  return (parts[parts.length - 1][0] || '?').toUpperCase();
};
</script>
