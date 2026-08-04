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
          <svg class="h-5 w-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 18c1.8 0 1.8-1.5 3.6-1.5S8.4 18 10.2 18s1.8-1.5 3.6-1.5S15.6 18 17.4 18s1.8-1.5 3.6-1.5" />
            <path d="M4 15V6a1 1 0 0 1 1-1h9l6 6v4" />
            <path d="M14 5v5h5" />
          </svg>
        </div>
        <div class="min-w-0 leading-tight">
          <div class="truncate text-[15px] font-bold tracking-wide text-white">OCEAN</div>
          <div class="truncate text-[11px] font-medium tracking-[0.2em] text-teal-400">EXPRESS</div>
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
              :to="item.to"
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
            <p class="truncate text-sm font-semibold text-white">{{ authStore.user?.name || 'Người dùng' }}</p>
            <p class="truncate text-[11px] capitalize" style="color: var(--sidebar-text)">{{ roleLabel }}</p>
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
        <button
          @click="handleLogout"
          class="inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium text-body transition-colors hover:bg-subtle"
        >
          <LogOut class="h-4 w-4" />
          <span class="hidden sm:inline">Đăng xuất</span>
        </button>
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
import {
  LayoutDashboard, Package, MapPin, DollarSign, Store, Users, Menu, LogOut, ClipboardCheck, Wallet,
} from 'lucide-vue-next';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const sidebarOpen = ref(false);

const isAdmin = computed(() => authStore.user?.role === 'admin');

const navGroups = computed(() => {
  const groups = [
    {
      label: 'Vận hành',
      items: [
        { to: '/', label: 'Tổng quan', icon: LayoutDashboard },
        { to: '/orders', label: 'Vận đơn', icon: Package },
      ],
    },
    {
      label: 'Cấu hình',
      items: [
        { to: '/locations', label: 'Khu vực & Bưu cục', icon: MapPin },
        { to: '/rates', label: 'Bảng giá cước', icon: DollarSign },
      ],
    },
  ];
  if (isAdmin.value) {
    groups.push({
      label: 'Quản trị',
      items: [
        { to: '/approvals', label: 'Duyệt tài khoản', icon: ClipboardCheck },
        { to: '/settlements', label: 'Đối soát & Chi trả', icon: Wallet },
        { to: '/shops', label: 'Đối tác (Shops)', icon: Store },
        { to: '/employees', label: 'Nhân sự', icon: Users },
      ],
    });
  }
  return groups;
});

const isActive = (to) => {
  if (to === '/') return route.path === '/';
  return route.path.startsWith(to);
};

const titleMap = {
  '/': ['Tổng quan', 'Bức tranh vận hành toàn hệ thống'],
  '/orders': ['Vận đơn', 'Quản lý và theo dõi đơn hàng'],
  '/locations': ['Khu vực & Bưu cục', 'Sơ đồ hành chính và mạng lưới hub'],
  '/rates': ['Bảng giá cước', 'Cấu hình tính cước tự động'],
  '/shops': ['Đối tác Shops', 'Quản lý đối tác E-commerce'],
  '/employees': ['Nhân sự', 'Tài khoản và phân quyền nhân viên'],
  '/approvals': ['Duyệt tài khoản', 'Phê duyệt đối tác và tài xế đăng ký mới'],
  '/settlements': ['Đối soát COD', 'Chốt và chi trả tiền thu hộ cho đối tác'],
};

const routeMeta = computed(() => {
  const key = Object.keys(titleMap).find((k) => isActive(k) && (k !== '/' || route.path === '/'));
  return titleMap[route.path] || titleMap[key] || ['Ocean Express', ''];
});
const routeTitle = computed(() => routeMeta.value[0]);
const routeSubtitle = computed(() => routeMeta.value[1]);

const roleLabels = {
  admin: 'Quản trị viên',
  hub_staff: 'Nhân viên kho',
  first_mile_driver: 'Tài xế lấy hàng',
  last_mile_driver: 'Tài xế giao hàng',
};
const roleLabel = computed(() => roleLabels[authStore.user?.role] || authStore.user?.role || '');

const initials = computed(() => {
  const name = authStore.user?.name || 'U';
  return name.trim().split(/\s+/).slice(-2).map((w) => w[0]).join('').toUpperCase();
});

const handleLogout = () => {
  authStore.logout();
  router.push('/login');
};
</script>

<style scoped>
.page-enter-active, .page-leave-active { transition: opacity 0.15s ease, transform 0.15s ease; }
.page-enter-from { opacity: 0; transform: translateY(6px); }
.page-leave-to { opacity: 0; transform: translateY(-6px); }
</style>
