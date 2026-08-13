import { ref } from 'vue';

export function usePdfPrint() {
  const printing = ref(false);

  const printOrderPDF = async (orderId) => {
    if (!orderId) return;
    printing.value = true;
    try {
      const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
      const token = localStorage.getItem('ocean_token');
      const res = await fetch(`${baseURL}/orders/${orderId}/pdf`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      if (!res.ok) {
        throw new Error('Lỗi khi tải file PDF vận đơn');
      }
      const blob = await res.blob();
      const blobUrl = URL.createObjectURL(blob);
      window.open(blobUrl, '_blank');
    } catch (err) {
      console.error('Lỗi in PDF:', err);
      alert('Không thể tải file PDF vận đơn');
    } finally {
      printing.value = false;
    }
  };

  return {
    printing,
    printOrderPDF
  };
}
