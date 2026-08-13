
# SHIPMENT MANAGEMENT SYSTEM — BUSINESS & TECHNICAL AUDIT SPECIFICATION

> **Document purpose:** Đây là tài liệu đặc tả nghiệp vụ + checklist audit dành cho AI Agent/Developer dùng để kiểm tra và hoàn thiện một hệ thống quản lý vận đơn đang xây dựng dở.
>
> **Nguyên tắc quan trọng:** Không được tự ý viết lại toàn bộ hệ thống, không được phá vỡ chức năng đang chạy. Agent phải **đọc code hiện tại → đối chiếu tài liệu → phát hiện gap → đề xuất → sửa theo từng module → test → báo cáo**.

---

## 0. CONTEXT

Hệ thống là một **Shipment / Logistics Management System** dùng để quản lý toàn bộ vòng đời vận đơn:

```text
Create Shipment
    ↓
Validation
    ↓
Confirmed
    ↓
Waiting Pickup
    ↓
Picked Up
    ↓
Hub / Warehouse
    ↓
In Transit
    ↓
Out For Delivery
    ├──→ Delivered
    │      ↓
    │   COD / Settlement
    │
    └──→ Delivery Failed
             ├──→ Retry
             └──→ Return
                      ↓
                   Returned
                      ↓
                   Settlement
```

Mục tiêu không phải chỉ xây CRUD vận đơn, mà phải tạo một hệ thống có thể:

- quản lý vận đơn;
- theo dõi tracking;
- quản lý pickup/delivery;
- điều phối shipper;
- quản lý hub/kho;
- quản lý giao thất bại;
- quản lý hoàn hàng;
- quản lý COD;
- đối soát;
- quản lý merchant/shop;
- báo cáo;
- notification;
- RBAC;
- audit log;
- API/webhook;
- SLA;
- mở rộng về sau.

---

# 1. AGENT OPERATING RULES

## 1.1. Trước khi sửa code

Agent MUST:

1. Scan toàn bộ repository.
2. Xác định frontend, backend, database, API, authentication, storage, queue, notification.
3. Xác định framework/version đang sử dụng.
4. Xác định cấu trúc folder.
5. Xác định các module đã hoàn thành.
6. Xác định các module đang làm dở.
7. Xác định các TODO/FIXME/mock/hard-code.
8. Kiểm tra migration/schema/database.
9. Kiểm tra API contract.
10. Kiểm tra frontend route/component/store/service.
11. Kiểm tra authentication và authorization.
12. Kiểm tra test hiện có.

**Không được giả định code đang dùng kiến trúc nào nếu chưa kiểm tra repository.**

---

## 1.2. Nguyên tắc sửa

Agent phải ưu tiên:

```text
Existing Architecture
        ↓
Existing Business Logic
        ↓
Existing Database
        ↓
Existing API Contract
        ↓
Minimal Safe Changes
        ↓
Test
```

Không được:


- rewrite toàn bộ project chỉ vì code chưa đẹp;
- đổi framework;
- đổi database;
- đổi API contract không có lý do;
- xóa dữ liệu/migration đang dùng;
- thay đổi business rule hiện tại mà không ghi nhận;
- tạo duplicate component/service/model;
- hard-code dữ liệu nghiệp vụ;
- bỏ qua validation;
- bỏ qua authorization;
- sửa frontend mà không kiểm tra backend contract.

---

# 2. EXPECTED USER ROLES

Tối thiểu phải xem xét:

| Role | Responsibility |
|---|---|
| Super Admin | Toàn quyền hệ thống |
| Admin | Quản trị vận hành |
| Operations | Xử lý và điều phối vận đơn |
| Warehouse/Hub Staff | Nhập/xuất/trung chuyển |
| Shipper | Pickup, delivery, cập nhật trạng thái |
| Finance | COD, transaction, reconciliation |
| Merchant/Shop | Tạo và theo dõi vận đơn |
| Customer | Tra cứu vận đơn |

