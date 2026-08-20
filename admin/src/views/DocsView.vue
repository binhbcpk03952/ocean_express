<template>
  <div class="bg-[#f8fafc] text-slate-800 font-sans antialiased min-h-screen selection:bg-sky-200">
    <!-- Header -->
    <header class="bg-white/80 backdrop-blur-md shadow-sm sticky top-0 z-50 border-b border-slate-200/60">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <div class="flex items-center gap-3">
            <div class="bg-gradient-to-br from-sky-400 to-blue-600 p-2 rounded-xl shadow-sm">
              <Package class="w-6 h-6 text-white" />
            </div>
            <h1 class="text-xl font-bold text-slate-900 tracking-tight">
              Ocean Express <span class="text-xs font-semibold px-2 py-0.5 rounded-full bg-slate-100 text-slate-500 ml-2 border border-slate-200 uppercase tracking-wider">Docs</span>
            </h1>
          </div>
          <div>
            <router-link
              to="/"
              class="group flex items-center gap-1.5 text-sm font-medium text-slate-500 hover:text-sky-600 transition-colors bg-white hover:bg-sky-50 px-4 py-2 rounded-full border border-slate-200 hover:border-sky-200"
            >
              Về trang Quản trị 
              <ArrowRight class="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
            </router-link>
          </div>
        </div>
      </div>
    </header>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 flex flex-col lg:flex-row gap-10">
      <!-- Sidebar Navigation -->
      <aside class="w-full lg:w-72 flex-shrink-0">
        <div class="sticky top-24">
          <h3 class="text-xs font-bold text-slate-400 uppercase tracking-wider mb-4 px-3">Nội dung hướng dẫn</h3>
          <nav class="flex flex-col gap-1.5">
            <a
              href="#shop"
              class="nav-link flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200"
              :class="activeSection === 'shop' ? 'bg-white text-sky-700 shadow-sm ring-1 ring-slate-900/5' : 'text-slate-600 hover:bg-slate-100/80 hover:text-slate-900'"
              @click="scrollTo('shop')"
            >
              <Store class="w-4 h-4" :class="activeSection === 'shop' ? 'text-sky-500' : 'text-slate-400'" />
              Dành cho Chủ Shop
            </a>
            <a
              href="#api"
              class="nav-link flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200"
              :class="activeSection === 'api' ? 'bg-white text-cyan-700 shadow-sm ring-1 ring-slate-900/5' : 'text-slate-600 hover:bg-slate-100/80 hover:text-slate-900'"
              @click="scrollTo('api')"
            >
              <Code class="w-4 h-4" :class="activeSection === 'api' ? 'text-cyan-500' : 'text-slate-400'" />
              Tích hợp API & In vận đơn
            </a>
            <a
              href="#hub"
              class="nav-link flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200"
              :class="activeSection === 'hub' ? 'bg-white text-emerald-700 shadow-sm ring-1 ring-slate-900/5' : 'text-slate-600 hover:bg-slate-100/80 hover:text-slate-900'"
              @click="scrollTo('hub')"
            >
              <Warehouse class="w-4 h-4" :class="activeSection === 'hub' ? 'text-emerald-500' : 'text-slate-400'" />
              Dành cho Nhân viên Kho
            </a>
            <a
              href="#driver"
              class="nav-link flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200"
              :class="activeSection === 'driver' ? 'bg-white text-orange-700 shadow-sm ring-1 ring-slate-900/5' : 'text-slate-600 hover:bg-slate-100/80 hover:text-slate-900'"
              @click="scrollTo('driver')"
            >
              <Bike class="w-4 h-4" :class="activeSection === 'driver' ? 'text-orange-500' : 'text-slate-400'" />
              Dành cho Tài xế
            </a>
            <a
              href="#admin"
              class="nav-link flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200"
              :class="activeSection === 'admin' ? 'bg-white text-purple-700 shadow-sm ring-1 ring-slate-900/5' : 'text-slate-600 hover:bg-slate-100/80 hover:text-slate-900'"
              @click="scrollTo('admin')"
            >
              <ShieldCheck class="w-4 h-4" :class="activeSection === 'admin' ? 'text-purple-500' : 'text-slate-400'" />
              Dành cho Quản trị viên
            </a>
          </nav>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="flex-1 max-w-4xl min-w-0">
        <div class="mb-14">
          <h2 class="text-4xl sm:text-5xl font-extrabold text-slate-900 tracking-tight mb-6">Tài liệu & Hướng dẫn</h2>
          <p class="text-lg text-slate-500 leading-relaxed max-w-3xl">
            Tài liệu này cung cấp hướng dẫn chi tiết từng bước để sử dụng nền tảng Logistics Ocean Express. Chọn vai trò của bạn ở menu bên trái để bắt đầu.
          </p>
        </div>

        <!-- SECTION: SHOP -->
        <section id="shop" class="mb-20 scroll-mt-32">
          <div class="flex items-center gap-4 mb-8 pb-4 border-b border-slate-200/60">
            <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-blue-100 to-sky-100 border border-blue-200/50 flex items-center justify-center text-blue-600 shadow-sm">
              <Store class="w-6 h-6" />
            </div>
            <h3 class="text-3xl font-bold text-slate-900 tracking-tight">Dành cho Khách Hàng (Chủ Shop)</h3>
          </div>
          <div class="prose max-w-none text-slate-600 space-y-6 text-base leading-relaxed">
            <p>Với vai trò là Chủ Shop, bạn có thể tự tạo vận đơn, theo dõi lộ trình giao hàng và quản lý doanh thu đối soát (COD).</p>

            <div class="bg-blue-50/80 border border-blue-100/80 rounded-2xl p-5 flex gap-4">
              <Info class="w-6 h-6 text-blue-500 flex-shrink-0" />
              <div>
                <p class="font-semibold text-blue-900 mb-1">Mẹo truy cập nhanh</p>
                <p class="text-blue-800 text-sm">Truy cập Cổng quản lý dành cho Shop tại tên miền chính của hệ thống. Đăng nhập bằng email và mật khẩu được cấp.</p>
              </div>
            </div>

            <div class="bg-white rounded-2xl p-8 shadow-sm border border-slate-200/60 mt-8">
              <h4 class="text-xl font-bold text-slate-900 mb-4">Cách tạo Vận đơn (Shipping Order)</h4>
              <ul class="space-y-3 list-none p-0">
                <li class="flex gap-3">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center text-sm font-semibold">1</span>
                  <span>Đăng nhập vào hệ thống bằng tài khoản của Shop.</span>
                </li>
                <li class="flex gap-3">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center text-sm font-semibold">2</span>
                  <span>Vào mục <strong>Tạo đơn mới</strong>. Nhập thông tin người nhận: Tên, Số điện thoại, Địa chỉ chi tiết.</span>
                </li>
                <li class="flex gap-3">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center text-sm font-semibold">3</span>
                  <span>Chọn <strong>Khu vực nhận hàng</strong> để hệ thống tính cước phí tự động.</span>
                </li>
                <li class="flex gap-3">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center text-sm font-semibold">4</span>
                  <span>Nhập thông tin kiện hàng: Khối lượng (g), kích thước và tiền thu hộ (COD) nếu có.</span>
                </li>
                <li class="flex gap-3">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center text-sm font-semibold">5</span>
                  <span>Bấm <strong>Tạo vận đơn</strong>. Mã vận đơn dạng <code>OE-XXXXXX</code> sẽ được sinh tự động. Bạn có thể in tem nhãn trực tiếp bằng nút In vận đơn.</span>
                </li>
              </ul>
            </div>
          </div>
        </section>

        <!-- SECTION: API & DEVELOPER -->
        <section id="api" class="mb-20 scroll-mt-32">
          <div class="flex items-center gap-4 mb-8 pb-4 border-b border-slate-200/60">
            <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-cyan-100 to-sky-100 border border-cyan-200/50 flex items-center justify-center text-cyan-600 shadow-sm">
              <Code class="w-6 h-6" />
            </div>
            <h3 class="text-3xl font-bold text-slate-900 tracking-tight">Tích hợp API & In Vận Đơn (Developer)</h3>
          </div>

          <div class="prose max-w-none text-slate-600 space-y-6 text-base leading-relaxed">
            <p>Ocean Express cung cấp hệ thống REST API hoàn chỉnh để các sàn TMĐT, Website bán hàng (WooCommerce, Shopify, Laravel, Haravan...) hoặc phần mềm ERP kết nối tự động tạo đơn, tính cước và in tem vận đơn.</p>

            <div class="bg-slate-900 text-slate-300 rounded-2xl p-6 sm:p-8 shadow-md border border-slate-800">
              <div class="space-y-8">
                <!-- 1. Header auth -->
                <div>
                  <h5 class="text-lg font-bold text-white mb-2 flex items-center gap-2">
                    <span class="text-sky-400 text-sm">01.</span> Xác thực qua X-API-Key (Header)
                  </h5>
                  <p class="text-sm mb-3">Lấy API Key tại mục <strong>Thông tin tài khoản</strong> trên Portal. Gắn vào HTTP Header của tất cả request:</p>
                  <pre class="bg-slate-950 p-4 rounded-xl text-sm font-mono text-emerald-400 overflow-x-auto border border-slate-800">X-API-Key: oe_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx</pre>
                </div>

                <!-- 2. In tem vận đơn PDF & JSON -->
                <div>
                  <h5 class="text-lg font-bold text-white mb-2 flex items-center gap-2">
                    <span class="text-sky-400 text-sm">02.</span> In Tem Vận Đơn (A6/A7 Standard)
                  </h5>
                  <p class="text-sm mb-3">Hỗ trợ 2 cách in tem: Trực tiếp lấy stream file PDF (để mở xem/in) hoặc gọi API JSON lấy đường dẫn in:</p>
                  
                  <div class="space-y-4">
                    <div>
                      <div class="text-xs font-semibold text-sky-400 mb-1">Cách 1: Lấy file PDF trực tiếp (Stream)</div>
                      <div class="flex items-center gap-2 mb-2">
                        <span class="px-2 py-0.5 rounded text-xs font-bold bg-blue-500/20 text-blue-400 border border-blue-500/30">GET</span>
                        <code class="text-sky-300 text-sm font-mono">https://api.oceanexpress.bcbdev.id.vn/api/v1/orders/{order_id}/label</code>
                      </div>
                      <pre class="bg-slate-950 p-4 rounded-xl text-xs font-mono text-slate-300 overflow-x-auto border border-slate-800">curl -X GET "https://api.oceanexpress.bcbdev.id.vn/api/v1/orders/12345/label" \
  -H "X-API-Key: {YOUR_API_KEY}"</pre>
                    </div>

                    <div>
                      <div class="text-xs font-semibold text-emerald-400 mb-1">Cách 2: Lấy Link in vận đơn qua JSON API (Dành cho Laravel / Web App)</div>
                      <div class="flex items-center gap-2 mb-2">
                        <span class="px-2 py-0.5 rounded text-xs font-bold bg-green-500/20 text-green-400 border border-green-500/30">POST</span>
                        <code class="text-sky-300 text-sm font-mono">https://api.oceanexpress.bcbdev.id.vn/api/v1/orders/print-label</code>
                      </div>
                      <pre class="bg-slate-950 p-4 rounded-xl text-xs font-mono text-slate-300 overflow-x-auto border border-slate-800">curl -X POST "https://api.oceanexpress.bcbdev.id.vn/api/v1/orders/print-label" \
  -H "X-API-Key: {YOUR_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"order_id": "12345"}'</pre>
                      <p class="text-xs mt-2 text-slate-400">Response JSON:</p>
                      <pre class="bg-slate-950 p-4 rounded-xl text-xs font-mono text-emerald-400 overflow-x-auto border border-slate-800">{
  "code": 200,
  "status": "success",
  "data": {
    "order_id": "8f8b3c10-...",
    "tracking_number": "OE-100234",
    "label_url": "https://api.oceanexpress.bcbdev.id.vn/api/v1/public/orders/8f8b3c10-.../label",
    "tracking_url": "https://oceanexpress.bcbdev.id.vn/tracking?code=OE-100234"
  }
}</pre>
                    </div>
                  </div>
                </div>

                <!-- 3. Tạo đơn -->
                <div>
                  <h5 class="text-lg font-bold text-white mb-2 flex items-center gap-2">
                    <span class="text-sky-400 text-sm">03.</span> Tạo Vận Đơn Tự Động
                  </h5>
                  <div class="flex items-center gap-2 mb-2">
                    <span class="px-2 py-0.5 rounded text-xs font-bold bg-green-500/20 text-green-400 border border-green-500/30">POST</span>
                    <code class="text-sky-300 text-sm font-mono">https://api.oceanexpress.bcbdev.id.vn/api/v1/orders</code>
                  </div>
                  <pre class="bg-slate-950 p-4 rounded-xl text-xs font-mono text-slate-300 overflow-x-auto border border-slate-800">{
  "receiver_name": "Nguyễn Văn A",
  "receiver_phone": "0987654321",
  "receiver_location_id": "VN-HN-CG",
  "receiver_address_detail": "123 Đường Cầu Giấy, Phường Dịch Vọng",
  "weight": 1500,
  "length": 20,
  "width": 15,
  "height": 10,
  "cod_amount": 350000,
  "note": "Cho khách xem hàng"
}</pre>
                </div>

                <!-- 4. Tính cước -->
                <div>
                  <h5 class="text-lg font-bold text-white mb-2 flex items-center gap-2">
                    <span class="text-sky-400 text-sm">04.</span> Tính Trước Cước Phí Vận Chuyển
                  </h5>
                  <div class="flex items-center gap-2 mb-2">
                    <span class="px-2 py-0.5 rounded text-xs font-bold bg-green-500/20 text-green-400 border border-green-500/30">POST</span>
                    <code class="text-sky-300 text-sm font-mono">https://api.oceanexpress.bcbdev.id.vn/api/v1/rates/calculate</code>
                  </div>
                  <pre class="bg-slate-950 p-4 rounded-xl text-xs font-mono text-slate-300 overflow-x-auto border border-slate-800">{
  "receiver_location_id": "VN-HN-CG", 
  "weight": 1500
}</pre>
                </div>

                <!-- 5. In hàng loạt -->
                <div>
                  <h5 class="text-lg font-bold text-white mb-2 flex items-center gap-2">
                    <span class="text-sky-400 text-sm">05.</span> In Hàng Loạt Tem Vận Đơn (Batch PDF)
                  </h5>
                  <div class="flex items-center gap-2 mb-2">
                    <span class="px-2 py-0.5 rounded text-xs font-bold bg-green-500/20 text-green-400 border border-green-500/30">POST</span>
                    <code class="text-sky-300 text-sm font-mono">https://api.oceanexpress.bcbdev.id.vn/api/v1/orders/labels/batch</code>
                  </div>
                  <pre class="bg-slate-950 p-4 rounded-xl text-xs font-mono text-slate-300 overflow-x-auto border border-slate-800">{
  "order_ids": ["uuid-1", "uuid-2", "OE-100234"]
}</pre>
                </div>

                <!-- 6. Webhook -->
                <div>
                  <h5 class="text-lg font-bold text-white mb-2 flex items-center gap-2">
                    <span class="text-sky-400 text-sm">06.</span> Cấu hình Webhook Thông Báo Trạng Thái Đơn
                  </h5>
                  <p class="text-sm mb-3">Tại mục <strong>Thông tin tài khoản</strong>, cập nhật <code>Webhook URL</code> của hệ thống bạn. Khi trạng thái đơn thay đổi (Vd: Giao thành công, Đang giao hàng, Hoàn hàng), Ocean Express sẽ tự động POST dữ liệu JSON cập nhật về URL của bạn.</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- SECTION: HUB STAFF -->
        <section id="hub" class="mb-20 scroll-mt-32">
          <div class="flex items-center gap-4 mb-8 pb-4 border-b border-slate-200/60">
            <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-emerald-100 to-green-100 border border-emerald-200/50 flex items-center justify-center text-emerald-600 shadow-sm">
              <Warehouse class="w-6 h-6" />
            </div>
            <h3 class="text-3xl font-bold text-slate-900 tracking-tight">Dành cho Nhân viên Kho</h3>
          </div>
          <div class="prose max-w-none text-slate-600 space-y-6 text-base leading-relaxed">
            <p>Nhân viên kho đóng vai trò là trạm kiểm soát trung chuyển, thực hiện quét mã vạch để ghi nhận hàng vào/ra khỏi kho.</p>

            <div class="bg-amber-50/80 border border-amber-200/60 rounded-2xl p-5 flex gap-4">
              <AlertTriangle class="w-6 h-6 text-amber-500 flex-shrink-0" />
              <div>
                <p class="font-semibold text-amber-900 mb-1">Lưu ý Quan trọng</p>
                <p class="text-amber-800 text-sm">Chỉ thực hiện quét Inbound/Outbound đối với các đơn hàng <strong>vật lý</strong> đang cầm trên tay tại kho của bạn để tránh lệch tồn kho hệ thống.</p>
              </div>
            </div>

            <div class="bg-white rounded-2xl p-8 shadow-sm border border-slate-200/60 mt-8">
              <h4 class="text-xl font-bold text-slate-900 mb-4">Quy trình Nhập kho (Inbound)</h4>
              <ul class="space-y-3 list-none p-0">
                <li class="flex gap-3">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center text-sm font-semibold">1</span>
                  <span>Khi Tài xế mang hàng về kho, nhân viên kho mở tính năng <strong class="text-slate-800">Quét mã vạch</strong>.</span>
                </li>
                <li class="flex gap-3">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center text-sm font-semibold">2</span>
                  <span>Dùng máy quét (barcode scanner) hoặc nhập tay Tracking Number.</span>
                </li>
                <li class="flex gap-3">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center text-sm font-semibold">3</span>
                  <span>Bấm xác nhận <strong class="text-slate-800 text-emerald-600">Inbound</strong>. Trạng thái đơn sẽ cập nhật thành <code>Đã nhập kho</code>.</span>
                </li>
              </ul>
            </div>
          </div>
        </section>

        <!-- SECTION: DRIVER -->
        <section id="driver" class="mb-20 scroll-mt-32">
          <div class="flex items-center gap-4 mb-8 pb-4 border-b border-slate-200/60">
            <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-orange-100 to-amber-100 border border-orange-200/50 flex items-center justify-center text-orange-600 shadow-sm">
              <Bike class="w-6 h-6" />
            </div>
            <h3 class="text-3xl font-bold text-slate-900 tracking-tight">Dành cho Tài xế (Shipper)</h3>
          </div>
          <div class="prose max-w-none text-slate-600 space-y-6 text-base leading-relaxed">
            <p>Tài xế (bao gồm lấy hàng và giao hàng) sử dụng ứng dụng để cập nhật trạng thái lộ trình thực tế của gói hàng.</p>

            <div class="grid md:grid-cols-2 gap-6 mt-8">
              <div class="bg-white border border-slate-200/60 rounded-2xl p-6 shadow-sm hover:shadow-md transition-shadow">
                <div class="w-10 h-10 rounded-full bg-orange-50 flex items-center justify-center mb-4">
                  <Package class="w-5 h-5 text-orange-500" />
                </div>
                <h4 class="text-lg font-bold text-slate-900 mb-2">Lấy hàng (First-mile)</h4>
                <p class="text-sm text-slate-500 leading-relaxed">Đến Shop theo định vị GPS. Nhận hàng vật lý từ Shop và cập nhật trạng thái thành <strong class="text-slate-700">Đã lấy hàng (Picked Up)</strong>. Sau đó chở hàng về Kho tổng.</p>
              </div>
              <div class="bg-white border border-slate-200/60 rounded-2xl p-6 shadow-sm hover:shadow-md transition-shadow">
                <div class="w-10 h-10 rounded-full bg-sky-50 flex items-center justify-center mb-4">
                  <MapPin class="w-5 h-5 text-sky-500" />
                </div>
                <h4 class="text-lg font-bold text-slate-900 mb-2">Giao hàng (Last-mile)</h4>
                <p class="text-sm text-slate-500 leading-relaxed">Nhận hàng xuất kho. Chạy xe đến địa chỉ người nhận. Giao thành công thu tiền COD và cập nhật trạng thái <strong class="text-slate-700">Giao thành công (Delivered)</strong>.</p>
              </div>
            </div>
          </div>
        </section>

        <!-- SECTION: ADMIN -->
        <section id="admin" class="mb-8 scroll-mt-32">
          <div class="flex items-center gap-4 mb-8 pb-4 border-b border-slate-200/60">
            <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-purple-100 to-fuchsia-100 border border-purple-200/50 flex items-center justify-center text-purple-600 shadow-sm">
              <ShieldCheck class="w-6 h-6" />
            </div>
            <h3 class="text-3xl font-bold text-slate-900 tracking-tight">Dành cho Quản trị viên</h3>
          </div>
          <div class="prose max-w-none text-slate-600 space-y-6 text-base leading-relaxed">
            <p>Admin có toàn quyền quản lý bảng giá, mạng lưới bưu cục và kiểm duyệt đối tác kinh doanh.</p>

            <div class="bg-white rounded-2xl p-8 shadow-sm border border-slate-200/60 mt-8">
              <h4 class="text-xl font-bold text-slate-900 mb-3">Cấu hình Cước phí (Shipping Rates)</h4>
              <p class="mb-6 text-slate-500">Truy cập mục <strong>Bảng Giá</strong> để thiết lập cước phí vận chuyển. Bạn có thể định cấu hình phí gửi từ khu vực A sang khu vực B, quy định mức giá cho 1000g đầu tiên và bước giá phụ thu.</p>

              <h4 class="text-xl font-bold text-slate-900 mb-4 pt-6 border-t border-slate-100">Quản lý Mạng lưới</h4>
              <ul class="space-y-4 list-none p-0">
                <li class="flex items-start gap-3">
                  <div class="w-1.5 h-1.5 rounded-full bg-purple-500 mt-2 flex-shrink-0"></div>
                  <div>
                    <strong class="text-slate-800">Tạo Kho (Hub):</strong> 
                    <span class="text-slate-500 block mt-0.5">Định vị chính xác các bưu cục trên bản đồ số để hệ thống điều hướng đơn hàng tự động.</span>
                  </div>
                </li>
                <li class="flex items-start gap-3">
                  <div class="w-1.5 h-1.5 rounded-full bg-purple-500 mt-2 flex-shrink-0"></div>
                  <div>
                    <strong class="text-slate-800">Duyệt Shop:</strong> 
                    <span class="text-slate-500 block mt-0.5">Kích hoạt tài khoản cho các Đối tác kinh doanh đăng ký tham gia mạng lưới.</span>
                  </div>
                </li>
                <li class="flex items-start gap-3">
                  <div class="w-1.5 h-1.5 rounded-full bg-purple-500 mt-2 flex-shrink-0"></div>
                  <div>
                    <strong class="text-slate-800">Nhân sự:</strong> 
                    <span class="text-slate-500 block mt-0.5">Phân quyền tài khoản (Tài xế / Thủ kho) trực thuộc các bưu cục cụ thể.</span>
                  </div>
                </li>
              </ul>
            </div>
          </div>
        </section>

      </main>
    </div>

    <footer class="bg-white border-t border-slate-200/60 py-10 mt-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <p class="text-sm font-medium text-slate-900">&copy; 2026 Ocean Express Logistics. Tất cả các quyền được bảo lưu.</p>
        <p class="mt-2 text-sm text-slate-500">Nền tảng được thiết kế chuyên nghiệp bởi Hệ thống hỗ trợ tích hợp.</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { 
  Package, 
  ArrowRight, 
  Store, 
  Info, 
  Warehouse, 
  AlertTriangle, 
  Bike, 
  MapPin, 
  ShieldCheck,
  Code
} from 'lucide-vue-next';

const activeSection = ref('shop');

const scrollTo = (id) => {
  activeSection.value = id;
  const el = document.getElementById(id);
  if (el) {
    window.scrollTo({
      top: el.offsetTop - 100,
      behavior: 'smooth'
    });
  }
};

const handleScroll = () => {
  const sections = ['shop', 'api', 'hub', 'driver', 'admin'];
  for (const id of sections) {
    const el = document.getElementById(id);
    if (el && window.scrollY >= el.offsetTop - 150) {
      activeSection.value = id;
    }
  }
};

onMounted(() => {
  window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
});
</script>
