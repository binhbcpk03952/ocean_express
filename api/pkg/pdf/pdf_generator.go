package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"ocean-express-api/internal/domain"

	"github.com/signintech/gopdf"
	"github.com/skip2/go-qrcode"
)

//go:embed fonts/Roboto-Regular.ttf
var robotoRegular []byte

//go:embed fonts/Roboto-Bold.ttf
var robotoBold []byte

const (
	fontRegular = "Roboto"
	fontBold    = "Roboto-Bold"

	// A5 là khổ giấy in (148 x 210 mm).
	pageW = 419.53
	pageH = 595.28

	margin = 20.0 // ~7mm viền ngoài
	padX   = 8.0  // đệm trong giữa viền và chữ

	textX = margin + padX
	textW = pageW - 2*(margin+padX)

	borderRight = pageW - margin

	// Dấu thanh tiếng Việt cần nhiều khoảng trống phía trên hơn chữ Latin.
	lineFactor = 1.5

	signatureH = 70.0 // chiều cao khối ký nhận neo ở đáy tem
)

// newMeasuringPDF tạo một tài liệu A6 đã nhúng sẵn font, dùng để dựng tem
// hoặc chỉ để đo chiều rộng chữ.
func newMeasuringPDF() (*gopdf.GoPdf, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: pageW, H: pageH}})
	pdf.AddPage()

	if err := pdf.AddTTFFontData(fontRegular, robotoRegular); err != nil {
		return nil, fmt.Errorf("nhúng font Roboto Regular: %w", err)
	}
	if err := pdf.AddTTFFontData(fontBold, robotoBold); err != nil {
		return nil, fmt.Errorf("nhúng font Roboto Bold: %w", err)
	}
	return pdf, nil
}

// GenerateOrderLabelPDF sinh ra tệp PDF tem vận đơn dạng byte array.
func GenerateOrderLabelPDF(order *domain.ShippingOrder, shop *domain.Shop) ([]byte, error) {
	pdf, err := newMeasuringPDF()
	if err != nil {
		return nil, err
	}

	drawOrderLabel(pdf, order, shop)

	var buf bytes.Buffer
	if _, err := pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("ghi PDF: %w", err)
	}
	return buf.Bytes(), nil
}

// GenerateBatchOrderLabelsPDF sinh ra tệp PDF gồm nhiều tem vận đơn ghép trang.
func GenerateBatchOrderLabelsPDF(orders []*domain.ShippingOrder, shopMap map[string]*domain.Shop) ([]byte, error) {
	if len(orders) == 0 {
		return nil, fmt.Errorf("danh sách vận đơn trống")
	}

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: pageW, H: pageH}})

	if err := pdf.AddTTFFontData(fontRegular, robotoRegular); err != nil {
		return nil, fmt.Errorf("nhúng font Roboto Regular: %w", err)
	}
	if err := pdf.AddTTFFontData(fontBold, robotoBold); err != nil {
		return nil, fmt.Errorf("nhúng font Roboto Bold: %w", err)
	}

	for i, order := range orders {
		if order == nil {
			continue
		}
		if i > 0 {
			pdf.AddPage()
		} else {
			pdf.AddPage()
		}

		var shop *domain.Shop
		if shopMap != nil {
			shop = shopMap[order.ShopID]
		}
		drawOrderLabel(pdf, order, shop)
	}

	var buf bytes.Buffer
	if _, err := pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("ghi PDF batch: %w", err)
	}
	return buf.Bytes(), nil
}

