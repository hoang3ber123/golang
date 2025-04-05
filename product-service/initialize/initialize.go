package initialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"product-service/config"
	"product-service/internal/db"
	"product-service/internal/models"
	"time"
)

type KongService struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type KongRoute struct {
	Paths     []string `json:"paths"`
	Name      string   `json:"name"`
	StripPath bool     `json:"strip_path,omitempty"`
}

type PluginConfig struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config,omitempty"`
}

type Consumer struct {
	Username string `json:"username"`
}

type KeyAuth struct {
	Key string `json:"key"`
}

func ConnectToApiGateway() {
	kongAdminURL := config.Config.APIGatewayHost
	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Printf("Starting Kong configuration with Admin URL: %s\n", kongAdminURL)

	// Chờ Kong sẵn sàng
	for {
		fmt.Println("Checking Kong status...")
		resp, err := client.Get(fmt.Sprintf("%s/status", kongAdminURL))
		if err != nil {
			fmt.Printf("🔄 Waiting for Kong to be ready... Error: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()

		fmt.Printf("Received status response with code: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status response body: %s\n", string(body))

		var status map[string]interface{}
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&status); err != nil {
			fmt.Printf("Error decoding status response: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if database, ok := status["database"].(map[string]interface{}); ok {
			if reachable, ok := database["reachable"].(bool); ok && reachable {
				fmt.Println("Kong database is reachable!")
				break
			}
			fmt.Println("Database not reachable yet...")
		} else {
			fmt.Println("No database status found in response")
		}
		fmt.Println("🔄 Waiting for Kong to be ready...")
		time.Sleep(5 * time.Second)
	}

	fmt.Println("✅ Kong is ready! Applying configurations...")

	// 1. Tạo Service cho auth-service
	service := KongService{
		Name: config.Config.ServiceName,
		URL:  config.Config.ServiceURL,
	}
	fmt.Printf("Creating service: %s with URL: %s\n", service.Name, service.URL)
	if err := postJSON(client, fmt.Sprintf("%s/services", kongAdminURL), service); err != nil {
		fmt.Printf("Error creating service: %v\n", err)
	} else {
		fmt.Println("Service created successfully")
	}

	// 2. Tạo Route cho Service
	route := KongRoute{
		Paths:     []string{config.Config.ServicePath},
		Name:      config.Config.ServiceRoute,
		StripPath: true,
	}
	fmt.Printf("Creating route: %s with path: %s\n", route.Name, route.Paths)
	if err := postJSON(client, fmt.Sprintf("%s/services/%s/routes", kongAdminURL, service.Name), route); err != nil {
		fmt.Printf("Error creating route: %v\n", err)
	} else {
		fmt.Println("Route created successfully")
	}

	fmt.Println("✅ Kong setup completed!")
}

// Hàm helper để gửi POST request với JSON
func postJSON(client *http.Client, url string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return err
	}

	fmt.Printf("Sending POST request to %s with data: %s\n", url, string(jsonData))
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending POST request: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Received response with status code: %d\n", resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response body: %s\n", string(body))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func InitializingDatabase() {
	// Khởi tạo category
	InitializeCategory()
	// Khởi tạo product
}

// hàm khởi tạo category
func InitializeCategory() {
	categories := []models.Category{
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "backend",
			},
			Description: "Server-side, APIs, databases, RESTful, GraphQL, authentication, authorization, MVC, ORM, backend development, middleware, caching, scalability, backend architecture, load balancing, API gateway",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "nodejs",
			},
			Description: "JavaScript runtime, scalable backend, Express.js, NestJS, event-driven, asynchronous programming, non-blocking I/O, V8 engine, serverless, WebSockets, microservices, API development, package management (npm, yarn)",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "microservices",
			},
			Description: "Distributed services, architecture, containerization, Kubernetes, Docker, service discovery, API gateway, inter-service communication, CQRS, event sourcing, decentralized architecture, domain-driven design (DDD), fault tolerance",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "performance optimization",
			},
			Description: "Speed, latency, efficiency, caching strategies, database indexing, load balancing, CDN, query optimization, profiling, memory management, parallel processing, multithreading, performance monitoring, response time, request throttling",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "budget-friendly",
			},
			Description: "Low cost, affordable solutions, cost optimization, open-source tools, free-tier services, cloud cost management, pricing models, SaaS alternatives, resource efficiency, cost-effective hosting, serverless pricing, FinOps",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "frontend",
			},
			Description: "UI, client-side, responsive design, HTML, CSS, JavaScript, TypeScript, CSS frameworks (Bootstrap, Tailwind, Material UI), UI components, animations, DOM manipulation, progressive web apps (PWA), single-page applications (SPA), WebAssembly, frontend optimization",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "react",
			},
			Description: "Interactive UI, JavaScript library, React.js, React Native, JSX, virtual DOM, component-based architecture, hooks, state management (Redux, Recoil, Zustand), SSR (Next.js), hydration, reconciliation, frontend framework",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "ai",
			},
			Description: "Artificial intelligence, machine learning, deep learning, neural networks, NLP, computer vision, reinforcement learning, generative AI, transformer models, large language models (LLM), AI ethics, data science, predictive analytics, AI-driven automation",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "ecommerce",
			},
			Description: "Online stores, shopping, transactions, payment gateways (Stripe, PayPal, VNPay), shopping cart, checkout flow, order management, product catalog, customer reviews, dropshipping, marketplace, subscription model, SEO for ecommerce, user experience (UX), conversion rate optimization (CRO)",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "saas",
			},
			Description: "Software as a Service, cloud solutions, multi-tenant architecture, subscription-based, SaaS pricing models, API-first development, microservices for SaaS, customer onboarding, usage analytics, scalability, CI/CD, DevOps, cloud hosting (AWS, Azure, GCP), security compliance (SOC 2, GDPR)",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "portfolio",
			},
			Description: "Showcase projects, personal branding, web portfolio, design portfolio, developer portfolio, case studies, UI/UX presentation, interactive resume, testimonials, online presence, custom domain, SEO optimization, responsive design",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "blog",
			},
			Description: "Content writing, publishing, news, CMS (WordPress, Ghost, Strapi), Markdown, SEO, social media integration, email newsletters, blog monetization, affiliate marketing, audience engagement, blog analytics, content strategy, editorial workflow",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "landing-page",
			},
			Description: "Marketing, conversions, lead generation, sales funnel, call-to-action (CTA), A/B testing, copywriting, UI/UX design, high-converting pages, one-page websites, performance tracking, Google Ads, Facebook Pixel, SEO optimization",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "news",
			},
			Description: "Media, articles, latest updates, journalism, online magazines, breaking news, press releases, RSS feeds, news aggregation, real-time updates, media coverage, social media trends, digital publishing, fact-checking",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "real-estate",
			},
			Description: "Property listings, real estate solutions, rental properties, commercial real estate, mortgage calculators, house valuation, property management, real estate CRM, MLS (Multiple Listing Service), home-buying process, real estate investments, virtual tours",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "web3",
			},
			Description: "Decentralized applications, blockchain, smart contracts, Ethereum, NFTs, DeFi (Decentralized Finance), DAOs (Decentralized Autonomous Organizations), tokenomics, crypto wallets, metaverse, on-chain governance, Web3 authentication, Layer 2 scaling solutions",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "startup",
			},
			Description: "Entrepreneurship, business growth, startup funding, venture capital, bootstrapping, business model canvas, go-to-market strategy, pitch decks, MVP (Minimum Viable Product), customer acquisition, product-market fit, accelerator programs, startup scaling",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "tech",
			},
			Description: "Technology, innovations, IT, artificial intelligence, cloud computing, cybersecurity, data science, IoT (Internet of Things), big data, quantum computing, 5G networks, emerging technologies, IT infrastructure, digital transformation",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "modern",
			},
			Description: "Contemporary design, latest trends, minimalism, futuristic UI, neomorphic design, glassmorphism, dark mode, responsive layouts, creative direction, user-centric design, modern typography, digital aesthetics, web trends",
		},
		{
			BaseSlugUnique: models.BaseSlugUnique{
				Title: "animated",
			},
			Description: "Motion graphics, interactive UI, CSS animations, Lottie animations, SVG animations, microinteractions, transitions, parallax effects, 3D animations, WebGL, After Effects, real-time rendering, immersive user experience",
		},
	}
	if err := db.DB.Model(models.Category{}).Create(&categories).Error; err != nil {
		fmt.Println("Error when create category")
	} else {
		fmt.Println("Create successfully")
	}
}