Nếu project hiện tại có role khác, Agent phải giữ role hiện tại và map vào business capability tương ứng.

---

# 3. CORE MODULES

Hệ thống cần audit các module:

```text
01. Dashboard
02. Shipment Management
03. Pickup Management
04. Delivery Management
05. Shipper Management
06. Dispatch / Assignment
07. Hub / Warehouse
08. Transit / Transfer
09. Delivery Attempts
10. Return Management
11. Merchant Management
12. Customer Management
13. COD / Finance
14. Reconciliation / Settlement
15. Reports / Analytics
16. Notifications
17. Authentication
18. RBAC
19. Audit Logs
20. API / Webhooks
21. SLA
22. Settings
```

---

# 4. DASHBOARD

## Required KPI

Dashboard phải có khả năng hiển thị tối thiểu:

- Total shipments
- New shipments
- Waiting pickup
- Picked up
- In transit
- Out for delivery
- Delivered
- Delivery failed
- Returned
- Cancelled
- COD pending
- COD collected
- COD awaiting settlement
- Shipping revenue/cost

## Analytics

Nên có:

- Shipment volume by day/week/month
- Shipment status distribution
- Delivery success rate
- First-attempt success rate
- Failed delivery rate
- Return rate
- Average delivery time
- COD overview
- Hub performance
- Shipper performance
- Merchant performance
- SLA breach

## Audit questions

Agent phải trả lời:

- Dashboard lấy data từ API nào?
- KPI có chính xác với database không?
- Có hard-code không?
- Date filter có hoạt động không?
- Timezone có thống nhất không?
- Empty/loading/error state có không?

---

# 5. SHIPMENT MANAGEMENT

## 5.1. Shipment list

Phải hỗ trợ:

- Search tracking number
- Search order number
- Search customer phone
- Search customer name
- Search merchant
- Filter status
- Filter date
- Filter province/district/ward
- Filter hub
- Filter shipper
- Filter COD
- Filter service
- Filter delivery attempt
- Pagination
- Sorting
- Bulk actions
- Export

## 5.2. Shipment creation

Nên có wizard hoặc form rõ ràng:

### Sender

- Merchant
- Sender name
- Sender phone
- Sender address

### Receiver

- Receiver name
- Receiver phone
- Province
- District
- Ward
- Detailed address

### Package

- Product/item
- Quantity
- Weight
- Dimensions
- Declared value
- Notes

### Shipping

- Service
- COD
- Shipping fee
- Additional fee
- Insurance
- Pickup time
- Delivery note

## 5.3. Shipment detail

Phải bao quát:

- Shipment identity
- Current status
- Sender
- Receiver
- Package
- Financial data
- Assigned shipper
- Hub
- Tracking timeline
- Delivery attempts
- Return history
- Notes
- Documents/proof
- Audit history

---

# 6. SHIPMENT STATE MACHINE

Agent MUST kiểm tra state transition.

Một baseline:

```text
CREATED
  ↓
CONFIRMED
  ↓
WAITING_PICKUP
  ↓
PICKED_UP
  ↓
AT_HUB
  ↓
IN_TRANSIT
  ↓
OUT_FOR_DELIVERY
  ├──→ DELIVERED
  │
  └──→ DELIVERY_FAILED
           ├──→ RETRY
           └──→ RETURN_REQUESTED
                    ↓
                 RETURNING
                    ↓
                 RETURNED
```

Có thể có:

```text
CANCELLED
ON_HOLD
LOST
DAMAGED
```

nhưng chỉ thêm nếu business hiện tại cần.

## Critical rule

Không được chỉ lưu current status.

Phải có:

```text
shipments.status
shipment_events / shipment_status_histories
```

`shipments.status` = current state.

`shipment_events` = immutable history.

---

# 7. TRACKING

Tracking phải cho biết:

- timestamp
- status
- location
- hub
- actor
- shipper
- note
- proof/evidence nếu có

