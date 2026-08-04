# Lộ trình Triển khai Dự án Ocean Express (4 Phases)

Dự án Ocean Express - Hệ thống SaaS Logistics được chia thành 4 giai đoạn phát triển chính để đảm bảo tính ổn định, dễ dàng mở rộng và theo dõi tiến độ.

## Phase 1: Khởi tạo Nền tảng (Platform Scaffolding)
- **Mục tiêu**: Thiết lập bộ khung kiến trúc, môi trường phát triển và tài liệu định hướng.
- **Các tác vụ chính**:
  1. Tạo Context Routing cho AI (`/.agent/backend.md`, `frontend.md`, `devops.md`).
  2. Tạo tài liệu nghiệp vụ cơ bản, lộ trình (`/docs/ROADMAP_AND_PHASES.md`).
  3. Cấu hình Hạ tầng Docker (`docker-compose.yml`, `init.sql`).
  4. Scaffold cấu trúc Backend Go (Clean Architecture, Gin, GORM).
  5. Scaffold cấu trúc Frontend Vue 3 (Vite, Tailwind, Pinia, Vue Router).

## Phase 2: Phát triển Core Backend Services (Go)
- **Mục tiêu**: Xây dựng toàn bộ API nghiệp vụ, quản lý đơn hàng và tính cước.
- **Các tác vụ chính**:
  1. **Auth & RBAC**: Đăng nhập, phân quyền 4 role (Admin, First-mile, Hub, Last-mile). Quản lý JWT & Session qua Redis.
  2. **Location & Hubs**: API quản lý cây phân cấp địa lý hành chính và danh sách các bưu cục.
  3. **Shipping Rates**: Logic tính tiền cước tự động dựa trên khoảng cách (Location) và khối lượng (Steps).
  4. **Order Management**: Flow tạo đơn hàng, quản lý trạng thái đơn hàng (State Machine).
  5. **Tracking & GPS**: API lưu log hành trình và tọa độ.
  6. **Webhook & Async**: Tích hợp gửi Webhook về cho Shop (bằng Goroutine).

## Phase 3: Phát triển Frontend UI/UX (Vue 3 Admin)
- **Mục tiêu**: Cung cấp giao diện quản trị cho Admin, Nhân viên điều phối và Hub Staff.
- **Các tác vụ chính**:
  1. **Layout & Theme**: Xây dựng UI Component system với Tailwind.
  2. **Auth Views**: Màn hình Login và điều hướng dựa theo Role.
  3. **Order Dashboard**: Màn hình xem danh sách vận đơn, filter trạng thái, hiển thị chi tiết (Chi phí, Khối lượng, Lộ trình).
  4. **Hub Operator Panel**: Bảng điều khiển riêng cho Hub Staff thực hiện Inbound/Outbound.
  5. **Tracking Map**: Tích hợp bản đồ hiển thị toạ độ thực tế của Shipper.

## Phase 4: Kiểm thử, Tối ưu & Triển khai (CI/CD)
- **Mục tiêu**: Đảm bảo chất lượng hệ thống và go-live.
- **Các tác vụ chính**:
  1. Viết Unit Test cho các UseCase quan trọng (tính cước, state machine).
  2. Thực hiện API Load testing.
  3. Cấu hình CI/CD Pipelines (GitHub Actions / GitLab CI).
  4. Tối ưu hoá Docker Image (Multi-stage) để đưa lên môi trường Production.
  5. Bàn giao hệ thống.
