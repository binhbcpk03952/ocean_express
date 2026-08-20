package main

import (
	"log"
	"os"

	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/pdf"
)

func main() {
	order := &domain.ShippingOrder{
		TrackingNumber:        "OCEAN-12345678",
		ReceiverName:          "Cửa Hàng Xăng Dầu",
		ReceiverPhone:         "0912345678",
		ReceiverAddressDetail: "Cây xăng Quảng Tân, Xã Hải Trạch, Huyện Bố Trạch, Tỉnh Quảng Bình",
		Weight:                1200,
		CodAmount:             250000,
	}

	shop := &domain.Shop{
		Name:          "Shop Thời Trang ABC",
		Phone:         stringPtr("0909123456"),
		AddressDetail: "123 Lê Lợi, Phường Bến Nghé, Quận 1, TP.HCM",
	}

	pdfBytes, err := pdf.GenerateOrderLabelPDF(order, shop)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("sample_label.pdf", pdfBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully generated sample_label.pdf")
}

func stringPtr(s string) *string {
	return &s
}