Ví dụ:

```text
05/08 09:42
OUT_FOR_DELIVERY
Shipper: Nguyễn A
Location: Buôn Ma Thuột

05/08 07:12
DEPARTED_HUB
Hub: Buôn Ma Thuột

04/08 22:30
ARRIVED_HUB
Hub: Buôn Ma Thuột

04/08 15:20
PICKED_UP
Shipper: Trần B

04/08 13:10
CREATED
```

---

# 8. PICKUP MANAGEMENT

Cần kiểm tra:

- Pickup request
- Pickup assignment
- Pickup schedule
- Pickup status
- Shipper
- Pickup failure
- Retry pickup
- Bulk pickup
- Pickup proof

Flow:

```text
Pickup Requested
      ↓
Assigned
      ↓
Shipper Accepted
      ↓
Arrived
      ↓
Picked Up
      └──→ Failed → Retry
```

---

# 9. DELIVERY MANAGEMENT

Cần có:

- Orders assigned to shipper
- Daily route/list
- Customer contact
- Address
- COD
- Delivery status
- Proof of delivery
- Delivery note
- Delivery attempt

Shipper cần thao tác nhanh:

```text
CALL
MAP
NAVIGATE
DELIVERED
FAILED
RETRY
```

---

# 10. DELIVERY FAILURE

Không được chỉ có `FAILED`.

Phải có reason:

- Customer unavailable
- Cannot contact
- Wrong address
- Customer rescheduled
- Customer rejected
- Insufficient cash
- Damaged package
- Other

Phải lưu:

```text
attempt_number
failure_reason
note
timestamp
shipper
location
```

Flow:

```text
OUT_FOR_DELIVERY
      ↓
FAILED
      ↓
RETRY
      ↓
OUT_FOR_DELIVERY
      ↓
FAILED
      ↓
RETURN
```

---

# 11. SHIPPER MANAGEMENT

Thông tin:

- Profile
- Phone
- Status
- Hub
- Service area
- Vehicle
- Assigned shipments
- Active shipments
- Delivered shipments
- Failed shipments
- COD held

Performance:

```text
Assigned
Delivered
Failed
Success Rate
First Attempt Success
Average Delivery Time
COD Held
```

---

# 12. DISPATCH / ASSIGNMENT

Phải kiểm tra:

- Manual assignment
- Bulk assignment
- Reassignment
- Auto assignment nếu có
- Area-based assignment
- Capacity consideration
- Shipper availability

Không được assign shipment cho shipper không hợp lệ nếu business rule cấm.

---

# 13. HUB / WAREHOUSE

Core entities:

```text
Hub
Hub Staff
Inbound
Outbound
Transfer
Hub Shipment
```

Flow:

```text
Pickup
 ↓
Origin Hub
 ↓
Transfer
 ↓
Destination Hub
 ↓
Last Mile
```

Phải audit:

- Scan in
- Scan out
- Hub transfer
- Shipment location
- Missing shipment
- Damaged shipment

---

# 14. RETURN MANAGEMENT

Return flow:

```text
Delivery Failed
      ↓
Return Requested
      ↓
Return Pickup
      ↓
Return In Transit
      ↓
Return Hub
      ↓
Returned To Merchant
      ↓
Completed
```

Phải quản lý:

- Return reason
- Return status
- Return fee
- Return shipment
- Return tracking
- Merchant confirmation

---

# 15. MERCHANT / SHOP

Merchant có thể:

- Create shipment
- Import shipment
- Track shipment
- Cancel shipment nếu rule cho phép
- Print label
- View COD
- View fees
- View settlement
- Export reports
- API integration

Merchant dashboard:

```text
Today shipments
Delivered
In transit
Failed
Returned
COD pending
Settlement pending
```

---

# 16. CUSTOMER

Customer-facing tracking nên tối giản:

```text
Tracking Number
      ↓
Current Status
      ↓
Estimated Delivery
      ↓
Tracking Timeline
      ↓
Receiver / Shipment summary
```

