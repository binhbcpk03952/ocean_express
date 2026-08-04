package domain

import "errors"

// Sentinel errors phân loại lỗi nghiệp vụ để tầng Delivery (HTTP handler) map sang
// đúng status code, thay vì trả 500 cho mọi lỗi. Dùng errors.Is để so khớp.
//
// Quy ước map (xem HTTPStatusForError):
//   ErrForbidden          -> 403  (sai role / ngoài phạm vi dữ liệu)
//   ErrInvalidTransition  -> 422  (chuyển trạng thái sai luồng state machine)
//   ErrNotFound           -> 404  (không tìm thấy tài nguyên)
//   ErrValidation         -> 400  (dữ liệu đầu vào không hợp lệ)
//   còn lại               -> 500
var (
	ErrForbidden         = errors.New("bạn không có quyền thực hiện thao tác này")
	ErrInvalidTransition = errors.New("chuyển trạng thái không hợp lệ")
	ErrNotFound          = errors.New("không tìm thấy tài nguyên")
	ErrValidation        = errors.New("dữ liệu không hợp lệ")
)

// HTTPStatusForError map sentinel error sang HTTP status code. Trả 500 nếu không
// khớp sentinel nào (lỗi hệ thống không phân loại được).
func HTTPStatusForError(err error) int {
	switch {
	case errors.Is(err, ErrForbidden):
		return 403
	case errors.Is(err, ErrInvalidTransition):
		return 422
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.Is(err, ErrValidation):
		return 400
	default:
		return 500
	}
}

// ErrorCode trả về mã lỗi máy-đọc-được (dùng cho trường error.code trong JSON)
// dựa trên sentinel mà err bọc. Trả rỗng nếu không khớp sentinel nào (lỗi hệ thống).
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrForbidden):
		return "FORBIDDEN"
	case errors.Is(err, ErrInvalidTransition):
		return "INVALID_TRANSITION"
	case errors.Is(err, ErrNotFound):
		return "NOT_FOUND"
	case errors.Is(err, ErrValidation):
		return "VALIDATION_ERROR"
	default:
		return ""
	}
}
