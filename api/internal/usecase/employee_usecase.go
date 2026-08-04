package usecase

import (
	"context"
	"errors"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/utils"
)

type employeeUseCase struct {
	employeeRepo domain.EmployeeRepository
}

func NewEmployeeUseCase(repo domain.EmployeeRepository) domain.EmployeeUseCase {
	return &employeeUseCase{employeeRepo: repo}
}

func (u *employeeUseCase) GetEmployees(ctx context.Context, hubID *string, status *string, pageParams domain.PaginationParams) (*domain.PaginatedResponse, error) {
	emps, total, err := u.employeeRepo.FindAll(ctx, hubID, status, pageParams)
	if err != nil {
		return nil, errors.New("không thể lấy danh sách nhân viên")
	}

	return &domain.PaginatedResponse{
		Data: emps,
		Meta: domain.PaginationMeta{
			Page:       pageParams.Page,
			Limit:      pageParams.Limit,
			TotalItems: total,
			TotalPages: domain.CalculateTotalPages(total, pageParams.GetLimit()),
		},
	}, nil
}

func (u *employeeUseCase) CreateEmployee(ctx context.Context, name, phone, email, password, role string, hubID *string) (*domain.Employee, error) {
	if name == "" || phone == "" || password == "" {
		return nil, errors.New("tên, số điện thoại và mật khẩu không được để trống")
	}

	// Chỉ chấp nhận 5 role hợp lệ của hệ thống
	validRoles := map[string]bool{
		string(domain.RoleAdmin):           true,
		string(domain.RoleFirstMileDriver): true,
		string(domain.RoleHubStaff):        true,
		string(domain.RoleTransitDriver):   true,
		string(domain.RoleLastMileDriver):  true,
	}
	if !validRoles[role] {
		return nil, errors.New("vai trò không hợp lệ")
	}

	// Hub staff và tài xế địa phương bắt buộc gắn với một bưu cục để phân luồng/lọc đơn.
	requiresHub := role == string(domain.RoleHubStaff) || role == string(domain.RoleFirstMileDriver) || role == string(domain.RoleLastMileDriver)
	if requiresHub && (hubID == nil || *hubID == "") {
		return nil, errors.New("nhân viên kho và tài xế địa phương bắt buộc phải thuộc một bưu cục")
	}

	// Không cho trùng số điện thoại
	existing, err := u.employeeRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, errors.New("lỗi hệ thống khi kiểm tra số điện thoại")
	}
	if existing != nil {
		return nil, errors.New("số điện thoại đã được sử dụng")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("không thể mã hóa mật khẩu")
	}

	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	emp := &domain.Employee{
		Name:         name,
		Phone:        phone,
		Email:        emailPtr,
		PasswordHash: hash,
		Role:         domain.EmployeeRole(role),
		HubID:        hubID,
		Status:       domain.StatusApproved, // Admin tạo trực tiếp -> duyệt ngay
		IsActive:     true,
	}

	if err := u.employeeRepo.Create(ctx, emp); err != nil {
		return nil, errors.New("lỗi khi tạo tài khoản nhân viên")
	}

	return emp, nil
}

// RegisterEmployee là shipper tự đăng ký. Chỉ chấp nhận role tài xế (first/last-mile),
// bắt buộc tự chọn hub, và tài khoản ở trạng thái chờ duyệt (pending + inactive).
func (u *employeeUseCase) RegisterEmployee(ctx context.Context, name, phone, email, password, role string, hubID *string) (*domain.Employee, error) {
	if name == "" || phone == "" || password == "" {
		return nil, errors.New("tên, số điện thoại và mật khẩu không được để trống")
	}

	// Self-register chỉ dành cho tài xế; admin và hub_staff do Admin tạo trực tiếp.
	if role != string(domain.RoleFirstMileDriver) && role != string(domain.RoleLastMileDriver) {
		return nil, errors.New("chỉ tài xế lấy hàng hoặc giao hàng mới được tự đăng ký")
	}

	// Tài xế bắt buộc chọn một bưu cục để phân luồng đơn.
	if hubID == nil || *hubID == "" {
		return nil, errors.New("vui lòng chọn bưu cục trực thuộc")
	}

	existing, err := u.employeeRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, errors.New("lỗi hệ thống khi kiểm tra số điện thoại")
	}
	if existing != nil {
		return nil, errors.New("số điện thoại đã được sử dụng")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("không thể mã hóa mật khẩu")
	}

	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	emp := &domain.Employee{
		Name:         name,
		Phone:        phone,
		Email:        emailPtr,
		PasswordHash: hash,
		Role:         domain.EmployeeRole(role),
		HubID:        hubID,
		Status:       domain.StatusPending, // chờ Admin duyệt
		IsActive:     false,                // chưa đăng nhập được cho tới khi duyệt
	}

	if err := u.employeeRepo.Create(ctx, emp); err != nil {
		return nil, errors.New("lỗi khi tạo tài khoản")
	}

	return emp, nil
}