Không expose dữ liệu nội bộ không cần thiết.

---

# 17. COD / FINANCE

Phải phân biệt:

```text
COD expected
COD collected
COD held by shipper
COD transferred
COD reconciled
COD settled to merchant
```

Các loại phí có thể gồm:

- Shipping fee
- COD fee
- Return fee
- Insurance
- Additional fee
- Discount
- Adjustment

Không được chỉ có một field `total_price` nếu nghiệp vụ cần đối soát chi tiết.

---

# 18. RECONCILIATION / SETTLEMENT

Ví dụ:

```text
Gross COD
- Shipping Fee
- COD Fee
- Return Fee
- Other Fee
+/- Adjustment
----------------
Net Settlement
```

Phải có:

- Settlement period
- Merchant
- Included shipments
- Gross amount
- Fees
- Net amount
- Status
- Payment date
- Payment reference

Status:

```text
PENDING
RECONCILED
APPROVED
PAID
FAILED
CANCELLED
```

---

# 19. IMPORT / EXPORT

Import phải có:

```text
Upload
 ↓
Validate
 ↓
Preview
 ↓
Show row errors
 ↓
Confirm
 ↓
Create
```

Không được import trực tiếp mà không preview nếu dữ liệu có rủi ro.

Error phải chỉ rõ:

```text
Row 12:
Invalid phone

Row 15:
Missing ward

Row 22:
Invalid weight
```

Export:

- Shipment
- COD
- Settlement
- Reports
- Returns

---

# 20. LABEL / PRINT

Hỗ trợ:

- Shipment label
- Barcode
- QR Code
- Tracking number
- COD
- Receiver
- Sender
- Print single
- Bulk print

Nếu có printer-specific requirement thì phải tách print template khỏi business logic.

---

# 21. NOTIFICATIONS

Events nên có thể trigger:

```text
shipment.created
shipment.assigned
shipment.picked_up
shipment.arrived_hub
shipment.in_transit
shipment.out_for_delivery
shipment.delivered
shipment.failed
shipment.returned
settlement.created
settlement.paid
sla.breached
```

Channels tùy project:

- In-app
- Email
- SMS
- Push
- Telegram

Phải tránh gửi duplicate notification.

---

# 22. RBAC / AUTHORIZATION

Phải kiểm tra:

```text
User
Role
Permission
Role Permission
```

Không được chỉ ẩn button ở frontend.

Authorization phải được enforce ở backend/API.

Ví dụ:

```text
Frontend hides "Delete"
        ≠
Backend protects DELETE endpoint
```

Backend mới là source of truth.

---

# 23. AUDIT LOG

Các hành động quan trọng phải audit:

- Create shipment
- Update shipment
- Status change
- Assign shipper
- Reassign
- Change COD
- Change fee
- Cancel
- Return
- Settlement
- User permission change

Mỗi log nên có:

```text
actor
action
entity
entity_id
before
after
timestamp
ip/device nếu phù hợp
```

---

# 24. REPORTS / ANALYTICS

## Shipment

- Total shipments
- Volume by time
- Status distribution

## Delivery

- Success rate
- First attempt success
- Failure rate
- Average delivery time

## Return

- Return rate
- Return reasons
- Return by merchant
- Return by area

## Shipper

- Assigned
- Delivered
- Failed
- Success rate
- COD

## Hub

- Inbound
- Outbound
- Pending
- Processing time

## Finance

- COD
- Fees
- Settlement
- Outstanding

## SLA

- SLA compliance
- SLA breach
- Average delay

---

# 25. SLA

Nếu project có SLA:

```text
Created
 ↓ <= X hours
Pickup
 ↓ <= X hours
Hub
 ↓ <= X hours
Delivery
```

Cần xác định:

- SLA target
- Start time
- End time
- Working hours nếu có
- Exclusion
- Breach status

Dashboard nên có:

