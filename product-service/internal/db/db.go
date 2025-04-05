package db

import (
	"fmt"
	"math/rand"
	"product-service/config"
	"product-service/internal/models"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Config.DatabaseUsername, config.Config.DatabasePassword,
		config.Config.DatabaseHost, config.Config.DatabasePort, config.Config.DatabaseName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,                                // turn off transaction for all of query, increase 30% performance
		Logger:                 logger.Default.LogMode(logger.Info), // Log all SQL queries
	})
	if err != nil {
		panic("Failed to connect to database")
	}
	DB = db

	// Auto-migrate models
	db.AutoMigrate(&models.Category{}, &models.Product{}, &models.ProductCategory{}, &models.Media{}, &models.Cart{}, &models.CartItem{}, &models.HistoryView{}, &models.HistorySearch{})
	// Init database

	// Insert default categories using raw SQL
	initCategory()
	// Insert default product using raw SQL
	initProduct()
	// Insert default product_categories using raw SQL
	initProductCategories()
}
func initCategory() {
	var count int64
	DB.Model(models.Category{}).Count(&count)
	if count > 0 {
		fmt.Println("Data is exit.")
		return
	}
	sql := "INSERT INTO `categories` (`id`,`created_at`,`updated_at`,`title`,`slug`,`description`,`parent_id`) VALUES ('66adbf73-a8d2-46dd-ad03-5e7e69ed3367','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','backend','backend','Server-side, APIs, databases, RESTful, GraphQL, authentication, authorization, MVC, ORM, backend development, middleware, caching, scalability, backend architecture, load balancing, API gateway',NULL),('546cff2a-f067-441a-a4c7-fd5b2ff1817b','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','nodejs','nodejs','JavaScript runtime, scalable backend, Express.js, NestJS, event-driven, asynchronous programming, non-blocking I/O, V8 engine, serverless, WebSockets, microservices, API development, package management (npm, yarn)',NULL),('df0f4c62-2c7b-4055-a31e-32837fe84432','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','microservices','microservices','Distributed services, architecture, containerization, Kubernetes, Docker, service discovery, API gateway, inter-service communication, CQRS, event sourcing, decentralized architecture, domain-driven design (DDD), fault tolerance',NULL),('5f640a87-1eb0-4da8-aa96-bbf7ae30c640','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','performance optimization','performance-optimization','Speed, latency, efficiency, caching strategies, database indexing, load balancing, CDN, query optimization, profiling, memory management, parallel processing, multithreading, performance monitoring, response time, request throttling',NULL),('8050542b-c101-4e2d-861d-b5eef547b703','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','budget-friendly','budget-friendly','Low cost, affordable solutions, cost optimization, open-source tools, free-tier services, cloud cost management, pricing models, SaaS alternatives, resource efficiency, cost-effective hosting, serverless pricing, FinOps',NULL),('ee7cbb2f-2694-46b0-a024-ecae6927522e','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','frontend','frontend','UI, client-side, responsive design, HTML, CSS, JavaScript, TypeScript, CSS frameworks (Bootstrap, Tailwind, Material UI), UI components, animations, DOM manipulation, progressive web apps (PWA), single-page applications (SPA), WebAssembly, frontend optimization',NULL),('87dec9bf-f012-4eea-a059-24493665c865','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','react','react','Interactive UI, JavaScript library, React.js, React Native, JSX, virtual DOM, component-based architecture, hooks, state management (Redux, Recoil, Zustand), SSR (Next.js), hydration, reconciliation, frontend framework',NULL),('c1e710fa-e04b-4405-b0a2-13aa07bfd268','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','ai','ai','Artificial intelligence, machine learning, deep learning, neural networks, NLP, computer vision, reinforcement learning, generative AI, transformer models, large language models (LLM), AI ethics, data science, predictive analytics, AI-driven automation',NULL),('4778a70c-ff6f-4e13-ab10-7ea62610ad54','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','ecommerce','ecommerce','Online stores, shopping, transactions, payment gateways (Stripe, PayPal, VNPay), shopping cart, checkout flow, order management, product catalog, customer reviews, dropshipping, marketplace, subscription model, SEO for ecommerce, user experience (UX), conversion rate optimization (CRO)',NULL),('6f56895b-bc44-46af-8e06-bb5a27b0286a','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','saas','saas','Software as a Service, cloud solutions, multi-tenant architecture, subscription-based, SaaS pricing models, API-first development, microservices for SaaS, customer onboarding, usage analytics, scalability, CI/CD, DevOps, cloud hosting (AWS, Azure, GCP), security compliance (SOC 2, GDPR)',NULL),('dfc27355-33c7-4f57-a1be-0715d63133a0','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','portfolio','portfolio','Showcase projects, personal branding, web portfolio, design portfolio, developer portfolio, case studies, UI/UX presentation, interactive resume, testimonials, online presence, custom domain, SEO optimization, responsive design',NULL),('c5199b24-64b4-4deb-8bb6-6eec47298e3f','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','blog','blog','Content writing, publishing, news, CMS (WordPress, Ghost, Strapi), Markdown, SEO, social media integration, email newsletters, blog monetization, affiliate marketing, audience engagement, blog analytics, content strategy, editorial workflow',NULL),('bf492675-a8e7-4ed7-8c39-34b8f5318e25','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','landing-page','landing-page','Marketing, conversions, lead generation, sales funnel, call-to-action (CTA), A/B testing, copywriting, UI/UX design, high-converting pages, one-page websites, performance tracking, Google Ads, Facebook Pixel, SEO optimization',NULL),('59e69862-6a7b-4d30-bfa9-d70cd28b418e','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','news','news','Media, articles, latest updates, journalism, online magazines, breaking news, press releases, RSS feeds, news aggregation, real-time updates, media coverage, social media trends, digital publishing, fact-checking',NULL),('5d4566dc-dbad-4923-b792-246f292a17fe','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','real-estate','real-estate','Property listings, real estate solutions, rental properties, commercial real estate, mortgage calculators, house valuation, property management, real estate CRM, MLS (Multiple Listing Service), home-buying process, real estate investments, virtual tours',NULL),('8196c191-9b67-4960-ac2a-6fa2a6fb8fc3','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','web3','web3','Decentralized applications, blockchain, smart contracts, Ethereum, NFTs, DeFi (Decentralized Finance), DAOs (Decentralized Autonomous Organizations), tokenomics, crypto wallets, metaverse, on-chain governance, Web3 authentication, Layer 2 scaling solutions',NULL),('816ba3eb-624b-48c3-84a5-8d8db7bbfa5a','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','startup','startup','Entrepreneurship, business growth, startup funding, venture capital, bootstrapping, business model canvas, go-to-market strategy, pitch decks, MVP (Minimum Viable Product), customer acquisition, product-market fit, accelerator programs, startup scaling',NULL),('106f419f-4fa8-4d11-947d-bb0f1aa58dbd','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','tech','tech','Technology, innovations, IT, artificial intelligence, cloud computing, cybersecurity, data science, IoT (Internet of Things), big data, quantum computing, 5G networks, emerging technologies, IT infrastructure, digital transformation',NULL),('62b6f354-541e-4322-9453-8361977f26e3','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','modern','modern','Contemporary design, latest trends, minimalism, futuristic UI, neomorphic design, glassmorphism, dark mode, responsive layouts, creative direction, user-centric design, modern typography, digital aesthetics, web trends',NULL),('86bf9fe5-0dfa-418c-953c-21c6f5ca124b','2025-04-05 12:07:29.965','2025-04-05 12:07:29.965','animated','animated','Motion graphics, interactive UI, CSS animations, Lottie animations, SVG animations, microinteractions, transitions, parallax effects, 3D animations, WebGL, After Effects, real-time rendering, immersive user experience',NULL)"
	DB.Exec(sql)
}
func initProduct() {
	var count int64
	DB.Model(models.Product{}).Count(&count)
	if count > 0 {
		fmt.Println("Data is exit.")
		return
	}
	sql := `INSERT INTO products (id, created_at, updated_at, title, slug, description, link, user_id, price)
VALUES
('550e8400-e29b-41d4-a716-446655440000', NOW(), NOW(), 'product1', 'product1', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 10.00),
('550e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'product2', 'product2', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 15.50),
('550e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'product3', 'product3', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 20.00),
('550e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), 'product4', 'product4', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 25.75),
('550e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), 'product5', 'product5', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 30.00),
('550e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), 'product6', 'product6', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 35.25),
('550e8400-e29b-41d4-a716-446655440006', NOW(), NOW(), 'product7', 'product7', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 40.00),
('550e8400-e29b-41d4-a716-446655440007', NOW(), NOW(), 'product8', 'product8', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 45.50),
('550e8400-e29b-41d4-a716-446655440008', NOW(), NOW(), 'product9', 'product9', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 50.00),
('550e8400-e29b-41d4-a716-446655440009', NOW(), NOW(), 'product10', 'product10', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 55.75),
('550e8400-e29b-41d4-a716-446655440010', NOW(), NOW(), 'product11', 'product11', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 60.00),
('550e8400-e29b-41d4-a716-446655440011', NOW(), NOW(), 'product12', 'product12', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 65.25),
('550e8400-e29b-41d4-a716-446655440012', NOW(), NOW(), 'product13', 'product13', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 70.00),
('550e8400-e29b-41d4-a716-446655440013', NOW(), NOW(), 'product14', 'product14', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 75.50),
('550e8400-e29b-41d4-a716-446655440014', NOW(), NOW(), 'product15', 'product15', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 80.00),
('550e8400-e29b-41d4-a716-446655440015', NOW(), NOW(), 'product16', 'product16', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 85.75),
('550e8400-e29b-41d4-a716-446655440016', NOW(), NOW(), 'product17', 'product17', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 90.00),
('550e8400-e29b-41d4-a716-446655440017', NOW(), NOW(), 'product18', 'product18', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 95.25),
('550e8400-e29b-41d4-a716-446655440018', NOW(), NOW(), 'product19', 'product19', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 100.00),
('550e8400-e29b-41d4-a716-446655440019', NOW(), NOW(), 'product20', 'product20', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 10.50),
('550e8400-e29b-41d4-a716-446655440020', NOW(), NOW(), 'product21', 'product21', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 15.00),
('550e8400-e29b-41d4-a716-446655440021', NOW(), NOW(), 'product22', 'product22', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 20.25),
('550e8400-e29b-41d4-a716-446655440022', NOW(), NOW(), 'product23', 'product23', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 25.00),
('550e8400-e29b-41d4-a716-446655440023', NOW(), NOW(), 'product24', 'product24', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 30.75),
('550e8400-e29b-41d4-a716-446655440024', NOW(), NOW(), 'product25', 'product25', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 35.00),
('550e8400-e29b-41d4-a716-446655440025', NOW(), NOW(), 'product26', 'product26', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 40.25),
('550e8400-e29b-41d4-a716-446655440026', NOW(), NOW(), 'product27', 'product27', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 45.00),
('550e8400-e29b-41d4-a716-446655440027', NOW(), NOW(), 'product28', 'product28', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 50.50),
('550e8400-e29b-41d4-a716-446655440028', NOW(), NOW(), 'product29', 'product29', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 55.00),
('550e8400-e29b-41d4-a716-446655440029', NOW(), NOW(), 'product30', 'product30', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 60.75),
('550e8400-e29b-41d4-a716-446655440030', NOW(), NOW(), 'product31', 'product31', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 65.00),
('550e8400-e29b-41d4-a716-446655440031', NOW(), NOW(), 'product32', 'product32', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 70.25),
('550e8400-e29b-41d4-a716-446655440032', NOW(), NOW(), 'product33', 'product33', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 75.00),
('550e8400-e29b-41d4-a716-446655440033', NOW(), NOW(), 'product34', 'product34', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 80.50),
('550e8400-e29b-41d4-a716-446655440034', NOW(), NOW(), 'product35', 'product35', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 85.00),
('550e8400-e29b-41d4-a716-446655440035', NOW(), NOW(), 'product36', 'product36', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 90.75),
('550e8400-e29b-41d4-a716-446655440036', NOW(), NOW(), 'product37', 'product37', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 95.00),
('550e8400-e29b-41d4-a716-446655440037', NOW(), NOW(), 'product38', 'product38', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 100.25),
('550e8400-e29b-41d4-a716-446655440038', NOW(), NOW(), 'product39', 'product39', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 10.75),
('550e8400-e29b-41d4-a716-446655440039', NOW(), NOW(), 'product40', 'product40', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 15.00);
`
	DB.Exec(sql)
}
func randomSample(arr []string, count int) []string {
	if count >= len(arr) {
		return append([]string(nil), arr...)
	}
	perm := rand.Perm(len(arr))
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = arr[perm[i]]
	}
	return result
}
func initProductCategories() {
	// Kiểm tra nếu đã có ít nhất 1 dòng
	var count int64
	DB.Model(models.ProductCategory{}).Count(&count)
	if count > 0 {
		fmt.Println("Data is exit.")
		return
	}
	categories := []string{
		"66adbf73-a8d2-46dd-ad03-5e7e69ed3367",
		"546cff2a-f067-441a-a4c7-fd5b2ff1817b",
		"df0f4c62-2c7b-4055-a31e-32837fe84432",
		"5f640a87-1eb0-4da8-aa96-bbf7ae30c640",
		"8050542b-c101-4e2d-861d-b5eef547b703",
		"ee7cbb2f-2694-46b0-a024-ecae6927522e",
		"87dec9bf-f012-4eea-a059-24493665c865",
		"c1e710fa-e04b-4405-b0a2-13aa07bfd268",
		"4778a70c-ff6f-4e13-ab10-7ea62610ad54",
		"6f56895b-bc44-46af-8e06-bb5a27b0286a",
		"dfc27355-33c7-4f57-a1be-0715d63133a0",
		"c5199b24-64b4-4deb-8bb6-6eec47298e3f",
		"bf492675-a8e7-4ed7-8c39-34b8f5318e25",
		"59e69862-6a7b-4d30-bfa9-d70cd28b418e",
		"5d4566dc-dbad-4923-b792-246f292a17fe",
		"8196c191-9b67-4960-ac2a-6fa2a6fb8fc3",
		"816ba3eb-624b-48c3-84a5-8d8db7bbfa5a",
		"106f419f-4fa8-4d11-947d-bb0f1aa58dbd",
		"62b6f354-541e-4322-9453-8361977f26e3",
		"86bf9fe5-0dfa-418c-953c-21c6f5ca124b",
	}

	products := make([]string, 40)
	for i := 0; i < 40; i++ {
		products[i] = fmt.Sprintf("550e8400-e29b-41d4-a716-4466554400%02d", i)
	}

	rand.Seed(time.Now().UnixNano())
	productCategorySet := make(map[string]struct{})
	var values []string

	// Mỗi category có ít nhất 1 sản phẩm
	for i, category := range categories {
		productID := products[i%len(products)]
		key := productID + ":" + category
		if _, exists := productCategorySet[key]; !exists {
			productCategorySet[key] = struct{}{}
			values = append(values, fmt.Sprintf("('%s', '%s')", productID, category))
		}
	}

	// Gán ngẫu nhiên category cho sản phẩm
	for _, product := range products {
		numCategories := rand.Intn(10) + 1 // 1 đến 10
		selected := randomSample(categories, numCategories)
		for _, category := range selected {
			key := product + ":" + category
			if _, exists := productCategorySet[key]; !exists {
				productCategorySet[key] = struct{}{}
				values = append(values, fmt.Sprintf("('%s', '%s')", product, category))
			}
		}
	}

	sql := "INSERT INTO product_categories (product_id, category_id) VALUES\n"
	sql += strings.Join(values, ",\n") + ";"

	DB.Exec(sql)
}
