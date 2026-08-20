import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '../services/api';

export const useNotificationStore = defineStore('notification', () => {
    const notifications = ref([]);
    const unreadCount = ref(0);
    const loading = ref(false);

    const fetchNotifications = async () => {
        loading.value = true;
        try {
            const res = await api.get('/notifications?limit=20');
            notifications.value = res.data || []; // fallback to empty array if no data
        } catch (error) {
            console.error('Failed to fetch notifications', error);
        } finally {
            loading.value = false;
        }
    };

    const fetchUnreadCount = async () => {
        try {
            const res = await api.get('/notifications/unread-count');
            unreadCount.value = res.data?.count || 0; 
        } catch (error) {
            console.error('Failed to fetch unread count', error);
        }
    };

    const markAsRead = async (id) => {
        try {
            await api.put(`/notifications/${id}/read`);
            const notif = notifications.value.find(n => n.id === id);
            if (notif && !notif.is_read) {
                notif.is_read = true;
                unreadCount.value = Math.max(0, unreadCount.value - 1);
            }
        } catch (error) {
            console.error('Failed to mark as read', error);
        }
    };

    const markAllAsRead = async () => {
        try {
            await api.put('/notifications/read-all');
            notifications.value.forEach(n => n.is_read = true);
            unreadCount.value = 0;
        } catch (error) {
            console.error('Failed to mark all as read', error);
        }
    };

    return {
        notifications,
        unreadCount,
        loading,
        fetchNotifications,
        fetchUnreadCount,
        markAsRead,
        markAllAsRead
    };
});