```text
SLA Compliance
SLA Breached
At Risk
```

---

# 26. SEARCH

Global search nên hỗ trợ:

```text
Tracking Number
Order Number
Phone
Customer
Merchant
Shipper
```

Search phải có:

- debounce
- pagination
- empty state
- loading state
- error state

---

# 27. UI/UX REQUIREMENTS

## Admin layout

```text
┌──────────────────────────────────────────────────────────┐
│ Logo      Global Search        Notification    User       │
├──────────────┬───────────────────────────────────────────┤
│ Dashboard    │                                           │
│              │              PAGE CONTENT                 │
│ Operations   │                                           │
│  Shipments   │                                           │
│  Pickup      │                                           │
│  Delivery    │                                           │
│  Dispatch    │                                           │
│              │                                           │
│ Logistics    │                                           │
│  Hub         │                                           │
│  Transfer    │                                           │
│  Returns     │                                           │
│              │                                           │
│ Finance      │                                           │
│  COD         │                                           │
│  Settlement  │                                           │
│              │                                           │
│ Business     │                                           │
│  Merchants   │                                           │
│  Reports     │                                           │
│              │                                           │
│ System       │                                           │
│  Users       │                                           │
│  Roles       │                                           │
│  Settings    │                                           │
└──────────────┴───────────────────────────────────────────┘
```

## UI principles

- Clear hierarchy
- Consistent spacing
- Consistent status system
- Responsive
- Accessible
- Loading state
- Empty state
- Error state
- Confirmation for destructive action
- Bulk action where appropriate
- Keyboard-friendly admin workflow
- Avoid unnecessary modal nesting
- Avoid information overload

---

# 28. SHIPMENT DETAIL UI

Recommended structure:

```text
Header
 ├── Tracking number
 ├── Current status
 ├── Actions
 └── More

Summary
 ├── Sender
 ├── Receiver
 ├── Package
 └── Financial

Tracking Timeline

Delivery Attempts

Assignment

Return

Documents / Proof

Audit History
```

---

# 29. API REQUIREMENTS

Minimum API concepts:

```text
POST   /shipments
GET    /shipments
GET    /shipments/{id}
PATCH  /shipments/{id}
POST   /shipments/{id}/cancel

GET    /shipments/{id}/tracking

POST   /shipments/{id}/assign
POST   /shipments/{id}/pickup
POST   /shipments/{id}/delivery
POST   /shipments/{id}/delivery-attempts

POST   /shipments/{id}/return

GET    /merchants
GET    /customers
GET    /shippers
GET    /hubs

GET    /cod
GET    /settlements

GET    /reports
```

**Không được tạo endpoint trùng chức năng chỉ vì frontend cần cách gọi khác.**

---

# 30. WEBHOOK / EVENT

Nếu có integration:

```text
shipment.created
shipment.updated
shipment.status_changed
shipment.delivered
shipment.failed
shipment.returned
settlement.created
settlement.paid
```

Webhook cần:

- signature/security
- retry
- idempotency
- delivery log
- failure handling

---

# 31. DATABASE AUDIT

Agent phải kiểm tra entity relationship.

Baseline:

```text
users
roles
permissions

merchants
customers
addresses

shipments
shipment_items
shipment_events
shipment_status_histories

drivers
driver_assignments
delivery_attempts

hubs
hub_shipments
hub_transfers

returns
return_items

cod_transactions
payments
settlements

notifications
audit_logs
```

Không bắt buộc project phải giống 100%. Agent phải map schema hiện tại vào business requirement.

---

# 32. DATA INTEGRITY

Agent phải kiểm tra:

- Foreign keys
- Unique tracking number
- Unique order number nếu business yêu cầu
- Decimal cho tiền
- Integer/decimal phù hợp cho weight
- Timestamp
- Timezone
- Soft delete nếu cần
- Index
- Pagination
- Transaction
- Concurrency
- Race condition
- Idempotency

