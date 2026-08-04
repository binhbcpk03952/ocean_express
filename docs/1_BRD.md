# Tài liệu Yêu cầu Nghiệp vụ (BRD) - Ocean Express

## 1. Mục tiêu
Hệ thống lõi Logistics (SaaS) quản lý toàn bộ chu trình vận chuyển hàng hóa, từ điểm lấy hàng, xử lý tại kho (Hub), cho tới điểm giao nhận cuối cùng.

## 2. Các Vai trò (Roles)
1. **First-mile Driver (Tài xế lấy hàng)**:
   - Nhận nhiệm vụ lấy hàng từ Shop/Người gửi.
   - Xác nhận đã lấy hàng và chở hàng về Hub.
2. **Hub Staff (Nhân viên Kho)**:
   - Quét mã vạch (scan) tiếp nhận hàng từ First-mile Driver (Inbound).
   - Phân loại và xuất hàng giao cho Last-mile Driver (Outbound).
3. **Last-mile Driver (Tài xế giao hàng)**:
   - Nhận hàng từ Hub.
   - Giao hàng tới người nhận cuối cùng (End-user) và cập nhật trạng thái Giao thành công / Thất bại.

## 3. Kịch bản Luân chuyển Kho
- Khi shop tạo đơn, đơn tự động gán vào khu vực của First-mile Driver (theo mã bưu cục/kho vực).
- Driver mang hàng về Hub (Kho trung tâm). Tại Hub, nhân viên thực hiện thao tác Inbound để ghi nhận hàng đã nhập kho an toàn.
- Dựa trên địa chỉ người nhận, hàng hóa được phân loại và thực hiện Outbound giao cho Last-mile Driver.
- Một đơn hàng có thể đi qua nhiều Hub (Từ Hub A -> Hub B) nếu lộ trình liên tỉnh. 

## 4. Phân quyền và Ràng buộc
- **Driver** chỉ được xem và cập nhật trạng thái đơn hàng được gán cho chính mình.
- **Hub Staff** có quyền quét tất cả đơn hàng nhưng chỉ được thực hiện thao tác Inbound/Outbound đối với các đơn hàng vật lý có mặt tại Hub của họ.
- **Admin** có toàn quyền xem thống kê, điều phối đơn hàng và chỉnh sửa thông tin nhân viên, shop.
