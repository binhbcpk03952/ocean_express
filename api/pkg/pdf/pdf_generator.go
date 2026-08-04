package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"ocean-express-api/internal/domain"

	"github.com/jung-kurt/gofpdf"
)

//go:embed fonts/Roboto-Regular.ttf
var robotoRegular []byte

//go:embed fonts/Roboto-Bold.ttf
var robotoBold []byte

// GenerateOrderLabelPDF sinh ra tệp PDF tem vận đơn dạng byte array.
func GenerateOrderLabelPDF(order *domain.ShippingOrder) ([]byte, error) {
	// A6 size is standard for shipping labels (105 x 148 mm)
	pdf := gofpdf.New("P", "mm", "A6", "")
	pdf.SetMargins(5, 5, 5)
	pdf.AddPage()

	// Đăng ký font UTF-8 bằng file được embed sẵn trong binary
	pdf.AddUTF8FontFromBytes("Roboto", "", robotoRegular)
	pdf.AddUTF8FontFromBytes("Roboto", "B", robotoBold)

	// Vẽ viền bao quanh toàn bộ tem
	pdf.Rect(5, 5, 95, 138, "D")

	// Header - Tên Đơn vị vận chuyển
	pdf.SetFont("Roboto", "B", 18)
	pdf.CellFormat(95, 12, "OCEAN EXPRESS", "B", 1, "C", false, 0, "")

	// Mã vận đơn (Tracking Number)
	pdf.Ln(4)
	pdf.SetFont("Roboto", "B", 16)
	pdf.CellFormat(95, 10, order.TrackingNumber, "", 1, "C", false, 0, "")
	
	// Giả lập vạch Barcode bằng text
	pdf.SetFont("Roboto", "", 11)
	pdf.CellFormat(95, 5, "* " + order.TrackingNumber + " *", "", 1, "C", false, 0, "")
	pdf.Ln(2)

	// Đường gạch ngang
	pdf.Line(5, pdf.GetY(), 100, pdf.GetY())
	pdf.Ln(3)

	// SENDER INFO (Thông tin người gửi)
	pdf.SetFont("Roboto", "B", 10)
	pdf.CellFormat(95, 5, "Từ (Người gửi):", "", 1, "L", false, 0, "")
	pdf.SetFont("Roboto", "", 10)
	pdf.CellFormat(95, 5, "Cửa hàng / Đối tác Ocean Express", "", 1, "L", false, 0, "")
	
	pdf.Ln(3)
	pdf.Line(5, pdf.GetY(), 100, pdf.GetY())
	pdf.Ln(3)

	// RECEIVER INFO (Thông tin người nhận)
	pdf.SetFont("Roboto", "B", 12)
	pdf.CellFormat(95, 6, "Đến (Người nhận):", "", 1, "L", false, 0, "")
	
	pdf.SetFont("Roboto", "B", 14)
	pdf.CellFormat(95, 7, order.ReceiverName + " - " + order.ReceiverPhone, "", 1, "L", false, 0, "")
	
	pdf.SetFont("Roboto", "", 11)
	pdf.MultiCell(95, 5, order.ReceiverAddressDetail, "", "L", false)
	
	pdf.Ln(3)
	pdf.Line(5, pdf.GetY(), 100, pdf.GetY())
	pdf.Ln(3)

	// CHI TIẾT ĐƠN HÀNG (Weight & COD)
	pdf.SetFont("Roboto", "B", 12)
	pdf.CellFormat(47, 8, fmt.Sprintf("Khối lượng: %d g", order.Weight), "R", 0, "L", false, 0, "")
	
	pdf.SetFont("Roboto", "B", 14)
	pdf.CellFormat(48, 8, fmt.Sprintf("COD: %.0f đ", order.CodAmount), "", 1, "R", false, 0, "")

	// GHI CHÚ
	pdf.Line(5, pdf.GetY(), 100, pdf.GetY())
	pdf.Ln(3)
	pdf.SetFont("Roboto", "B", 10)
	pdf.CellFormat(95, 5, "Ghi chú:", "", 1, "L", false, 0, "")
	pdf.SetFont("Roboto", "", 10)
	pdf.MultiCell(95, 5, "Cho xem hàng, không thử. Quý khách vui lòng quay video khi mở kiện hàng để được hỗ trợ tốt nhất.", "", "L", false)

	// Ký nhận
	pdf.Line(5, pdf.GetY()+5, 100, pdf.GetY()+5)
	pdf.Ln(8)
	pdf.SetFont("Roboto", "B", 10)
	pdf.CellFormat(47, 5, "Chữ ký người nhận", "", 0, "C", false, 0, "")
	pdf.CellFormat(48, 5, "Chữ ký bưu tá", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}