Đặc biệt:

### Money

Không dùng floating-point tùy tiện cho tiền.

### Status transition

Không cho phép cập nhật status tùy ý từ frontend.

---

# 33. SECURITY

Audit:

- Authentication
- Authorization
- CSRF nếu applicable
- CORS
- Rate limiting
- Input validation
- SQL injection
- XSS
- File upload validation
- Sensitive data exposure
- API token security
- Password security
- Permission escalation

Không expose:

- Internal notes
- Sensitive financial data
- Admin-only data
- Private customer data

cho public tracking.

---

# 34. PERFORMANCE

Kiểm tra:

- N+1 queries
- Pagination
- Indexes
- Eager loading
- Heavy dashboard queries
- Large exports
- Queue jobs
- Notification queue
- Image/file optimization
- API response size

Không load toàn bộ shipment vào frontend chỉ để paginate bằng JavaScript nếu dữ liệu lớn.

---

# 35. FRONTEND AUDIT

Kiểm tra:

```text
Routes
Pages
Components
Layouts
Stores
Composables
API services
Types/interfaces
Guards
Permissions
Loading
Errors
Empty states
Form validation
```

Không để:

- API URL hard-code
- token hard-code
- fake data trong production
- status hard-code rải rác
- duplicate API service
- duplicate business logic

---

# 36. BACKEND AUDIT

Kiểm tra:

```text
Routes
Controllers
Requests / Validators
Services
Models
Repositories nếu project dùng
Policies
Resources / Transformers
Jobs
Events
Listeners
Notifications
Observers
Migrations
Seeders
Tests
```

Business logic quan trọng không nên nhét hết vào Controller.

---

# 37. ERROR HANDLING

API phải có response nhất quán.

Ví dụ:

```json
{
  "success": false,
  "message": "Unable to create shipment",
  "errors": {
    "receiver_phone": [
      "The receiver phone is invalid."
    ]
  }
}
```

Frontend phải xử lý:

```text
Success
Loading
Validation Error
Unauthorized
Forbidden
Not Found
Conflict
Server Error
Network Error
```

---

# 38. TESTING

## Unit Test

Test:

- status transition
- fee calculation
- COD calculation
- settlement
- permission
- delivery attempt
- return rules

## Feature/API Test

Test:

- create shipment
- update shipment
- assign shipper
- pickup
- delivery
- failed delivery
- return
- COD
- settlement

## E2E

Test flow:

```text
Merchant creates shipment
      ↓
Admin confirms
      ↓
Dispatcher assigns shipper
      ↓
Shipper picks up
      ↓
Hub processes
      ↓
Shipper delivers
      ↓
Customer receives
      ↓
COD reconciled
      ↓
Merchant settlement
```

---

# 39. BUSINESS RULES TO VERIFY

Agent phải tìm trong code/documentation và xác định rõ:

- Khi nào shipment được cancel?
- Ai được cancel?
- Khi nào shipment được edit?
- Khi nào shipment được assign?
- Có được reassign không?
- Bao nhiêu delivery attempts?
- Sau bao nhiêu failed attempt thì return?
- Ai chịu return fee?
- COD tính như thế nào?
- Shipping fee tính như thế nào?
- Khi nào merchant nhận COD?
- Settlement theo ngày/tuần/tháng?
- Có SLA không?
- Có working hours không?
- Có vùng cấm giao không?
- Có giới hạn cân nặng/kích thước không?

Nếu code đang có business rule nhưng tài liệu chưa ghi, **document rule hiện tại trước khi thay đổi**.

---

# 40. AGENT AUDIT CHECKLIST

Agent phải tạo bảng kết quả:

| Module | Status | Existing | Missing | Bug | Priority | Action |
|---|---|---|---|---|---|---|
| Dashboard | 🟡 | Partial | KPI | Yes | P1 | Fix |
| Shipment | 🟢 | Yes | - | - | - | Keep |
| Tracking | 🟡 | Partial | Events | Yes | P0 | Fix |
| COD | 🔴 | Missing | Full | - | P0 | Build |
| Return | 🟡 | Partial | Flow | Yes | P1 | Complete |

