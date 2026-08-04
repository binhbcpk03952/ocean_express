package pdf

import (
	"bytes"
	"fmt"
	"ocean-express-api/internal/domain"

	"github.com/jung-kurt/gofpdf"
)

// GenerateOrderLabelPDF sinh ra tệp PDF tem vận đơn dạng byte array.
func GenerateOrderLabelPDF(order *domain.ShippingOrder) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A5", "")
	pdf.AddPage()

	// Cài đặt font Arial
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "OCEAN EXPRESS", "", 1, "C", false, 0, "")
	
	pdf.Ln(5)
	
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 15, order.TrackingNumber, "1", 1, "C", false, 0, "")
	
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "NGUOI NHAN:", "", 1, "L", false, 0, "")
	
	pdf.SetFont("Arial", "", 12)
	// Loại bỏ dấu tiếng Việt hoặc dùng font Unicode (ở đây dùng tiếng Anh không dấu cho đơn giản với gofpdf)
	pdf.CellFormat(0, 8, order.ReceiverName, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 8, order.ReceiverPhone, "", 1, "L", false, 0, "")
	
	// Cắt địa chỉ nếu quá dài
	addr := order.ReceiverAddressDetail
	if len(addr) > 50 {
		addr = addr[:50] + "..."
	}
	pdf.CellFormat(0, 8, addr, "", 1, "L", false, 0, "")
	
	pdf.Ln(5)
	
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, fmt.Sprintf("COD: %.0f VND", order.CodAmount), "1", 1, "C", false, 0, "")
	
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 8, fmt.Sprintf("Khoi luong: %d g", order.Weight), "", 1, "L", false, 0, "")
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}
