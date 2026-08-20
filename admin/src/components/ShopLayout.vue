<template>
  <div class="flex min-h-screen bg-base text-body">
    <!-- Sidebar -->
    <aside
      class="fixed inset-y-0 left-0 z-30 flex w-[260px] flex-col transition-transform duration-200 lg:translate-x-0"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
      style="background: var(--sidebar-bg)"
    >
      <!-- Brand -->
      <div class="flex h-16 shrink-0 items-center gap-3 px-5" style="background: var(--sidebar-bg-deep)">
        <div
          class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md"
          style="background: var(--brand-gradient)"
        >
          <Store class="h-5 w-5 text-white" />
        </div>
        <div class="min-w-0 leading-tight">
          <div class="truncate text-[15px] font-bold tracking-wide text-white">CỔNG ĐỐI TÁC</div>
          <div class="truncate text-[11px] font-medium tracking-[0.2em] text-teal-400">OCEAN EXPRESS</div>
        </div>
      </div>

      <!-- Nav -->
      <nav class="flex-1 space-y-6 overflow-y-auto px-3 py-5">
        <div v-for="group in navGroups" :key="group.label">
          <p class="px-3 pb-2 text-[10px] font-semibold uppercase tracking-[0.12em]" style="color: var(--sidebar-text); opacity: 0.6">
            {{ group.label }}
          </p>
          <div class="space-y-1">
            <router-link
              v-for="item in group.items"
              :key="item.to"
              :to="{ name: item.to }"
              class="group relative flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-colors"
              :class="isActive(item.to)
                ? 'text-white'
                : 'hover:bg-white/5'"
              :style="isActive(item.to)
                ? 'background: var(--sidebar-active-bg)'
                : 'color: var(--sidebar-text)'"
            >
              <span
                v-if="isActive(item.to)"
                class="absolute left-0 top-1/2 h-6 -translate-y-1/2 rounded-r-full"
                style="width: 3px; background: var(--teal-400)"
              />
              <component :is="item.icon" class="h-[18px] w-[18px] shrink-0" :style="isActive(item.to) ? 'color: var(--teal-400)' : ''" />
              <span class="truncate">{{ item.label }}</span>
            </router-link>
          </div>
        </div>
      </nav>

      <!-- Footer: user + theme -->
      <div class="shrink-0 border-t p-3" style="border-color: rgba(255,255,255,0.08)">
        <div class="flex items-center gap-3 rounded-md px-2 py-2">
          <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-sm font-semibold text-white" style="background: var(--navy-700)">
            {{ initials }}
          </div>
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-semibold text-white">{{ authStore.user?.name || 'Đối tác' }}</p>
            <p class="truncate text-[11px] capitalize" style="color: var(--sidebar-text)">Shop</p>
          </div>
          <ThemeToggle />
        </div>
      </div>
    </aside>

    <!-- Mobile overlay -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-20 bg-black/50 lg:hidden"
      @click="sidebarOpen = false"
    />

    <!-- Main -->
    <div class="flex min-h-screen flex-1 flex-col lg:pl-[260px]">
      <header class="sticky top-0 z-10 flex h-16 items-center justify-between border-b bg-surface/80 px-4 backdrop-blur-md sm:px-6">
        <div class="flex items-center gap-3">
          <button class="rounded-md p-2 text-meta hover:bg-subtle lg:hidden" @click="sidebarOpen = true">
            <Menu class="h-5 w-5" />
          </button>
          <div>
            <h1 class="text-[17px] font-semibold text-strong">{{ routeTitle }}</h1>
            <p class="text-[12px] text-meta">{{ routeSubtitle }}</p>
          </div>
        </div>
        <div class="flex items-center gap-4">
          <NotificationBell />
          <button
            @click="handleLogout"
            class="inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium text-body transition-colors hover:bg-subtle"
          >
            <LogOut class="h-4 w-4" />
            <span class="hidden sm:inline">Đăng xuất</span>
          </button>
        </div>
      </header>

      <main class="flex-1 p-4 sm:p-6">
        <router-view v-slot="{ Component }">
          <transition name="page" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '../stores/authStore';
import ThemeToggle from './ui/ThemeToggle.vue';
import NotificationBell from './NotificationBell.vue';
import {
  Store, LayoutDashboard, PackagePlus, Package, Wallet, UserCircle, LogOut, Menu
} from 'lucide-vue-next';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const sidebarOpen = ref(false);

const navGroups = [
  {
    label: 'Vận hành',
    items: [
      { to: 'ShopDashboard', label: 'Tổng quan', icon: LayoutDashboard },
      { to: 'ShopCreateOrder', label: 'Tạo đơn mới', icon: PackagePlus },
      { to: 'ShopOrders', label: 'Quản lý Đơn hàng', icon: Package },
    ],
  },
  {
    label: 'Tài chính',
    items: [
      { to: 'ShopWallet', label: 'Ví & Đối soát', icon: Wallet },
    ],
  },
  {
    label: 'Cấu hình',
    items: [
      { to: 'ShopAccount', label: 'Tài khoản & API', icon: UserCircle },
    ],
  }
];

const isActive = (to) => route.name === to;

const titleMap = {
  'ShopDashboard': ['Tổng quan', 'Tình hình vận đơn của cửa hàng'],
  'ShopCreateOrder': ['Tạo vận đơn', 'Gửi hàng nhanh chóng'],
  'ShopOrders': ['Quản lý đơn hàng', 'Theo dõi trạng thái và lịch sử'],
  'ShopOrderDetail': ['Chi tiết vận đơn', 'Xem lộ trình GPS và in tem vận đơn'],
  'ShopWallet': ['Ví & Đối soát', 'Số dư và đối soát tiền thu hộ COD'],
  'ShopAccount': ['Tài khoản & API', 'Hồ sơ, địa chỉ lấy hàng và tích hợp'],
};

const routeMeta = computed(() => {
  return titleMap[route.name] || ['Ocean Express Shop', ''];
});
const routeTitle = computed(() => routeMeta.value[0]);
const routeSubtitle = computed(() => routeMeta.value[1]);

const initials = computed(() => {
  const name = authStore.user?.name || 'S';
  return name.trim().split(/\s+/).slice(-2).map((w) => w[0]).join('').toUpperCase();
});

const handleLogout = () => {
  authStore.logout();
  router.push({ name: 'ShopLogin' });
};
</script>

<style scoped>
.page-enter-active, .page-leave-active { transition: opacity 0.15s ease, transform 0.15s ease; }
.page-enter-from { opacity: 0; transform: translateY(6px); }
.page-leave-to { opacity: 0; transform: translateY(-6px); }
</style>