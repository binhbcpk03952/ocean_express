<template>
  <div class="space-y-6 max-w-3xl">
    <div>
      <h1 class="text-2xl font-semibold text-strong">Thông tin tài khoản</h1>
      <p class="text-meta text-sm mt-1">Hồ sơ đối tác, khu vực lấy hàng và cấu hình API tích hợp</p>
    </div>

    <div v-if="loading" class="space-y-4">
      <div class="skeleton h-40 w-full rounded-[var(--r-lg)]"></div>
      <div class="skeleton h-32 w-full rounded-[var(--r-lg)]"></div>
    </div>

    <template v-else-if="shop">
      <!-- Nhắc cấu hình khu vực khi còn thiếu -->
      <div
        v-if="!shop.location_id"
        class="flex items-start gap-2 px-4 py-3 rounded-[var(--r-md)] text-sm bg-warning/10 text-[var(--warning)]"
      >
        <TriangleAlert class="w-4 h-4 mt-0.5 shrink-0" />
        <span>Bạn cần chọn <b>khu vực gửi hàng</b> bên dưới thì mới tạo được vận đơn.</span>
      </div>

      <BaseCard title="Hồ sơ đối tác">
        <div class="space-y-4">
          <div class="flex justify-between items-center text-sm border-b pb-3">
            <span class="text-meta">Email đăng nhập</span>
            <span class="text-body font-medium">{{ shop.email || '—' }}</span>
          </div>
          <div class="flex justify-between items-center text-sm border-b pb-3">
            <span class="text-meta">Trạng thái đối tác</span>
            <span
              class="inline-flex items-center gap-1.5 rounded-sm px-2.5 py-1 text-xs font-medium"
              :style="statusStyle(shop.status)"
            >
              <span class="h-1.5 w-1.5 rounded-full" style="background: currentColor"></span>
              {{ statusLabel(shop.status) }}
            </span>
          </div>

          <FormField v-model="form.name" label="Tên shop / Cửa hàng" placeholder="VD: Shop Thời Trang ABC" required />
          <FormField v-model="form.phone" label="Số điện thoại liên hệ" placeholder="VD: 0987654321" required />
          <LocationCascader
            v-model="form.location_id"
            label="Khu vực lấy hàng (Huyện/Xã)"
            hint="Để tối ưu việc tạo vận đơn và cước phí, hãy chọn đúng khu vực."
          />
          <FormField v-model="form.address_detail" label="Địa chỉ lấy hàng chi tiết (Số nhà, đường)" placeholder="VD: 123 Đường Lê Lợi, Phường 1" required />
          <div>
            <div class="text-sm font-medium mb-1">Định vị lấy hàng trên bản đồ</div>
            <div class="text-xs text-meta mb-3">Tùy chọn. Khi ghim đúng vị trí trên bản đồ, tài xế lấy hàng sẽ tìm đến nhanh chóng hơn.</div>
            <LocationPickerMap v-model="form.coordinates" />
          </div>
          <FormField
            v-model="form.webhook_url"
            label="Webhook URL (tùy chọn)"
            placeholder="https://shop.com/api/webhook"
            hint="Nhận thông báo tự động (HTTP POST) mỗi khi đơn hàng đổi trạng thái. Để trống nếu không dùng."
          />
        </div>
      </BaseCard>

      <!-- Khóa API & Tích hợp hệ thống -->
      <BaseCard title="Khóa API & Tích hợp hệ thống">
        <div class="space-y-5">
          <div>
            <div class="flex items-center justify-between gap-2 mb-1">
              <span class="text-sm font-semibold text-strong">Khóa xác thực API (X-API-Key)</span>
              <span v-if="shop.api_key" class="text-xs px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400 font-medium border border-emerald-200 dark:border-emerald-800">
                Đang hoạt động
              </span>
            </div>
            <p class="text-xs text-meta leading-relaxed">
              Dùng khóa này để tích hợp tự động tạo đơn và tính cước từ Website bán hàng (WooCommerce, Shopify, Haravan...), ERP hoặc phần mềm quản lý của Shop.
            </p>
          </div>

          <!-- Khung hiển thị API Key -->
          <div v-if="shop.api_key" class="space-y-3">
            <div class="flex items-center gap-2">
              <div class="relative flex-1">
                <input
                  :type="showApiKeyText ? 'text' : 'password'"
                  :value="shop.api_key"
                  readonly
                  class="w-full font-mono text-sm px-3.5 py-2.5 bg-subtle border border-slate-200 dark:border-slate-700 rounded-lg text-strong pr-10 select-all tracking-wider"
                />
                <button
                  type="button"
                  @click="showApiKeyText = !showApiKeyText"
                  class="absolute right-2.5 top-1/2 -translate-y-1/2 p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
                  :title="showApiKeyText ? 'Ẩn Key' : 'Hiện Key'"
                >
                  <EyeOff v-if="showApiKeyText" class="w-4 h-4" />
                  <Eye v-else class="w-4 h-4" />
                </button>
              </div>

              <BaseButton variant="secondary" size="md" @click="copyApiKey(shop.api_key)">
                <Check v-if="copiedActiveKey" class="w-4 h-4 text-emerald-600" />
                <Copy v-else class="w-4 h-4" />
                <span>{{ copiedActiveKey ? 'Đã chép' : 'Sao chép' }}</span>
              </BaseButton>

              <BaseButton variant="secondary" size="md" @click="showKeyModal = true">
                <Key class="w-4 h-4" />
                <span>Tạo lại</span>
              </BaseButton>
            </div>

            <!-- Hướng dẫn nhanh -->
            <div class="bg-slate-50 dark:bg-slate-900/60 p-4 rounded-xl border border-slate-200/80 dark:border-slate-800 text-xs text-slate-600 dark:text-slate-300 space-y-2">
              <div class="font-semibold text-slate-800 dark:text-slate-100 flex items-center gap-1.5">
                <Code class="w-3.5 h-3.5 text-sky-500" /> Cách gắn API Key vào Header:
              </div>
              <div class="font-mono bg-white dark:bg-slate-950 px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-800 select-all overflow-x-auto text-emerald-600 dark:text-emerald-400">
                X-API-Key: {{ shop.api_key }}
              </div>
              <div class="flex items-center justify-between pt-1">
                <span>Endpoint tạo đơn: <code class="text-sky-600 dark:text-sky-400">POST /api/v1/orders</code></span>
                <router-link :to="{ name: 'Docs' }" class="text-sky-600 hover:text-sky-700 font-medium hover:underline inline-flex items-center gap-1">
                  Xem tài liệu API đầy đủ &rarr;
                </router-link>
              </div>
            </div>
          </div>

          <!-- Chưa có key -->
          <div v-else class="p-4 rounded-xl border border-dashed text-center space-y-3 bg-subtle">
            <p class="text-sm text-meta">Tài khoản chưa có API Key. Bấm nút bên dưới để cấp khóa tích hợp ngay lập tức.</p>
            <BaseButton variant="primary" size="sm" @click="showKeyModal = true">
              <Key class="w-4 h-4" /> Tạo API Key ngay
            </BaseButton>
          </div>
        </div>
      </BaseCard>

      <div class="flex justify-end pt-2">
        <BaseButton variant="primary" size="lg" :loading="isSaving" @click="submit">
          <Save class="w-4 h-4" /> Lưu thông tin tài khoản
        </BaseButton>
      </div>
    </template>

    <div v-else class="py-16 text-center">
      <PackageX class="w-12 h-12 mx-auto text-meta/40 mb-3" />
      <p class="text-meta">Không tải được thông tin tài khoản.</p>
    </div>

    <!-- Modal tạo lại API Key -->
    <BaseModal v-model="showKeyModal" title="Cấp mới / Tạo lại API Key" :close-on-backdrop="false">
      <div v-if="!createdApiKey" class="space-y-4">
        <div class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-warning/10 text-[var(--warning)]">
          <TriangleAlert class="w-4 h-4 mt-0.5 shrink-0" />
          <span>Lưu ý: Khi tạo API Key mới, <b>Key cũ sẽ bị vô hiệu hóa ngay lập tức</b>. Hãy cập nhật Key mới vào hệ thống bán hàng của bạn.</span>
        </div>

        <div class="space-y-3">
          <FormField
            v-model="confirmPassword"
            type="password"
            label="Mật khẩu tài khoản của bạn"
            placeholder="Nhập mật khẩu đăng nhập để xác thực"
            required
          />

          <div class="pt-2">
            <div class="flex items-center justify-between mb-1.5">
              <span class="text-xs font-medium text-meta">Mã OTP xác thực qua Email (Tùy chọn)</span>
              <button
                type="button"
                :disabled="isRequestingOTP || otpCountdown > 0"
                @click="requestOTP"
                class="text-xs text-sky-600 hover:text-sky-700 font-medium hover:underline disabled:opacity-50"
              >
                <span v-if="otpCountdown > 0">Gửi lại sau {{ otpCountdown }}s</span>
                <span v-else>Gửi OTP vào {{ shop?.email }}</span>
              </button>
            </div>
            <FormField v-model="otpCode" placeholder="Nhập mã 6 số nếu bạn vừa yêu cầu OTP" maxlength="6" />
            <div v-if="otpSent" class="text-xs text-[var(--success)] flex items-center gap-1 mt-1">
              <Check class="w-3.5 h-3.5" /> Mã OTP đã được gửi vào hộp thư!
            </div>
          </div>
        </div>
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-success/10 text-[var(--success)]">
          <Check class="w-4 h-4 mt-0.5 shrink-0" />
          <span>API Key mới đã được tạo và kích hoạt thành công!</span>
        </div>
        <div>
          <label class="block text-xs font-medium text-meta mb-1">Khóa API của bạn:</label>
          <div class="bg-subtle border rounded-[var(--r-md)] p-3.5 font-mono text-sm text-strong break-all select-all font-semibold text-emerald-600 dark:text-emerald-400">
            {{ createdApiKey }}
          </div>
        </div>
        <p class="text-xs text-meta leading-relaxed">
          Hãy sao chép và lưu khóa này vào cấu hình kết nối API của bạn.
        </p>
      </div>

      <template #footer>
        <template v-if="!createdApiKey">
          <BaseButton variant="secondary" @click="closeKeyModal">Hủy bỏ</BaseButton>
          <BaseButton variant="primary" :loading="isRegenerating" :disabled="!confirmPassword" @click="regenerateApiKey">
            <Key class="w-4 h-4" /> Xác nhận tạo Key
          </BaseButton>
        </template>
        <template v-else>
          <BaseButton variant="secondary" @click="copyCreatedKey">
            <Check v-if="copied" class="w-4 h-4 text-emerald-600" />
            <Copy v-else class="w-4 h-4" />
            {{ copied ? 'Đã sao chép' : 'Sao chép Key' }}
          </BaseButton>
          <BaseButton variant="primary" @click="closeKeyModal">Đóng</BaseButton>
        </template>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { PackageX, TriangleAlert, Save, Key, Check, Copy, Eye, EyeOff, Code } from 'lucide-vue-next';
