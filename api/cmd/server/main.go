package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	httpDelivery "ocean-express-api/internal/delivery/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"
	"ocean-express-api/internal/repository"
	"ocean-express-api/internal/usecase"
	"ocean-express-api/pkg/geocoding"
	"ocean-express-api/pkg/notification"
	"ocean-express-api/pkg/utils"
	"ocean-express-api/pkg/webhook"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Setup Database Connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "rootpassword")
	dbName := getEnv("DB_NAME", "ocean_express_db")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
		dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected successfully!")

	// Migration idempotent: init.sql chỉ chạy khi DB tạo mới, nên với DB đang chạy
	// ta bổ sung cột current_hub_id (nếu chưa có) mà không cần xóa dữ liệu.
	if err := db.Exec("ALTER TABLE shipping_orders ADD COLUMN IF NOT EXISTS current_hub_id UUID REFERENCES hubs(id)").Error; err != nil {
		log.Printf("Cảnh báo: không thể chạy migration current_hub_id: %v", err)
	}

	// UNIQUE trên fcm_token để Upsert thiết bị (OnConflict theo fcm_token) hoạt động.
	// Dùng CREATE UNIQUE INDEX IF NOT EXISTS vì Postgres không có ADD CONSTRAINT IF NOT EXISTS.
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_employee_devices_fcm_token ON employee_devices (fcm_token)").Error; err != nil {
		log.Printf("Cảnh báo: không thể tạo unique index fcm_token: %v", err)
	}

	// Onboarding tự phục vụ: bổ sung cột status cho employees/shops và credential
	// portal cho shops. Các cột này idempotent với DB đang chạy. Default 'approved'
	// để dữ liệu cũ (do Admin tạo trực tiếp) không bị kẹt ở trạng thái chờ duyệt.
	selfRegisterMigrations := []string{
		"ALTER TABLE employees ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'approved'",
		"ALTER TABLE shops ADD COLUMN IF NOT EXISTS email VARCHAR(255)",
		"ALTER TABLE shops ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255)",
		"ALTER TABLE shops ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'approved'",
		"ALTER TABLE shops ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_shops_email ON shops (email)",
		// Shop tự đăng ký chưa có webhook_url (set sau trong portal) nên bỏ ràng buộc NOT NULL.
		"ALTER TABLE shops ALTER COLUMN webhook_url DROP NOT NULL",
	}
	for _, m := range selfRegisterMigrations {
		if err := db.Exec(m).Error; err != nil {
			log.Printf("Cảnh báo: không thể chạy migration self-register: %v", err)
		}
	}

	// Phase A (vòng tiền): cột COD đã thu trên đơn + bảng ví/đối soát. Idempotent
	// với DB đang chạy. Xem module 7 trong init.sql để biết mô hình bút toán.
	walletMigrations := []string{
		"ALTER TABLE shipping_orders ADD COLUMN IF NOT EXISTS cod_collected BOOLEAN NOT NULL DEFAULT FALSE",
		"ALTER TABLE shipping_orders ADD COLUMN IF NOT EXISTS cod_collected_at TIMESTAMP",
		`CREATE TABLE IF NOT EXISTS settlements (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			shop_id UUID NOT NULL REFERENCES shops(id),
			total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			note TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			paid_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS wallet_transactions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			shop_id UUID NOT NULL REFERENCES shops(id),
			order_id UUID REFERENCES shipping_orders(id) ON DELETE SET NULL,
			type VARCHAR(20) NOT NULL,
			amount DECIMAL(12,2) NOT NULL,
			settlement_id UUID REFERENCES settlements(id) ON DELETE SET NULL,
			note TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		"CREATE INDEX IF NOT EXISTS idx_wallet_tx_shop ON wallet_transactions (shop_id)",
		"CREATE INDEX IF NOT EXISTS idx_wallet_tx_unsettled ON wallet_transactions (shop_id) WHERE settlement_id IS NULL",
	}
	for _, m := range walletMigrations {
		if err := db.Exec(m).Error; err != nil {
			log.Printf("Cảnh báo: không thể chạy migration wallet: %v", err)
		}
	}

	// Geocoding & Routing migrations
	routingMigrations := []string{
		"ALTER TABLE shipping_orders ADD COLUMN IF NOT EXISTS sender_latitude DECIMAL(10,8)",
		"ALTER TABLE shipping_orders ADD COLUMN IF NOT EXISTS sender_longitude DECIMAL(11,8)",
		"ALTER TABLE shipping_orders ADD COLUMN IF NOT EXISTS receiver_latitude DECIMAL(10,8)",
		"ALTER TABLE shipping_orders ADD COLUMN IF NOT EXISTS receiver_longitude DECIMAL(11,8)",
		"ALTER TABLE shipping_orders ADD COLUMN IF NOT EXISTS pickup_hub_id UUID REFERENCES hubs(id)",
		"ALTER TABLE shipping_orders ADD COLUMN IF NOT EXISTS delivery_hub_id UUID REFERENCES hubs(id)",
	}
	for _, m := range routingMigrations {
		if err := db.Exec(m).Error; err != nil {
			log.Printf("Cảnh báo: không thể chạy migration routing: %v", err)
		}
	}

	// 2. JWT Secret Key setup
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		utils.JWTSecretKey = []byte(secret)
	} else {
		log.Println("CẢNH BÁO: JWT_SECRET chưa được set — đang dùng secret mặc định (KHÔNG an toàn cho production)")
	}

	// Redis client cho session store (thu hồi token thật khi logout / khóa tài khoản)
	redisAddr := fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379"))
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", redisAddr, err)
	}
	log.Println("Redis connected successfully!")

	// 3. Dependency Injection (Wiring)
	// Repositories
	empRepo := repository.NewEmployeeRepository(db)
	shopRepo := repository.NewShopRepository(db)
	rateRepo := repository.NewRateRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	locRepo := repository.NewLocationRepository(db)
	hubRepo := repository.NewHubRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	sessionRepo := repository.NewSessionRepository(rdb)
	deviceRepo := repository.NewDeviceRepository(db)
	walletRepo := repository.NewWalletRepository(db)

	// Bật kiểm tra session trong AuthRequired middleware
	middleware.SessionStore = sessionRepo

	// Infrastructure Services
	webhookSvc := webhook.NewWebhookService()
	notifSvc := notification.NewStubService(deviceRepo)
	
	var emailSvc domain.EmailService
	smtpHost := getEnv("SMTP_HOST", "")
	if smtpHost != "" {
		emailSvc = notification.NewSmtpEmailService(
			smtpHost,
			getEnv("SMTP_PORT", "587"),
			getEnv("SMTP_USER", ""),
			getEnv("SMTP_PASS", ""),
		)
		log.Println("Initialized SMTP Email Service")
	} else {
		emailSvc = notification.NewStubEmailService()
		log.Println("Initialized Stub Email Service")
	}

	_ = notifSvc // sẽ gắn vào luồng đổi trạng thái đơn khi tích hợp push (hiện chưa dùng để tránh đụng order_usecase)

	// UseCases
	geocoder := geocoding.NewNominatimGeocoder()
	authUC := usecase.NewAuthUseCase(empRepo, sessionRepo, emailSvc)
	shopAuthUC := usecase.NewShopAuthUseCase(shopRepo, sessionRepo, emailSvc)
	rateUC := usecase.NewRateUseCase(rateRepo)
	walletUC := usecase.NewWalletUseCase(walletRepo)
	orderUC := usecase.NewOrderUseCase(orderRepo, rateUC, shopRepo, hubRepo, locRepo, geocoder, webhookSvc, walletUC)
	locUC := usecase.NewLocationUseCase(locRepo)
	hubUC := usecase.NewHubUseCase(hubRepo, locRepo)
	shopUC := usecase.NewShopUseCase(shopRepo, sessionRepo, emailSvc)
	empUC := usecase.NewEmployeeUseCase(empRepo)
	statsUC := usecase.NewStatsUseCase(statsRepo)
	deviceUC := usecase.NewDeviceUseCase(deviceRepo)

	// 4. Gin Router Setup
	r := gin.Default()

	// Với API thuần: tắt tự động redirect khi URL thừa/thiếu dấu "/" cuối.
	// Mặc định Gin trả 301/307 sang path chuẩn, nhưng response redirect KHÔNG
	// mang header CORS -> browser chặn preflight (lỗi "No Access-Control-Allow-Origin").
	// Tắt hẳn để mọi request khớp path chính xác hoặc trả 404 rõ ràng.
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// Cấu hình CORS middleware — chỉ cho phép các origin trong ALLOWED_ORIGINS
	// (danh sách cách nhau bởi dấu phẩy). Mặc định là các cổng dev của Vite.
	allowedOrigins := map[string]bool{}
	for _, o := range strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173,http://localhost:5175"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowedOrigins[o] = true
		}
	}
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// 5. Khởi tạo Handlers
	httpDelivery.NewAuthHandler(r, authUC, shopAuthUC)
	httpDelivery.NewOrderHandler(r, orderUC, rateUC, shopRepo)
	httpDelivery.NewLocationHandler(r, locUC)
	httpDelivery.NewHubHandler(r, hubUC)
	httpDelivery.NewRateHandler(r, rateUC)
	httpDelivery.NewShopHandler(r, shopUC)
	httpDelivery.NewEmployeeHandler(r, empUC)
	httpDelivery.NewStatsHandler(r, statsUC)
	httpDelivery.NewRoutesHandler(r)
	httpDelivery.NewDeviceHandler(r, deviceUC)
	httpDelivery.NewWalletHandler(r, walletUC)

	// Test Route yêu cầu đăng nhập
	api := r.Group("/api/v1")
	api.Use(middleware.AuthRequired())
	{
		api.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			role, _ := c.Get("role")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"user_id": userID,
					"role":    role,
				},
			})
		})
	}

	port := getEnv("PORT", "8080")
	log.Printf("Server is running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