Status:

```text
🟢 COMPLETE
🟡 PARTIAL
🔴 MISSING
⚠️ BUG
🔵 NEEDS REVIEW
```

Priority:

```text
P0 = Critical
P1 = High
P2 = Medium
P3 = Low
```

---

# 41. DEFINITION OF DONE

Một module chỉ được đánh dấu COMPLETE khi:

- [ ] UI hoàn chỉnh
- [ ] API hoàn chỉnh
- [ ] Database đúng
- [ ] Validation
- [ ] Authorization
- [ ] Error handling
- [ ] Loading state
- [ ] Empty state
- [ ] Responsive
- [ ] Business rules
- [ ] Audit nếu cần
- [ ] Tests
- [ ] No console error
- [ ] No obvious N+1
- [ ] No fake data
- [ ] No hard-coded business logic
- [ ] Existing features still work

---

# 42. AGENT EXECUTION PLAN

Không sửa toàn bộ cùng lúc.

## Phase 1 — Discovery

```text
Scan repository
↓
Map architecture
↓
Map database
↓
Map API
↓
Map frontend
↓
Map existing features
```

Output:

```text
PROJECT_AUDIT.md
```

---

## Phase 2 — Gap Analysis

Đối chiếu với document này.

Output:

```text
COMPLETE
PARTIAL
MISSING
BUG
UNKNOWN
```

---

## Phase 3 — Core Fix

Ưu tiên:

```text
P0
Authentication
Authorization
Shipment
Status
Tracking
Database integrity
```

---

## Phase 4 — Operations

```text
Pickup
Assignment
Shipper
Hub
Delivery
Delivery Attempts
Return
```

---

## Phase 5 — Finance

```text
COD
Transactions
Reconciliation
Settlement
```

---

## Phase 6 — Analytics

```text
Dashboard
Reports
SLA
Performance
```

---

## Phase 7 — Platform

```text
Notifications
Audit Logs
API
Webhook
Settings
```

---

## Phase 8 — QA

```text
Unit
Feature/API
E2E
Security
Performance
Responsive
Regression
```

---

# 43. IMPORTANT AGENT BEHAVIOR

Agent KHÔNG được:

1. Chỉ nhìn screenshot rồi đoán backend.
2. Chỉ kiểm tra frontend.
3. Chỉ kiểm tra database.
4. Tự tạo mock data thay cho business logic.
5. Đánh dấu feature COMPLETE chỉ vì UI tồn tại.
6. Đánh dấu API COMPLETE chỉ vì endpoint trả 200.
7. Bỏ qua permission.
8. Bỏ qua edge case.
9. Rewrite code không cần thiết.
10. Xóa feature cũ để làm feature mới.
11. Chạy migration destructive mà không kiểm tra.
12. Thay đổi database production mà không cảnh báo.
13. Hard-code trạng thái ở nhiều nơi.
14. Dùng frontend để enforce security.
15. Tự suy diễn business rule khi chưa kiểm tra code/documentation.

---

# 44. OUTPUT FORMAT BẮT BUỘC CỦA AGENT

Sau mỗi phase, Agent phải báo:

## A. Inspected

```text
Files:
- ...
- ...

Modules:
- ...
```

## B. Findings

```text
Complete:
- ...

Partial:
- ...

Missing:
- ...

Bug:
- ...
```

## C. Changes

```text
Changed:
- ...

Created:
- ...

Deleted:
- ...
```

## D. Tests

```text
Passed:
- ...

Failed:
- ...

Not tested:
- ...
```

## E. Remaining Work

```text
P0:
- ...

P1:
- ...

P2:
- ...
```

---

# 45. FINAL SYSTEM ACCEPTANCE CHECKLIST

## Core