import BaseCard from '../../components/ui/BaseCard.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import BaseModal from '../../components/ui/BaseModal.vue';
import FormField from '../../components/ui/FormField.vue';
import LocationCascader from '../../components/ui/LocationCascader.vue';
import LocationPickerMap from '../../components/LocationPickerMap.vue';

const toast = useToastStore();

const shop = ref(null);
const loading = ref(true);
const isSaving = ref(false);
const showApiKeyText = ref(false);
const copiedActiveKey = ref(false);

const form = ref({
  name: '',
  phone: '',
  webhook_url: '',
  location_id: '',
  address_detail: '',
  coordinates: null
});

const showKeyModal = ref(false);
const isRegenerating = ref(false);
const isRequestingOTP = ref(false);
const otpCode = ref('');
const confirmPassword = ref('');
const createdApiKey = ref('');
const copied = ref(false);
const otpSent = ref(false);
const otpCountdown = ref(0);
let countdownTimer = null;

const fetchMe = async () => {
  loading.value = true;
  try {
    const meRes = await api.get('/shops/me');
    if (meRes.success) {
      shop.value = meRes.data;
      form.value = {
        name: meRes.data.name || '',
        phone: meRes.data.phone || '',
        location_id: meRes.data.location_id || '',
        address_detail: meRes.data.address_detail || '',
        webhook_url: meRes.data.webhook_url || '',
        coordinates: meRes.data.latitude && meRes.data.longitude 
          ? { lat: meRes.data.latitude, lng: meRes.data.longitude }
          : null
      };
    }
  } catch (error) {
    toast.error('Không thể tải thông tin tài khoản');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchMe);

const submit = async () => {
  if (!form.value.name || !form.value.address_detail || !form.value.phone) {
    toast.warning('Vui lòng điền tên shop, số điện thoại và địa chỉ lấy hàng');
    return;
  }
  isSaving.value = true;
  try {
    const payload = {
      name: form.value.name,
      phone: form.value.phone,
      webhook_url: form.value.webhook_url,
      location_id: form.value.location_id || null,
      address_detail: form.value.address_detail,
      latitude: form.value.coordinates?.lat || null,
      longitude: form.value.coordinates?.lng || null,
    };
    const res = await api.put('/shops/me', payload);
    if (res.success) {
      shop.value = { ...shop.value, ...res.data };
      toast.success('Đã lưu thông tin tài khoản thành công');
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Lưu thất bại');
  } finally {
    isSaving.value = false;
  }
};

const statusLabel = (s) => ({ pending: 'Chờ duyệt', approved: 'Đã duyệt & Kích hoạt', rejected: 'Bị từ chối' }[s] || s);
const statusStyle = (s) => {
  if (s === 'approved') return { color: 'var(--st-delivered-fg)', background: 'var(--st-delivered-bg)' };
  if (s === 'rejected') return { color: 'var(--st-returned-fg)', background: 'var(--st-returned-bg)' };
  return { color: 'var(--st-ready-fg)', background: 'var(--st-ready-bg)' };
};

const copyApiKey = async (key) => {
  if (!key) return;
  try {
    await navigator.clipboard.writeText(key);
    copiedActiveKey.value = true;
    toast.success('Đã sao chép API Key');
    setTimeout(() => { copiedActiveKey.value = false; }, 2500);
  } catch {
    toast.info('Vui lòng bôi đen và sao chép thủ công');
  }
};

const copyCreatedKey = async () => {
  if (!createdApiKey.value) return;
  try {
    await navigator.clipboard.writeText(createdApiKey.value);
    copied.value = true;
    toast.success('Đã sao chép API Key');
    setTimeout(() => { copied.value = false; }, 2500);
  } catch {
    toast.info('Vui lòng bôi đen và sao chép thủ công');
  }
};

const closeKeyModal = () => {
  showKeyModal.value = false;
  confirmPassword.value = '';
  otpCode.value = '';
  createdApiKey.value = '';
  copied.value = false;
  otpSent.value = false;
  otpCountdown.value = 0;
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
};

const requestOTP = async () => {
  isRequestingOTP.value = true;
  try {
    const res = await api.post('/shops/me/api-key/request-otp', {});
    if (res.success) {
      otpSent.value = true;
      otpCountdown.value = 60;
      countdownTimer = setInterval(() => {
        otpCountdown.value--;
        if (otpCountdown.value <= 0) { clearInterval(countdownTimer); countdownTimer = null; }
      }, 1000);
      toast.success('Mã OTP đã gửi vào email!');
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Gửi OTP thất bại');
  } finally {
    isRequestingOTP.value = false;
  }
};

const regenerateApiKey = async () => {
  if (!confirmPassword.value) {
    toast.warning('Vui lòng nhập mật khẩu xác nhận');
    return;
  }
  isRegenerating.value = true;
  try {
    const res = await api.post('/shops/me/api-key', { password: confirmPassword.value, otp: otpCode.value });
    if (res.success && res.data.api_key) {
      createdApiKey.value = res.data.api_key;
      if (shop.value) {
        shop.value.api_key = res.data.api_key;
      }
      toast.success('Đã tạo lại API Key thành công!');
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Xác thực mật khẩu thất bại');
  } finally {
    isRegenerating.value = false;
  }
};
</script>

