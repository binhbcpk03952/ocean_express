package main

// seedloc nạp dữ liệu địa chỉ hành chính Việt Nam (2 cấp: Tỉnh/Thành → Phường/Xã,
// theo cải cách 2025) vào bảng locations.
//
// Dữ liệu KHÔNG hardcode trong code — đọc từ file JSON do người dùng cung cấp
// (nguồn chính thống). Xem data/locations.sample.json cho định dạng.
//
// Cách dùng:
//   go run ./cmd/seedloc -file data/locations.json
//   go run ./cmd/seedloc -file data/locations.json -dry-run   # chỉ kiểm tra file, không ghi DB
//
// Idempotent: chạy lại nhiều lần an toàn (ON CONFLICT theo id -> cập nhật name/type/parent_id).

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Cấu trúc file JSON đầu vào (khớp data/locations.sample.json).
type wardJSON struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type provinceJSON struct {
	Code  string     `json:"code"`
	Name  string     `json:"name"`
	Type  string     `json:"type"`
	Wards []wardJSON `json:"wards"`
}

type fileJSON struct {
	Provinces []provinceJSON `json:"provinces"`
}

// locationRow ánh xạ tối thiểu tới bảng locations (không import internal/domain để
// seedloc đứng độc lập, không phụ thuộc vòng đời package nghiệp vụ).
type locationRow struct {
	ID        string    `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"column:name"`
	Type      string    `gorm:"column:type"`
	ParentID  *string   `gorm:"column:parent_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (locationRow) TableName() string { return "locations" }

func main() {
	filePath := flag.String("file", "data/locations.json", "đường dẫn file JSON dữ liệu địa chỉ")
	dryRun := flag.Bool("dry-run", false, "chỉ kiểm tra & thống kê file, không ghi vào DB")
	flag.Parse()

	// 1. Đọc & parse file
	raw, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("Không đọc được file %s: %v", *filePath, err)
	}

	var data fileJSON
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Fatalf("File %s không phải JSON hợp lệ theo định dạng mong đợi: %v", *filePath, err)
	}

	// 2. Chuẩn hóa thành danh sách rows + validate cơ bản
	rows, provinceCount, wardCount, err := buildRows(data)
	if err != nil {
		log.Fatalf("Dữ liệu không hợp lệ: %v", err)
	}

	log.Printf("Đã phân tích %d tỉnh/thành và %d phường/xã (tổng %d bản ghi) từ %s",
		provinceCount, wardCount, len(rows), *filePath)

	// Cảnh báo mềm: cải cách hành chính 2025 còn 34 tỉnh/thành. Lệch số này thường
	// là dấu hiệu file thiếu dữ liệu hoặc dùng nhầm bộ cũ — không fail để vẫn cho
	// phép nạp tập con khi cần (vd môi trường test).
	if provinceCount != 34 {
		log.Printf("CẢNH BÁO: file có %d tỉnh/thành, khác con số kỳ vọng 34 (cải cách 2025). Kiểm tra lại nguồn dữ liệu nếu đây là bộ đầy đủ.", provinceCount)
	}

	if *dryRun {
		log.Println("dry-run: bỏ qua ghi DB. File hợp lệ.")
		return
	}

	// 3. Kết nối DB (cùng biến môi trường với API server)
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Không kết nối được database: %v", err)
	}

	// 4. Upsert theo batch. Province phải vào trước ward (FK parent_id).
	//    Sắp xếp: type=province trước. buildRows đã đảm bảo thứ tự này.
	const batchSize = 500
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "type", "parent_id"}),
	}).CreateInBatches(rows, batchSize).Error; err != nil {
		log.Fatalf("Lỗi khi ghi dữ liệu vào bảng locations: %v", err)
	}

	log.Printf("Hoàn tất: đã nạp/cập nhật %d đơn vị hành chính vào bảng locations.", len(rows))
}

// buildRows làm phẳng cấu trúc lồng thành danh sách rows, provinces trước wards,
// và validate: không rỗng code/name, không trùng code.
func buildRows(data fileJSON) (rows []locationRow, provinceCount, wardCount int, err error) {
	if len(data.Provinces) == 0 {
		return nil, 0, 0, fmt.Errorf("không có tỉnh/thành nào trong file (mảng 'provinces' rỗng)")
	}

	seen := map[string]bool{}
	var provinces, wards []locationRow

	for _, p := range data.Provinces {
		if p.Code == "" || p.Name == "" {
			return nil, 0, 0, fmt.Errorf("tỉnh/thành thiếu code hoặc name: %+v", p)
		}
		if seen[p.Code] {
			return nil, 0, 0, fmt.Errorf("trùng mã: %s", p.Code)
		}
		seen[p.Code] = true

		pType := p.Type
		if pType == "" {
			pType = "province"
		}
		provinces = append(provinces, locationRow{ID: p.Code, Name: p.Name, Type: pType, ParentID: nil})
		provinceCount++

		for _, w := range p.Wards {
			if w.Code == "" || w.Name == "" {
				return nil, 0, 0, fmt.Errorf("phường/xã thiếu code hoặc name (thuộc %s): %+v", p.Code, w)
			}
			if seen[w.Code] {
				return nil, 0, 0, fmt.Errorf("trùng mã: %s", w.Code)
			}
			seen[w.Code] = true

			wType := w.Type
			if wType == "" {
				wType = "ward"
			}
			parent := p.Code // cấu trúc 2 cấp: ward trỏ thẳng lên province
			wards = append(wards, locationRow{ID: w.Code, Name: w.Name, Type: wType, ParentID: &parent})
			wardCount++
		}
	}

	// provinces trước để thỏa ràng buộc FK parent_id khi ward tham chiếu lên.
	rows = append(rows, provinces...)
	rows = append(rows, wards...)
	return rows, provinceCount, wardCount, nil
}

func connectDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "root"),
		getEnv("DB_PASS", "rootpassword"),
		getEnv("DB_NAME", "ocean_express_db"),
		getEnv("DB_PORT", "5432"),
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