func drawOrderLabel(pdf *gopdf.GoPdf, order *domain.ShippingOrder, shop *domain.Shop) {
	l := &layout{pdf: pdf}

	// Viền ngoài
	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeftWithStyle(margin, margin, pageW-2*margin, pageH-2*margin, "D")

	qrSize := 80.0
	// Thêm QR Code góc trên phải (chứa Tracking Number để tiện quét mã kho)
	if order.TrackingNumber != "" {
		png, err := qrcode.Encode(order.TrackingNumber, qrcode.Medium, 256)
		if err == nil {
			imgH, err := gopdf.ImageHolderByReader(bytes.NewReader(png))
			if err == nil {
				pdf.ImageByHolder(imgH, pageW-margin-qrSize-5, margin+5, &gopdf.Rect{W: qrSize, H: qrSize})
			}
		}
	}

	// ----- Header -----
	l.y = margin + 10
	headerW := textW - qrSize - 10

	pdf.SetFont(fontBold, "", 24)
	pdf.SetXY(textX, l.y)
	pdf.CellWithOption(&gopdf.Rect{W: headerW, H: 24 * lineFactor}, "OCEAN EXPRESS", gopdf.CellOption{Align: gopdf.Center})
	l.y += 24 * lineFactor

	pdf.SetFont(fontBold, "", 20)
	pdf.SetXY(textX, l.y)
	pdf.CellWithOption(&gopdf.Rect{W: headerW, H: 20 * lineFactor}, order.TrackingNumber, gopdf.CellOption{Align: gopdf.Center})
	l.y += 20 * lineFactor

	// Đảm bảo không bị lẹm viền vào QR Code
	qrBottom := margin + 5 + qrSize + 5
	if l.y < qrBottom {
		l.y = qrBottom
	}

	l.divider(6)

	// ----- Người gửi -----
	l.label(fontBold, 14, "Từ (Người gửi):")

	senderStr := "Ocean Express Shop"
	if shop != nil && shop.Name != "" {
		senderStr = shop.Name
	}
	if order.SenderPhone != nil && *order.SenderPhone != "" {
		senderStr += " - " + *order.SenderPhone
	} else if shop != nil && shop.Phone != nil && *shop.Phone != "" {
		senderStr += " - " + *shop.Phone
	}
	l.paragraph(fontBold, 16, senderStr, 2)
	senderAddr := strings.TrimSpace(order.SenderAddressDetail)
	if senderAddr == "" && shop != nil {
		senderAddr = strings.TrimSpace(shop.AddressDetail)
	}
	l.paragraph(fontRegular, 14, senderAddr, 3)

	l.divider(6)

	// ----- Người nhận -----
	l.label(fontBold, 14, "Đến (Người nhận):")

	receiver := strings.TrimSpace(order.ReceiverName)
	if phone := strings.TrimSpace(order.ReceiverPhone); phone != "" {
		if receiver != "" {
			receiver += " - " + phone
		} else {
			receiver = phone
		}
	}
	l.paragraph(fontBold, 16, receiver, 2)
	l.paragraph(fontRegular, 14, strings.TrimSpace(order.ReceiverAddressDetail), 3)

	l.divider(6)

	// ----- Khối lượng & COD trên cùng một hàng -----
	rowY := l.y
	pdf.SetFont(fontBold, "", 16)
	pdf.SetXY(textX, rowY)
	pdf.Cell(nil, fmt.Sprintf("Khối lượng: %s g", formatThousands(int64(order.Weight))))

	pdf.SetFont(fontBold, "", 16)
	pdf.SetXY(textX, rowY)
	codW := textW
	pdf.CellWithOption(
		&gopdf.Rect{W: codW, H: 16 * lineFactor},
		fmt.Sprintf("COD: %s đ", formatThousands(int64(order.CodAmount))),
		gopdf.CellOption{Align: gopdf.Right},
	)
	l.y = rowY + 16*lineFactor

	l.divider(6)

	// ----- Ghi chú -----
	l.label(fontBold, 14, "Ghi chú:")
	l.paragraph(fontRegular, 14,
		"Cho xem hàng, không thử. Quý khách vui lòng quay video khi mở kiện hàng để được hỗ trợ tốt nhất.", 3)

	// ----- Ký nhận (neo ở đáy tem, không phụ thuộc nội dung phía trên) -----
	sigY := pageH - margin - signatureH
	pdf.Line(margin, sigY, borderRight, sigY)

	colW := textW / 2
	pdf.SetFont(fontBold, "", 14)
	pdf.SetXY(textX, sigY+10)
	pdf.CellWithOption(&gopdf.Rect{W: colW, H: 18}, "Chữ ký người nhận", gopdf.CellOption{Align: gopdf.Center})
	pdf.SetXY(textX+colW, sigY+10)
	pdf.CellWithOption(&gopdf.Rect{W: colW, H: 18}, "Chữ ký bưu tá", gopdf.CellOption{Align: gopdf.Center})
}

