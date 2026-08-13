package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"ocean-express-api/internal/domain"

	"github.com/signintech/gopdf"
)

//go:embed fonts/Roboto-Regular.ttf
var robotoRegular []byte

//go:embed fonts/Roboto-Bold.ttf
var robotoBold []byte

const (
	fontRegular = "Roboto"
	fontBold    = "Roboto-Bold"

	// A6 là khổ tem vận đơn tiêu chuẩn (105 x 148 mm).
	pageW = 297.64
	pageH = 419.53

	margin = 14.0 // ~5mm viền ngoài
	padX   = 6.0  // đệm trong giữa viền và chữ

	textX = margin + padX
	textW = pageW - 2*(margin+padX)

	borderRight = pageW - margin

	// Dấu thanh tiếng Việt cần nhiều khoảng trống phía trên hơn chữ Latin.
	lineFactor = 1.5

	signatureH = 46.0 // chiều cao khối ký nhận neo ở đáy tem
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
func GenerateOrderLabelPDF(order *domain.ShippingOrder) ([]byte, error) {
	pdf, err := newMeasuringPDF()
	if err != nil {
		return nil, err
	}

	l := &layout{pdf: pdf}

	// Viền ngoài
	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeftWithStyle(margin, margin, pageW-2*margin, pageH-2*margin, "D")

	// ----- Header -----
	l.y = margin + 6
	l.centered(fontBold, 18, "OCEAN EXPRESS")
	l.centered(fontBold, 16, order.TrackingNumber)
	// Giả lập vạch barcode bằng text
	l.centered(fontRegular, 11, "* "+order.TrackingNumber+" *")

	l.divider(4)

	// ----- Người gửi -----
	l.label(fontBold, 10, "Từ (Người gửi):")
	l.paragraph(fontRegular, 10, "Cửa hàng / Đối tác Ocean Express", 1)

	l.divider(4)

	// ----- Người nhận -----
	l.label(fontBold, 11, "Đến (Người nhận):")

	receiver := strings.TrimSpace(order.ReceiverName)
	if phone := strings.TrimSpace(order.ReceiverPhone); phone != "" {
		if receiver != "" {
			receiver += " - " + phone
		} else {
			receiver = phone
		}
	}
	l.paragraph(fontBold, 13, receiver, 2)
	l.paragraph(fontRegular, 10, strings.TrimSpace(order.ReceiverAddressDetail), 3)

	l.divider(4)

	// ----- Khối lượng & COD trên cùng một hàng -----
	rowY := l.y
	pdf.SetFont(fontBold, "", 11)
	pdf.SetXY(textX, rowY)
	pdf.Cell(nil, fmt.Sprintf("Khối lượng: %s g", formatThousands(int64(order.Weight))))

	pdf.SetFont(fontBold, "", 13)
	pdf.SetXY(textX, rowY)
	codW := textW
	pdf.CellWithOption(
		&gopdf.Rect{W: codW, H: 16},
		fmt.Sprintf("COD: %s đ", formatThousands(int64(order.CodAmount))),
		gopdf.CellOption{Align: gopdf.Right},
	)
	l.y = rowY + 13*lineFactor

	l.divider(4)

	// ----- Ghi chú -----
	l.label(fontBold, 10, "Ghi chú:")
	l.paragraph(fontRegular, 10,
		"Cho xem hàng, không thử. Quý khách vui lòng quay video khi mở kiện hàng để được hỗ trợ tốt nhất.", 3)

	// ----- Ký nhận (neo ở đáy tem, không phụ thuộc nội dung phía trên) -----
	sigY := pageH - margin - signatureH
	pdf.Line(margin, sigY, borderRight, sigY)

	colW := textW / 2
	pdf.SetFont(fontBold, "", 9)
	pdf.SetXY(textX, sigY+6)
	pdf.CellWithOption(&gopdf.Rect{W: colW, H: 12}, "Chữ ký người nhận", gopdf.CellOption{Align: gopdf.Center})
	pdf.SetXY(textX+colW, sigY+6)
	pdf.CellWithOption(&gopdf.Rect{W: colW, H: 12}, "Chữ ký bưu tá", gopdf.CellOption{Align: gopdf.Center})

	var buf bytes.Buffer
	if _, err := pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("ghi PDF: %w", err)
	}
	return buf.Bytes(), nil
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