// ReviewEmployee duyệt (approve) hoặc từ chối (reject) một tài khoản đang chờ.
func (u *employeeUseCase) ReviewEmployee(ctx context.Context, id string, approve bool) (*domain.Employee, error) {
	emp, err := u.employeeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("lỗi hệ thống khi truy xuất nhân viên")
	}
	if emp == nil {
		return nil, errors.New("không tìm thấy nhân viên")
	}

	if approve {
		emp.Status = domain.StatusApproved
		emp.IsActive = true
	} else {
		emp.Status = domain.StatusRejected
		emp.IsActive = false
	}

	if err := u.employeeRepo.Update(ctx, emp); err != nil {
		return nil, errors.New("lỗi khi cập nhật trạng thái duyệt")
	}

	return emp, nil
}

func (u *employeeUseCase) UpdateEmployee(ctx context.Context, id, name, phone, email, password, role string, hubID *string) (*domain.Employee, error) {
	if id == "" || name == "" || phone == "" {
		return nil, errors.New("id, tên và số điện thoại không được để trống")
	}

	emp, err := u.employeeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("lỗi hệ thống khi truy xuất nhân viên")
	}
	if emp == nil {
		return nil, errors.New("không tìm thấy nhân viên")
	}

	validRoles := map[string]bool{
		string(domain.RoleAdmin):           true,
		string(domain.RoleFirstMileDriver): true,
		string(domain.RoleHubStaff):        true,
		string(domain.RoleTransitDriver):   true,
		string(domain.RoleLastMileDriver):  true,
	}
	if !validRoles[role] {
		return nil, errors.New("vai trò không hợp lệ")
	}

	// Hub staff và tài xế địa phương bắt buộc gắn với một bưu cục để phân luồng/lọc đơn.
	requiresHub := role == string(domain.RoleHubStaff) || role == string(domain.RoleFirstMileDriver) || role == string(domain.RoleLastMileDriver)
	if requiresHub && (hubID == nil || *hubID == "") {
		return nil, errors.New("nhân viên kho và tài xế địa phương bắt buộc phải thuộc một bưu cục")
	}

	// Nếu đổi số điện thoại, không được trùng với người khác.
	if phone != emp.Phone {
		existing, err := u.employeeRepo.GetByPhone(ctx, phone)
		if err != nil {
			return nil, errors.New("lỗi hệ thống khi kiểm tra số điện thoại")
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("số điện thoại đã được sử dụng")
		}
	}

	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	emp.Name = name
	emp.Phone = phone
	emp.Email = emailPtr
	emp.Role = domain.EmployeeRole(role)
	emp.HubID = hubID

	// password rỗng = giữ nguyên mật khẩu cũ.
	if password != "" {
		hash, err := utils.HashPassword(password)
		if err != nil {
			return nil, errors.New("không thể mã hóa mật khẩu")
		}
		emp.PasswordHash = hash
	}

	if err := u.employeeRepo.Update(ctx, emp); err != nil {
		return nil, errors.New("lỗi khi cập nhật nhân viên")
	}

	return emp, nil
}

func (u *employeeUseCase) SetActive(ctx context.Context, id string, active bool) (*domain.Employee, error) {
	emp, err := u.employeeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("lỗi hệ thống khi truy xuất nhân viên")
	}
	if emp == nil {
		return nil, errors.New("không tìm thấy nhân viên")
	}

	emp.IsActive = active
	if err := u.employeeRepo.Update(ctx, emp); err != nil {
		return nil, errors.New("lỗi khi cập nhật trạng thái tài khoản")
	}

	return emp, nil
}
