package usecase_test

import (
	"context"
	"testing"
	"time"

	"ocean-express-api/internal/domain"
	"ocean-express-api/internal/usecase"
	"ocean-express-api/pkg/utils"
)

// ---------------------------------------------------------------------------
// Mocks (đặt tên riêng để không trùng với mock trong order_usecase_test.go)
// ---------------------------------------------------------------------------

type fakeEmployeeRepo struct {
	byPhone map[string]*domain.Employee
	byID    map[string]*domain.Employee
	created *domain.Employee
	updated *domain.Employee
}

func newFakeEmployeeRepo() *fakeEmployeeRepo {
	return &fakeEmployeeRepo{
		byPhone: map[string]*domain.Employee{},
		byID:    map[string]*domain.Employee{},
	}
}

func (r *fakeEmployeeRepo) GetByPhone(ctx context.Context, phone string) (*domain.Employee, error) {
	if e, ok := r.byPhone[phone]; ok {
		return e, nil
	}
	return nil, nil
}
func (r *fakeEmployeeRepo) GetByPhoneOrEmail(ctx context.Context, identifier string) (*domain.Employee, error) {
	return r.GetByPhone(ctx, identifier)
}
func (r *fakeEmployeeRepo) GetByID(ctx context.Context, id string) (*domain.Employee, error) {
	if e, ok := r.byID[id]; ok {
		return e, nil
	}
	return nil, nil
}
func (r *fakeEmployeeRepo) FindAll(ctx context.Context, hubID *string, status *string, pageParams domain.PaginationParams) ([]*domain.Employee, int64, error) {
	var out []*domain.Employee
	for _, e := range r.byID {
		if hubID != nil {
			if e.HubID == nil || *e.HubID != *hubID {
				continue
			}
		}
		if status != nil && e.Status != *status {
			continue
		}
		out = append(out, e)
	}
	return out, int64(len(out)), nil
}
func (r *fakeEmployeeRepo) Create(ctx context.Context, emp *domain.Employee) error {
	emp.ID = "emp-new"
	r.created = emp
	r.byID[emp.ID] = emp
	r.byPhone[emp.Phone] = emp
	return nil
}
func (r *fakeEmployeeRepo) Update(ctx context.Context, emp *domain.Employee) error {
	r.updated = emp
	r.byID[emp.ID] = emp
	return nil
}

type fakeShopRepo struct {
	byID    map[string]*domain.Shop
	created *domain.Shop
	updated *domain.Shop
}

func newFakeShopRepo() *fakeShopRepo {
	return &fakeShopRepo{byID: map[string]*domain.Shop{}}
}

