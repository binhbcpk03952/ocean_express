import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/authStore';

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('../views/ForgotPasswordView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('../views/ResetPasswordView.vue'),
    meta: { requiresAuth: false }
  },
  // Đăng nhập portal cho đối tác Shop (email + mật khẩu)
  {
    path: '/shop/login',
    name: 'ShopLogin',
    component: () => import('../views/shop/ShopLoginView.vue'),
    meta: { requiresAuth: false }
  },
  // Đăng ký tự phục vụ (công khai)
  {
    path: '/register/shop',
    name: 'RegisterShop',
    component: () => import('../views/register/RegisterShopView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register/shipper',
    name: 'RegisterShipper',
    component: () => import('../views/register/RegisterShipperView.vue'),
    meta: { requiresAuth: false }
  },
  // Tra cứu vận đơn công khai
  {
    path: '/tracking',
    name: 'PublicTracking',
    component: () => import('../views/PublicTrackingView.vue'),
    meta: { requiresAuth: false }
  },
  // Tài liệu Hướng dẫn sử dụng
  {
    path: '/docs',
    name: 'Docs',
    component: () => import('../views/DocsView.vue'),
    meta: { requiresAuth: false }
  },
  // Vùng Admin — chỉ role 'admin'
  {
    path: '/',
    component: () => import('../components/AdminLayout.vue'),
    meta: { requiresAuth: true, roles: ['admin'] },
    children: [
      { path: '', name: 'Dashboard', component: () => import('../views/DashboardView.vue') },
      { path: 'orders', name: 'Orders', component: () => import('../views/OrderListView.vue') },
      { path: 'orders/:id', name: 'Order Detail', component: () => import('../views/OrderDetailView.vue') },
      { path: 'locations', name: 'Locations & Hubs', component: () => import('../views/LocationHubView.vue') },
      { path: 'rates', name: 'Shipping Rates', component: () => import('../views/RateListView.vue') },
      { path: 'shops', name: 'Shops', component: () => import('../views/ShopListView.vue') },
      { path: 'employees', name: 'Employees', component: () => import('../views/EmployeeListView.vue') },
      { path: 'approvals', name: 'Approvals', component: () => import('../views/ApprovalsView.vue') },
      { path: 'settlements', name: 'Settlements', component: () => import('../views/SettlementsView.vue') },
    ]
  },
  // Vùng Member — các role vận hành (tài xế, nhân viên kho)
  {
    path: '/m',
    component: () => import('../components/MemberLayout.vue'),
    meta: { requiresAuth: true, roles: ['first_mile_driver', 'hub_staff', 'last_mile_driver'] },
    children: [
      { path: '', name: 'MemberTasks', component: () => import('../views/member/MemberTasksView.vue') },
      { path: 'orders/:id', name: 'MemberOrderDetail', component: () => import('../views/member/MemberOrderDetailView.vue') },
      { path: 'scan', name: 'HubScan', component: () => import('../views/member/HubScanView.vue'), meta: { roles: ['hub_staff'] } },
      { path: 'stats', name: 'MemberStats', component: () => import('../views/member/MemberStatsView.vue') },
    ]
  },
  // Vùng Shop portal — chỉ role 'shop'
  {
    path: '/shop',
    component: () => import('../components/ShopLayout.vue'),
    meta: { requiresAuth: true, roles: ['shop'] },
    children: [
      { path: '', name: 'ShopDashboard', component: () => import('../views/shop/ShopDashboardView.vue') },
      { path: 'orders/new', name: 'ShopCreateOrder', component: () => import('../views/shop/ShopCreateOrderView.vue') },
      { path: 'orders', name: 'ShopOrders', component: () => import('../views/shop/ShopOrdersView.vue') },
      { path: 'orders/:id', name: 'ShopOrderDetail', component: () => import('../views/shop/ShopOrderDetailView.vue') },
      { path: 'wallet', name: 'ShopWallet', component: () => import('../views/shop/ShopWalletView.vue') },
      { path: 'account', name: 'ShopAccount', component: () => import('../views/shop/ShopAccountView.vue') },
    ]
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach((to) => {
  const authStore = useAuthStore();

  if (to.meta.requiresAuth && !authStore.isAuthenticated()) {
    return { name: 'Login' };
  }
  // Trang không yêu cầu auth (login/đăng ký) mà đã đăng nhập -> về home đúng role.
  if (to.meta.requiresAuth === false && authStore.isAuthenticated()) {
    return authStore.homeRoute();
  }

  // Kiểm tra role: gộp roles của record cha + con trên đường match.
  // Nếu route có yêu cầu role mà user không thỏa -> đẩy về home đúng role của họ.
  if (authStore.isAuthenticated()) {
    const required = to.matched
      .map((r) => r.meta?.roles)
      .filter(Boolean);
    const denied = required.some((roles) => !authStore.hasRole(...roles));
    if (denied) {
      return authStore.homeRoute();
    }
  }

  return true;
});

export default router;
