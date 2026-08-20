import { ref } from 'vue';

export function usePdfPrint() {
  const printing = ref(false);

  const triggerBlobDownload = (blob, filename) => {
    const blobUrl = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = blobUrl;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    setTimeout(() => {
      document.body.removeChild(a);
      URL.revokeObjectURL(blobUrl);
    }, 1000);
  };

  const downloadOrderPDF = async (orderId) => {
    if (!orderId) return;
    printing.value = true;
    try {
      const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';
      const token = localStorage.getItem('ocean_token');
      const headers = {};
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const res = await fetch(`${baseURL}/orders/${orderId}/pdf`, { headers });
      if (!res.ok) {
        // Fallback thử qua /public/orders/:id/pdf
        const pubRes = await fetch(`${baseURL}/public/orders/${orderId}/pdf`);
        if (!pubRes.ok) {
          throw new Error('Lỗi khi tải file PDF vận đơn');
        }
        const blob = await pubRes.blob();
        triggerBlobDownload(blob, `label_${orderId}.pdf`);
        return;
      }

      const blob = await res.blob();
      triggerBlobDownload(blob, `label_${orderId}.pdf`);
    } catch (err) {
      console.error('Lỗi tải PDF:', err);
      throw err;
    } finally {
      printing.value = false;
    }
  };

  const downloadBatchPDF = async (orderIds) => {
    if (!orderIds || orderIds.length === 0) return;
    printing.value = true;
    try {
      const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';
      const token = localStorage.getItem('ocean_token');
      const headers = {
        'Content-Type': 'application/json'
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const res = await fetch(`${baseURL}/orders/labels/batch`, {
        method: 'POST',
        headers,
        body: JSON.stringify({ order_ids: orderIds })
      });

      if (!res.ok) {
        throw new Error('Lỗi khi tạo file in hàng loạt');
      }

      const blob = await res.blob();
      triggerBlobDownload(blob, `batch_labels_${new Date().toISOString().slice(0, 10)}.pdf`);
    } catch (err) {
      console.error('Lỗi in hàng loạt:', err);
      throw err;
    } finally {
      printing.value = false;
    }
  };

  const printOrderPDF = async (orderId) => {
    return downloadOrderPDF(orderId);
  };

  return {
    printing,
    printOrderPDF,
    downloadOrderPDF,
    downloadBatchPDF
  };
}
