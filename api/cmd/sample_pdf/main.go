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

	pdfBytes, err := pdf.GenerateOrderLabelPDF(order)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("sample_label.pdf", pdfBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully generated sample_label.pdf")
}