func (r *fakeShopRepo) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Shop, error) {
	return nil, nil
}
func (r *fakeShopRepo) GetByID(ctx context.Context, id string) (*domain.Shop, error) {
	if s, ok := r.byID[id]; ok {
		return s, nil
	}
	return nil, nil
}
func (r *fakeShopRepo) GetByEmail(ctx context.Context, email string) (*domain.Shop, error) {
	for _, s := range r.byID {
		if s.Email != nil && *s.Email == email {
			return s, nil
		}
	}
	return nil, nil
}
func (r *fakeShopRepo) GetByPhoneOrEmail(ctx context.Context, identifier string) (*domain.Shop, error) {
	for _, s := range r.byID {
		if (s.Phone != nil && *s.Phone == identifier) || (s.Email != nil && *s.Email == identifier) {
			return s, nil
		}
	}
	return nil, nil
}
func (r *fakeShopRepo) FindAll(ctx context.Context, status string) ([]*domain.Shop, error) {
	var out []*domain.Shop
	for _, s := range r.byID {
		if status != "" && s.Status != status {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}
func (r *fakeShopRepo) Create(ctx context.Context, shop *domain.Shop) error {
	shop.ID = "shop-new"
	r.created = shop
	r.byID[shop.ID] = shop
	return nil
}
func (r *fakeShopRepo) Update(ctx context.Context, shop *domain.Shop) error {
	r.updated = shop
	r.byID[shop.ID] = shop
	return nil
}

type fakeSessionRepo struct {
	sessions map[string]string
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{sessions: map[string]string{}}
}

func (r *fakeSessionRepo) Create(ctx context.Context, jti, userID string, ttl time.Duration) error {
	r.sessions[jti] = userID
	return nil
}
func (r *fakeSessionRepo) Exists(ctx context.Context, jti string) (bool, error) {
	_, ok := r.sessions[jti]
	return ok, nil
}
func (r *fakeSessionRepo) Revoke(ctx context.Context, jti string) error {
	delete(r.sessions, jti)
	return nil
}
func (r *fakeSessionRepo) SetOTP(ctx context.Context, identifier, otp string, ttl time.Duration) error {
	r.sessions["otp:"+identifier] = otp
	return nil
}
func (r *fakeSessionRepo) GetOTP(ctx context.Context, identifier string) (string, error) {
	return r.sessions["otp:"+identifier], nil
}
func (r *fakeSessionRepo) DeleteOTP(ctx context.Context, identifier string) error {
	delete(r.sessions, "otp:"+identifier)
	return nil
}

// ---------------------------------------------------------------------------
// Employee usecase
// ---------------------------------------------------------------------------

func TestCreateEmployee_RejectsNonAdminWithoutHub(t *testing.T) {
	repo := newFakeEmployeeRepo()
	uc := usecase.NewEmployeeUseCase(repo)

	_, err := uc.CreateEmployee(context.Background(), "Tài xế A", "0911", "test@test.com", "pass", "first_mile_driver", nil)
	if err == nil {
		t.Fatal("mong đợi lỗi khi tài xế không gắn bưu cục")
	}
}

func TestCreateEmployee_RejectsInvalidRole(t *testing.T) {
	repo := newFakeEmployeeRepo()
	uc := usecase.NewEmployeeUseCase(repo)

	hub := "hub-1"
	_, err := uc.CreateEmployee(context.Background(), "X", "0912", "test@test.com", "pass", "super_hacker", &hub)
	if err == nil {
		t.Fatal("mong đợi lỗi khi role không hợp lệ")
	}
}

func TestCreateEmployee_RejectsDuplicatePhone(t *testing.T) {
	repo := newFakeEmployeeRepo()
	repo.byPhone["0913"] = &domain.Employee{ID: "existing", Phone: "0913"}
	uc := usecase.NewEmployeeUseCase(repo)

	_, err := uc.CreateEmployee(context.Background(), "X", "0913", "test@test.com", "pass", "admin", nil)
	if err == nil {
		t.Fatal("mong đợi lỗi khi số điện thoại trùng")
	}
}

func TestCreateEmployee_HashesPassword(t *testing.T) {
	repo := newFakeEmployeeRepo()
	uc := usecase.NewEmployeeUseCase(repo)

	emp, err := uc.CreateEmployee(context.Background(), "Admin", "0914", "test@test.com", "secret123", "admin", nil)
	if err != nil {
		t.Fatalf("không mong đợi lỗi: %v", err)
	}
	if emp.PasswordHash == "secret123" || emp.PasswordHash == "" {
		t.Fatal("mật khẩu phải được hash, không lưu plaintext")
	}
	if !utils.CheckPasswordHash("secret123", emp.PasswordHash) {
		t.Fatal("hash không khớp với mật khẩu gốc")
	}
}

func TestSetActive_TogglesFlag(t *testing.T) {
	repo := newFakeEmployeeRepo()
	repo.byID["emp-1"] = &domain.Employee{ID: "emp-1", Phone: "0915", IsActive: true}
	uc := usecase.NewEmployeeUseCase(repo)

	emp, err := uc.SetActive(context.Background(), "emp-1", false)
	if err != nil {
		t.Fatalf("không mong đợi lỗi: %v", err)
	}
	if emp.IsActive {
		t.Fatal("mong đợi tài khoản bị khóa (is_active=false)")
	}
	if repo.updated == nil || repo.updated.IsActive {
		t.Fatal("mong đợi repo lưu trạng thái khóa")
	}
}

func TestUpdateEmployee_KeepsPasswordWhenBlank(t *testing.T) {
	repo := newFakeEmployeeRepo()
	oldHash := "$2a$14$oldhashvalue"
	repo.byID["emp-1"] = &domain.Employee{ID: "emp-1", Name: "Cũ", Phone: "0916", PasswordHash: oldHash, Role: domain.RoleAdmin}
	uc := usecase.NewEmployeeUseCase(repo)

	emp, err := uc.UpdateEmployee(context.Background(), "emp-1", "Mới", "0916", "test@test.com", "", "admin", nil)
	if err != nil {
		t.Fatalf("không mong đợi lỗi: %v", err)
	}
	if emp.PasswordHash != oldHash {
		t.Fatal("password rỗng phải giữ nguyên hash cũ")
	}
	if emp.Name != "Mới" {
		t.Fatal("tên phải được cập nhật")
	}
}

// ---------------------------------------------------------------------------
// Shop usecase
// ---------------------------------------------------------------------------

func TestCreateShop_GeneratesApiKey(t *testing.T) {
	repo := newFakeShopRepo()
	uc := usecase.NewShopUseCase(repo, nil, nil)

	loc := "VN-HN"
	shop, apiKey, err := uc.CreateShop(context.Background(), "BC Sport", "09999", "https://shop.com/hook", &loc, "123 Lê Lợi", nil, nil)
	if err != nil {
		t.Fatalf("không mong đợi lỗi: %v", err)
	}
	if apiKey == "" {
		t.Fatal("mong đợi sinh API key")
	}
	if shop.APIKey != apiKey {
		t.Fatal("API key trên shop phải khớp key trả về")
	}
}

func TestCreateShop_RejectsMissingFields(t *testing.T) {
	repo := newFakeShopRepo()
	uc := usecase.NewShopUseCase(repo, nil, nil)

	if _, _, err := uc.CreateShop(context.Background(), "", "09999", "https://x", nil, "addr", nil, nil); err == nil {
		t.Fatal("mong đợi lỗi khi thiếu tên")
	}
}

func TestUpdateShop_DoesNotChangeApiKey(t *testing.T) {
	repo := newFakeShopRepo()
	repo.byID["shop-1"] = &domain.Shop{ID: "shop-1", Name: "Cũ", WebhookURL: "https://old", APIKey: "oe_secret", AddressDetail: "addr"}
	uc := usecase.NewShopUseCase(repo, nil, nil)

	shop, err := uc.UpdateShop(context.Background(), "shop-1", "Mới", "09999", "https://new", nil, "addr mới", nil, nil)
	if err != nil {
		t.Fatalf("không mong đợi lỗi: %v", err)
	}
	if shop.APIKey != "oe_secret" {
		t.Fatal("API key không được đổi khi update thông tin shop")
	}
	if shop.WebhookURL != "https://new" {
		t.Fatal("webhook_url phải được cập nhật")
	}
}

// ---------------------------------------------------------------------------
// Auth usecase + session
// ---------------------------------------------------------------------------

type fakeEmailService struct{}

func (s *fakeEmailService) SendShopApprovedEmail(email string) {}
func (s *fakeEmailService) SendOTP(toEmail, otp, role string) error { return nil }

func TestLogin_CreatesSession(t *testing.T) {
	empRepo := newFakeEmployeeRepo()
	hash, _ := utils.HashPassword("matkhau")
	empRepo.byPhone["0900"] = &domain.Employee{ID: "u1", Phone: "0900", PasswordHash: hash, Role: domain.RoleAdmin, IsActive: true}
	sessRepo := newFakeSessionRepo()
	uc := usecase.NewAuthUseCase(empRepo, sessRepo, &fakeEmailService{})

	token, emp, err := uc.Login(context.Background(), "0900", "matkhau")
	if err != nil {
		t.Fatalf("không mong đợi lỗi: %v", err)
	}
	if token == "" || emp == nil {
		t.Fatal("mong đợi trả token + employee")
	}
	if len(sessRepo.sessions) != 1 {
		t.Fatalf("mong đợi tạo 1 session trong Redis, có %d", len(sessRepo.sessions))
	}
}

func TestLogin_RejectsInactiveAccount(t *testing.T) {
	empRepo := newFakeEmployeeRepo()
	hash, _ := utils.HashPassword("matkhau")
	empRepo.byPhone["0901"] = &domain.Employee{ID: "u2", Phone: "0901", PasswordHash: hash, Role: domain.RoleAdmin, IsActive: false}
	uc := usecase.NewAuthUseCase(empRepo, newFakeSessionRepo(), &fakeEmailService{})

	if _, _, err := uc.Login(context.Background(), "0901", "matkhau"); err == nil {
		t.Fatal("mong đợi lỗi khi tài khoản bị khóa")
	}
}

func TestLogin_RejectsWrongPassword(t *testing.T) {
	empRepo := newFakeEmployeeRepo()
	hash, _ := utils.HashPassword("dung")
	empRepo.byPhone["0902"] = &domain.Employee{ID: "u3", Phone: "0902", PasswordHash: hash, Role: domain.RoleAdmin, IsActive: true}
	uc := usecase.NewAuthUseCase(empRepo, newFakeSessionRepo(), &fakeEmailService{})

	if _, _, err := uc.Login(context.Background(), "0902", "sai"); err == nil {
		t.Fatal("mong đợi lỗi khi sai mật khẩu")
	}
}

func TestLogout_RevokesSession(t *testing.T) {
	empRepo := newFakeEmployeeRepo()
	hash, _ := utils.HashPassword("matkhau")
	empRepo.byPhone["0903"] = &domain.Employee{ID: "u4", Phone: "0903", PasswordHash: hash, Role: domain.RoleAdmin, IsActive: true}
	sessRepo := newFakeSessionRepo()
	uc := usecase.NewAuthUseCase(empRepo, sessRepo, &fakeEmailService{})

	if _, _, err := uc.Login(context.Background(), "0903", "matkhau"); err != nil {
		t.Fatalf("login lỗi: %v", err)
	}

	// Lấy jti vừa tạo rồi logout
	var jti string
	for k := range sessRepo.sessions {
		jti = k
	}
	if err := uc.Logout(context.Background(), jti); err != nil {
		t.Fatalf("logout lỗi: %v", err)
	}
	if len(sessRepo.sessions) != 0 {
		t.Fatal("mong đợi session bị thu hồi sau logout")
	}
}