- [ ] Authentication
- [ ] RBAC
- [ ] Dashboard
- [ ] Shipment CRUD
- [ ] Tracking
- [ ] Pickup
- [ ] Delivery
- [ ] Shipper
- [ ] Assignment
- [ ] Hub
- [ ] Return
- [ ] COD
- [ ] Settlement
- [ ] Merchant
- [ ] Customer
- [ ] Notification
- [ ] Audit log
- [ ] Reports
- [ ] Settings

## Quality

- [ ] Validation
- [ ] Authorization
- [ ] Error handling
- [ ] Loading state
- [ ] Empty state
- [ ] Responsive
- [ ] Security
- [ ] Performance
- [ ] Unit tests
- [ ] API tests
- [ ] E2E tests
- [ ] Regression test

## Production readiness

- [ ] No critical bugs
- [ ] No broken core workflow
- [ ] No fake production data
- [ ] No exposed secret
- [ ] Database integrity verified
- [ ] API contract verified
- [ ] Permission verified
- [ ] Logging verified
- [ ] Backup/recovery considered
- [ ] Deployment verified

---

# 46. FINAL INSTRUCTION TO AI AGENT

Bạn đang tiếp quản một **shipment management system đang phát triển dở**, không phải bắt đầu một project mới.

Mục tiêu của bạn:

> **Hiểu hệ thống hiện tại trước, sau đó hoàn thiện nó theo business specification này mà không phá vỡ những gì đang hoạt động.**

Thứ tự bắt buộc:

```text
READ
 ↓
SCAN
 ↓
UNDERSTAND
 ↓
MAP
 ↓
AUDIT
 ↓
REPORT GAPS
 ↓
PRIORITIZE
 ↓
IMPLEMENT
 ↓
TEST
 ↓
VERIFY
 ↓
REPORT
```

Không được nhảy thẳng từ `READ` sang `IMPLEMENT`.

Nếu phát hiện requirement chưa rõ:

1. Kiểm tra database.
2. Kiểm tra API.
3. Kiểm tra business logic hiện tại.
4. Kiểm tra UI/UX hiện tại.
5. Kiểm tra test.
6. Chỉ sau đó mới quyết định.

Nếu vẫn chưa xác định được, đánh dấu:

```text
UNKNOWN / NEEDS BUSINESS DECISION
```

Không tự ý bịa business rule.

---

# 47. PROJECT HEALTH SCORE

Sau khi audit, Agent phải đánh giá:

```text
Architecture       /10
Database           /10
Backend            /10
Frontend           /10
Business Logic     /10
Security           /10
UX                 /10
Testing            /10
Performance        /10
Production Ready   /10
```

Sau đó đưa ra:

```text
OVERALL SCORE: XX/100
```

và kết luận:

```text
NOT READY
        hoặc
NEAR PRODUCTION
        hoặc
PRODUCTION READY
```

Kèm danh sách blocker P0/P1.

---

# 48. END GOAL

Hệ thống hoàn chỉnh phải cho phép:

```text
MERCHANT
   ↓
CREATE SHIPMENT
   ↓
VALIDATE
   ↓
PICKUP
   ↓
HUB
   ↓
TRANSIT
   ↓
DESTINATION HUB
   ↓
ASSIGN SHIPPER
   ↓
OUT FOR DELIVERY
   ↓
┌───────────────────────┐
│                       │
DELIVERED            FAILED
│                       │
↓                    RETRY
COD                     │
│                       ↓
↓                    RETURN
RECONCILIATION           │
│                       ↓
SETTLEMENT            RETURNED
│
↓
REPORTING
```

**Mục tiêu cuối cùng không phải có nhiều màn hình.**

Mục tiêu là đảm bảo:

> **Mọi shipment đều có trạng thái rõ ràng, lịch sử truy vết được, người chịu trách nhiệm rõ ràng, tiền/COD đối soát được, exception xử lý được, và toàn bộ lifecycle có thể audit.**