// layout giữ con trỏ dọc để các trường trôi tự nhiên theo nội dung,
// tránh việc toạ độ cứng bị đè lên nhau khi text dài phải xuống dòng.
type layout struct {
	pdf *gopdf.GoPdf
	y   float64
}

func (l *layout) centered(font string, size float64, text string) {
	if text == "" {
		return
	}
	l.pdf.SetFont(font, "", size)
	l.pdf.SetXY(textX, l.y)
	l.pdf.CellWithOption(&gopdf.Rect{W: textW, H: size * lineFactor}, text,
		gopdf.CellOption{Align: gopdf.Center | gopdf.Middle})
	l.y += size * lineFactor
}

func (l *layout) label(font string, size float64, text string) {
	l.pdf.SetFont(font, "", size)
	l.pdf.SetXY(textX, l.y)
	l.pdf.Cell(nil, text)
	l.y += size * lineFactor
}

// paragraph xuống dòng theo ranh giới từ và cắt bớt khi vượt maxLines.
func (l *layout) paragraph(font string, size float64, text string, maxLines int) {
	if strings.TrimSpace(text) == "" {
		return
	}
	l.pdf.SetFont(font, "", size)
	for _, line := range wrapText(l.pdf, text, textW, maxLines) {
		l.pdf.SetXY(textX, l.y)
		l.pdf.Cell(nil, line)
		l.y += size * lineFactor
	}
}

func (l *layout) divider(gap float64) {
	l.y += gap
	l.pdf.Line(margin, l.y, borderRight, l.y)
	l.y += gap + 2
}

// wrapText ngắt dòng theo từ (gopdf.SplitText cắt giữa từ nên không dùng được).
// Từ nào dài hơn cả dòng thì mới cắt cứng theo ký tự.
func wrapText(pdf *gopdf.GoPdf, text string, maxW float64, maxLines int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	cur := ""
	for _, w := range words {
		candidate := w
		if cur != "" {
			candidate = cur + " " + w
		}
		if textWidth(pdf, candidate) <= maxW {
			cur = candidate
			continue
		}
		if cur != "" {
			lines = append(lines, cur)
			cur = ""
		}
		// Từ đơn dài hơn một dòng: cắt cứng theo rune.
		for textWidth(pdf, w) > maxW {
			runes := []rune(w)
			cut := len(runes)
			for cut > 1 && textWidth(pdf, string(runes[:cut])) > maxW {
				cut--
			}
			lines = append(lines, string(runes[:cut]))
			w = string(runes[cut:])
		}
		cur = w
	}
	if cur != "" {
		lines = append(lines, cur)
	}

	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
		last := lines[maxLines-1]
		for last != "" && textWidth(pdf, last+"…") > maxW {
			runes := []rune(last)
			last = string(runes[:len(runes)-1])
		}
		lines[maxLines-1] = strings.TrimRight(last, " ") + "…"
	}
	return lines
}

func textWidth(pdf *gopdf.GoPdf, s string) float64 {
	w, err := pdf.MeasureTextWidth(s)
	if err != nil {
		return 0
	}
	return w
}

// joinWords ghép các dòng đã ngắt lại thành một chuỗi, dùng để kiểm tra
// việc ngắt dòng không làm mất hoặc cắt vụn từ.
func joinWords(lines []string) string {
	return strings.Join(lines, " ")
}

// formatThousands định dạng số theo kiểu Việt Nam: 250000 -> "250.000".
func formatThousands(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}
