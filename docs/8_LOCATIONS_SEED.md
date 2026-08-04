# Nạp dữ liệu địa chỉ hành chính (Locations Seed)

Tài liệu hướng dẫn nạp cây địa chỉ hành chính Việt Nam vào bảng `locations`, theo
**cải cách hành chính 2025** (34 tỉnh/thành, **bỏ cấp huyện** → cấu trúc **2 cấp**:
Tỉnh/Thành → Phường/Xã).

## 1. Vì sao dùng loader thay vì hardcode trong `init.sql`

- Dữ liệu ~3.300 phường/xã **không được gõ tay từ trí nhớ** — sai một tên/mã là hỏng
  tính cước, phân tuyến, gán hub. Dữ liệu **phải đến từ nguồn chính thống**.
- `init.sql` chỉ chạy khi volume DB tạo mới. Loader chạy được **bất cứ lúc nào** trên
  DB đang sống, **idempotent** (chạy lại an toàn).
- Tách khỏi `init.sql` để không đụng phần seed khác (admin, migrations).

## 2. Cấu trúc 2 cấp

Cải cách 2025 bỏ cấp huyện. Trong bảng `locations`:

- Tỉnh/thành: `type = 'province'`, `parent_id = NULL`.
- Phường/xã: `type = 'ward'`, `parent_id =` mã tỉnh (trỏ **thẳng** lên tỉnh, không qua huyện).
- **Không** có bản ghi `type = 'district'`.

> Lưu ý: cột `type` vẫn cho phép giá trị `district` (schema cũ), nhưng bộ dữ liệu 2025
> không sinh ra bản ghi nào loại này. FE nào giả định đủ 3 cấp cần điều chỉnh về 2 cấp.

## 3. Định dạng file JSON đầu vào

Xem mẫu tại [`api/data/locations.sample.json`](../api/data/locations.sample.json). Cấu trúc lồng:

```json
{
  "provinces": [
    {
      "code": "VN-HN",
      "name": "Thành phố Hà Nội",
      "type": "province",
      "wards": [
        { "code": "VN-HN-00001", "name": "Phường Ba Đình", "type": "ward" }
      ]
    }
  ]
}
```

Yêu cầu:
- Mỗi tỉnh và mỗi xã phải có `code` (dùng làm khóa chính `locations.id`) và `name`.
- `code` **không được trùng** trong toàn file (loader sẽ báo lỗi và dừng).
- `type` có thể bỏ trống — loader mặc định tỉnh = `province`, xã = `ward`.

## 4. Lấy dữ liệu thật ở đâu

Loader **không kèm** dữ liệu đầy đủ (chỉ có file mẫu 3 tỉnh để chạy thử). Bạn tự tải
bộ 34 tỉnh + phường/xã 2025 từ nguồn bạn tin tưởng (cơ sở dữ liệu quốc gia về đơn vị
hành chính, hoặc dataset cộng đồng), rồi **chuyển về đúng định dạng mục 3** và lưu vào
`api/data/locations.json`.

> Nếu dataset nguồn ở định dạng khác (vd mảng phẳng, hoặc còn cấp huyện), cần viết một
> bước chuyển đổi nhỏ về cấu trúc lồng 2 cấp ở trên trước khi nạp.

## 5. Chạy loader

Từ thư mục `api/`:

```bash
# Kiểm tra file trước (không ghi DB) — nên chạy trước tiên
go run ./cmd/seedloc -file data/locations.json -dry-run

# Nạp thật vào DB
go run ./cmd/seedloc -file data/locations.json
```

Loader đọc cấu hình DB từ biến môi trường (giống API server): `DB_HOST`, `DB_PORT`,
`DB_USER`, `DB_PASS`, `DB_NAME` (mặc định `localhost:5432`, `root`/`rootpassword`,
`ocean_express_db`).

Chạy trong Docker (DB nằm trong compose network):

```bash
# từ máy host, trỏ vào Postgres đã map ra cổng 5432
DB_HOST=localhost go run ./cmd/seedloc -file data/locations.json
```

### Hành vi
- **Idempotent**: dùng `ON CONFLICT (id)` → chạy lại chỉ cập nhật `name`/`type`/`parent_id`,
  không tạo bản ghi trùng.
- **Thứ tự an toàn FK**: nạp toàn bộ tỉnh trước, xã sau (vì `parent_id` của xã trỏ lên tỉnh).
- **Cảnh báo mềm**: nếu số tỉnh ≠ 34, loader in cảnh báo (không dừng) để bạn soát lại
  bộ dữ liệu có đầy đủ không.

## 6. Kiểm tra sau khi nạp

```sql
-- Số tỉnh (kỳ vọng 34)
SELECT COUNT(*) FROM locations WHERE type = 'province';

-- Số xã
SELECT COUNT(*) FROM locations WHERE type = 'ward';

-- Xã mồ côi (parent trỏ tới tỉnh không tồn tại) — kỳ vọng 0 dòng
SELECT w.id, w.name, w.parent_id
FROM locations w
LEFT JOIN locations p ON p.id = w.parent_id
WHERE w.type = 'ward' AND p.id IS NULL;
```
