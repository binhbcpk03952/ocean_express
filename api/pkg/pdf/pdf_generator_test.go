package pdf

import (
	"bytes"
	"testing"

	"ocean-express-api/internal/domain"
)

func sampleOrder() *domain.ShippingOrder {
	return &domain.ShippingOrder{
		TrackingNumber:        "OCEAN-12345678",
		ReceiverName:          "Cửa Hàng Xăng Dầu",
		ReceiverPhone:         "0912345678",
		ReceiverAddressDetail: "Cây xăng Quảng Tân, Xã Hải Trạch, Huyện Bố Trạch, Tỉnh Quảng Bình",
		Weight:                1200,
		CodAmount:             250000,
	}
}

func sampleShop() *domain.Shop {
	phone := "0909090909"
	return &domain.Shop{
		Name:          "Shop Thời Trang Biển",
		Phone:         &phone,
		AddressDetail: "Số 10 Đường Hoàng Diệu, Quận Hải Châu, Đà Nẵng",
	}
}

func TestGenerateOrderLabelPDF(t *testing.T) {
	got, err := GenerateOrderLabelPDF(sampleOrder(), sampleShop())
	if err != nil {
		t.Fatalf("GenerateOrderLabelPDF() error = %v", err)
	}
	if !bytes.HasPrefix(got, []byte("%PDF-")) {
		t.Errorf("kết quả không phải PDF hợp lệ, 8 byte đầu = %q", got[:min(8, len(got))])
	}
	if len(got) < 1000 {
		t.Errorf("PDF quá nhỏ (%d byte), có thể font chưa được nhúng", len(got))
	}
}

func TestGenerateBatchOrderLabelsPDF(t *testing.T) {
	orders := []*domain.ShippingOrder{
		sampleOrder(),
		{TrackingNumber: "OCEAN-BATCH-2", ReceiverName: "Khách 2", ReceiverPhone: "0999888777", Weight: 800, CodAmount: 150000},
	}
	shopMap := map[string]*domain.Shop{
		sampleOrder().ShopID: sampleShop(),
	}
	got, err := GenerateBatchOrderLabelsPDF(orders, shopMap)
	if err != nil {
		t.Fatalf("GenerateBatchOrderLabelsPDF() error = %v", err)
	}
	if !bytes.HasPrefix(got, []byte("%PDF-")) {
		t.Errorf("kết quả batch không phải PDF hợp lệ")
	}
	if len(got) < 2000 {
		t.Errorf("PDF batch quá nhỏ (%d byte)", len(got))
	}
}

// Các trường rỗng/zero không được làm hàm panic hoặc trả lỗi.
func TestGenerateOrderLabelPDF_EmptyFields(t *testing.T) {
	if _, err := GenerateOrderLabelPDF(&domain.ShippingOrder{TrackingNumber: "OCEAN-0"}, nil); err != nil {
		t.Fatalf("đơn hàng rỗng gây lỗi: %v", err)
	}
}

// Địa chỉ và tên rất dài phải được ngắt dòng, không tràn ra ngoài viền tem.
func TestGenerateOrderLabelPDF_LongFields(t *testing.T) {
	o := sampleOrder()
	o.ReceiverName = "Nguyễn Thị Hoàng Phương Anh Tuyết Mai"
	o.ReceiverPhone = "0987654321"
	o.ReceiverAddressDetail = "Số 123 ngách 45 ngõ 67 đường Nguyễn Văn Cừ, Tổ dân phố Thượng Đình, " +
		"Phường Khương Trung, Quận Thanh Xuân, Thành phố Hà Nội, Việt Nam"
	o.CodAmount = 12500000
	if _, err := GenerateOrderLabelPDF(o, sampleShop()); err != nil {
		t.Fatalf("đơn hàng có trường dài gây lỗi: %v", err)
	}
}

// wrapText phải ngắt theo ranh giới từ, không cắt giữa từ như gopdf.SplitText.
func TestWrapTextBreaksOnWordBoundary(t *testing.T) {
	pdf, err := newMeasuringPDF()
	if err != nil {
		t.Fatalf("khởi tạo pdf đo chữ: %v", err)
	}
	pdf.SetFont(fontRegular, "", 10)

	const addr = "Cây xăng Quảng Tân, Thôn Hải Đông, Xã Hải Trạch, Huyện Bố Trạch, Tỉnh Quảng Bình, Việt Nam"
	lines := wrapText(pdf, addr, textW, 3)
	if len(lines) < 2 {
		t.Fatalf("mong đợi địa chỉ bị ngắt thành nhiều dòng, nhận được %d", len(lines))
	}
	for i, ln := range lines {
		if w := textWidth(pdf, ln); w > textW {
			t.Errorf("dòng [%d] rộng %.1f > %.1f: %q", i, w, textW, ln)
		}
	}
	// Mỗi dòng phải bắt đầu/kết thúc trọn từ: ghép lại bằng dấu cách là chuỗi gốc.
	if joined := joinWords(lines); joined != addr {
		t.Errorf("ngắt dòng làm hỏng nội dung:\n  gốc = %q\n  ghép= %q", addr, joined)
	}
}

func TestWrapTextTruncatesWithEllipsis(t *testing.T) {
	pdf, err := newMeasuringPDF()
	if err != nil {
		t.Fatalf("khởi tạo pdf đo chữ: %v", err)
	}
	pdf.SetFont(fontRegular, "", 10)

	long := "Số 123 ngách 45 ngõ 67 đường Nguyễn Văn Cừ, Tổ dân phố Thượng Đình, " +
		"Phường Khương Trung, Quận Thanh Xuân, Thành phố Hà Nội, Việt Nam, Đông Nam Á, Thế Giới"
	lines := wrapText(pdf, long, textW, 2)
	if len(lines) != 2 {
		t.Fatalf("mong đợi bị cắt còn 2 dòng, nhận được %d", len(lines))
	}
	last := []rune(lines[1])
	if last[len(last)-1] != '…' {
		t.Errorf("dòng cuối phải kết thúc bằng dấu …, nhận được %q", lines[1])
	}
	for i, ln := range lines {
		if w := textWidth(pdf, ln); w > textW {
			t.Errorf("dòng [%d] rộng %.1f > %.1f: %q", i, w, textW, ln)
		}
	}
}

func TestFormatThousands(t *testing.T) {
	cases := map[int64]string{
		0: "0", 5: "5", 999: "999", 1000: "1.000",
		250000: "250.000", 12500000: "12.500.000", -1500: "-1.500",
	}
	for in, want := range cases {
		if got := formatThousands(in); got != want {
			t.Errorf("formatThousands(%d) = %q, mong đợi %q", in, got, want)
		}
	}
}
