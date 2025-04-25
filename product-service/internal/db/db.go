package db

import (
	"fmt"
	"math/rand"
	"product-service/config"
	"product-service/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
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
	db.AutoMigrate(&models.Category{}, &models.Product{}, &models.ProductCategory{}, &models.Media{}, &models.Cart{}, &models.CartItem{}, &models.HistoryView{}, &models.HistorySearch{}, models.Task{})
	// Init database

	// Insert default categories using raw SQL
	initCategory()
	// Insert default product using raw SQL
	initProduct()
	// Insert default product_categories using raw SQL
	initProductCategories()
	// insert Media for product
	initMediaForProducts()
}
func initCategory() {
	var count int64
	DB.Model(models.Category{}).Count(&count)
	if count > 0 {
		fmt.Println("Data already exists.")
		return
	}

	// Step 1: Insert top-level categories
	topLevelCategories := `INSERT INTO categories (
  id,
  created_at,
  updated_at,
  title,
  slug,
  description,
  parent_id
) VALUES
-- === Top‐level categories ===
('66adbf73-a8d2-46dd-ad03-5e7e69ed3367', NOW(), NOW(), 'backend',           'backend',           'Server-side, APIs, databases, RESTful, GraphQL, authentication, authorization, MVC, ORM, backend development, middleware, caching, scalability, backend architecture, load balancing, API gateway',                                                                                       NULL),
('ee7cbb2f-2694-46b0-a024-ecae6927522e', NOW(), NOW(), 'frontend',          'frontend',          'UI, client-side, responsive design, HTML, CSS, JavaScript, TypeScript, CSS frameworks (Bootstrap, Tailwind, Material UI), UI components, animations, DOM manipulation, progressive web apps (PWA), single-page applications (SPA), WebAssembly, frontend optimization',                                         NULL),
('106f419f-4fa8-4d11-947d-bb0f1aa58dbd', NOW(), NOW(), 'tech',              'tech',              'Technology, innovations, IT, artificial intelligence, cloud computing, cybersecurity, data science, IoT (Internet of Things), big data, quantum computing, 5G networks, emerging technologies, IT infrastructure, digital transformation',                                                   NULL),
('8050542b-c101-4e2d-861d-b5eef547b703', NOW(), NOW(), 'budget-friendly',   'budget-friendly',   'Low cost, affordable solutions, cost optimization, open-source tools, free-tier services, cloud cost management, pricing models, SaaS alternatives, resource efficiency, cost-effective hosting, serverless pricing, FinOps',                                                                                               NULL),
('4778a70c-ff6f-4e13-ab10-7ea62610ad54', NOW(), NOW(), 'ecommerce',         'ecommerce',         'Online stores, shopping, transactions, payment gateways (Stripe, PayPal, VNPay), shopping cart, checkout flow, order management, product catalog, customer reviews, dropshipping, marketplace, subscription model, SEO for ecommerce, user experience (UX), conversion rate optimization (CRO)',                     NULL),
('6f56895b-bc44-46af-8e06-bb5a27b0286a', NOW(), NOW(), 'saas',              'saas',              'Software as a Service, cloud solutions, multi-tenant architecture, subscription-based, SaaS pricing models, API-first development, microservices for SaaS, customer onboarding, usage analytics, scalability, CI/CD, DevOps, cloud hosting (AWS, Azure, GCP), security compliance (SOC 2, GDPR)',    NULL),
('dfc27355-33c7-4f57-a1be-0715d63133a0', NOW(), NOW(), 'portfolio',         'portfolio',         'Showcase projects, personal branding, web portfolio, design portfolio, developer portfolio, case studies, UI/UX presentation, interactive resume, testimonials, online presence, custom domain, SEO optimization, responsive design',                                                                                                         NULL),
('c5199b24-64b4-4deb-8bb6-6eec47298e3f', NOW(), NOW(), 'blog',              'blog',              'Content writing, publishing, news, CMS (WordPress, Ghost, Strapi), Markdown, SEO, social media integration, email newsletters, blog monetization, affiliate marketing, audience engagement, blog analytics, content strategy, editorial workflow',                                                                                        NULL),
('bf492675-a8e7-4ed7-8c39-34b8f5318e25', NOW(), NOW(), 'landing-page',       'landing-page',       'Marketing, conversions, lead generation, sales funnel, call-to-action (CTA), A/B testing, copywriting, UI/UX design, high-converting pages, one-page websites, performance tracking, Google Ads, Facebook Pixel, SEO optimization',                                                                                   NULL),
('59e69862-6a7b-4d30-bfa9-d70cd28b418e', NOW(), NOW(), 'news',              'news',              'Media, articles, latest updates, journalism, online magazines, breaking news, press releases, RSS feeds, news aggregation, real-time updates, media coverage, social media trends, digital publishing, fact-checking',                                                                                         NULL),
('5d4566dc-dbad-4923-b792-246f292a17fe', NOW(), NOW(), 'real-estate',       'real-estate',       'Property listings, real estate solutions, rental properties, commercial real estate, mortgage calculators, house valuation, property management, real estate CRM, MLS (Multiple Listing Service), home-buying process, real estate investments, virtual tours',                                              NULL),
('8196c191-9b67-4960-ac2a-6fa2a6fb8fc3', NOW(), NOW(), 'web3',              'web3',              'Decentralized applications, blockchain, smart contracts, Ethereum, NFTs, DeFi (Decentralized Finance), DAOs (Decentralized Autonomous Organizations), tokenomics, crypto wallets, metaverse, on-chain governance, Web3 authentication, Layer 2 scaling solutions',                                             NULL),
('816ba3eb-624b-48c3-84a5-8d8db7bbfa5a', NOW(), NOW(), 'startup',           'startup',           'Entrepreneurship, business growth, startup funding, venture capital, bootstrapping, business model canvas, go-to-market strategy, pitch decks, MVP (Minimum Viable Product), customer acquisition, product-market fit, accelerator programs, startup scaling',                                             NULL),
('de3a5d13-e2b1-40d2-5d13-dad62d9f52e0', NOW(), NOW(), 'Agile',             'agile',             'Software development methodology, iterative development, incremental delivery, collaboration, customer feedback, flexibility, adaptability, Scrum, Kanban, Lean, sprints, user stories',                                                         NULL),
('ef4b6e24-f3c2-41e3-6e24-ebe73ea063f0', NOW(), NOW(), 'Software Testing',  'software-testing',  'Quality assurance (QA), bug detection, verification, validation, manual testing, automated testing, unit tests, integration tests, end-to-end (E2E) tests, performance testing, security testing, TDD',                                          NULL),
('f05c7f35-04d3-42f4-7f35-fcf84fb17400', NOW(), NOW(), 'Git',               'git',               'Distributed version control system, source code management, branching, merging, commits, repositories, collaboration, code history, GitHub, GitLab, Bitbucket, conflict resolution',                                                                                    NULL),
('016d8046-15e4-4305-8046-0d0950c28501', NOW(), NOW(), 'Social Media Platform','social-media', 'Online communities, user-generated content, networking, profiles, feeds, posts, likes, comments, shares, messaging, real-time updates, platform development, content moderation',                               NULL),
('127e9157-26f5-4416-9157-1e1a61d39602', NOW(), NOW(), 'E-learning Platform', 'elearning',     'Online education, learning management system (LMS), course delivery, online courses, virtual classrooms, student tracking, assessments, quizzes, interactive content, SCORM, Moodle, Canvas',                                                                               NULL),
('238fa268-3706-4527-a268-2f2b72e4a703', NOW(), NOW(), 'Booking System',    'booking-system',    'Reservations, scheduling, appointments, online booking, availability management, calendar integration, payment processing, confirmations, reminders, resource management (rooms, seats, staff)',                                      NULL),
('3490b379-4817-4638-b379-303c83f5b804', NOW(), NOW(), 'Forum / Community',  'forum',            'Online discussion board, community building, user interaction, threads, posts, replies, moderation, user profiles, categories, Q&A, knowledge sharing, Discourse, phpBB',                                                                              NULL),
('45a1c48a-5928-4749-c48a-414d9406c905', NOW(), NOW(), 'Job Board',         'job-board',         'Employment listings, job postings, candidate applications, resume submission, employer profiles, job search filters, career portal, recruitment platform, ATS integration',                                                               NULL),
('56b2d59b-6a39-4859-d59b-525ea517d0a6', NOW(), NOW(), 'CRM Software',       'crm',               'Customer relationship management, sales automation, marketing automation, customer service, contact management, lead tracking, pipeline management, reporting, analytics, Salesforce, HubSpot',                                          NULL);`
	DB.Exec(topLevelCategories)

	// Step 2: Insert first-level subcategories for frontend
	frontendSubcategories := `INSERT INTO categories (
  id,
  created_at,
  updated_at,
  title,
  slug,
  description,
  parent_id
) VALUES
('86bf9fe5-0dfa-418c-953c-21c6f5ca124b', NOW(), NOW(), 'animated',          'animated',          'Motion graphics, interactive UI, CSS animations, Lottie animations, SVG animations, microinteractions, transitions, parallax effects, 3D animations, WebGL, After Effects, real-time rendering, immersive user experience',                                                                                    'ee7cbb2f-2694-46b0-a024-ecae6927522e'),
('62b6f354-541e-4322-9453-8361977f26e3', NOW(), NOW(), 'modern',            'modern',            'Contemporary design, latest trends, minimalism, futuristic UI, neomorphic design, glassmorphism, dark mode, responsive layouts, creative direction, user-centric design, modern typography, digital aesthetics, web trends',                                                                                  'ee7cbb2f-2694-46b0-a024-ecae6927522e'),
('87dec9bf-f012-4eea-a059-24493665c865', NOW(), NOW(), 'react',             'react',             'Interactive UI, JavaScript library, React.js, React Native, JSX, virtual DOM, component-based architecture, hooks, state management (Redux, Recoil, Zustand), SSR (Next.js), hydration, reconciliation, frontend framework',                                             'ee7cbb2f-2694-46b0-a024-ecae6927522e'),
('78d4f7bb-8c5b-4a7b-f7bb-7470c739fc08', NOW(), NOW(), 'Vue.js',            'vuejs',             'Progressive JavaScript framework, component-based, virtual DOM, reactivity, single-file components (.vue), approachable, performant, Nuxt.js (SSR/SSG), Vuex/Pinia',                                                                                          'ee7cbb2f-2694-46b0-a024-ecae6927522e'),
('89e508cc-9d6c-4b8c-08cc-8581d84a0d09', NOW(), NOW(), 'Angular',           'angular',           'TypeScript-based framework, platform, component-based, dependency injection, RxJS, modules, services, routing, CLI, enterprise applications, opinionated structure',                                                                                         'ee7cbb2f-2694-46b0-a024-ecae6927522e'),
('9af619dd-ae7d-4c9d-19dd-9692e95b1e0a', NOW(), NOW(), 'TypeScript',        'typescript',        'Typed superset of JavaScript, static typing, code scalability, maintainability, developer productivity, interfaces, generics, enums, tooling support, compiles to JavaScript',                                                                                 'ee7cbb2f-2694-46b0-a024-ecae6927522e'),
('ab072ae0-bf8e-4daf-2ae0-a7a3fa6c2f0b', NOW(), NOW(), 'Tailwind CSS',      'tailwindcss',       'Utility-first CSS framework, rapid UI development, customizable, responsive design, JIT mode, low-level utilities, design system implementation, no predefined components',                                                                             'ee7cbb2f-2694-46b0-a024-ecae6927522e'),
('bc183bf1-c09f-4eb0-3bf1-b8b40b7d30c0', NOW(), NOW(), 'Accessibility',     'accessibility',     'a11y, inclusive design, WCAG, screen readers, keyboard navigation, ARIA attributes, semantic HTML, color contrast, usability for everyone, assistive technologies',                                                                                            'ee7cbb2f-2694-46b0-a024-ecae6927522e'),
('cd294c02-d1a0-4fc1-4c02-c9c51c8e41d0', NOW(), NOW(), 'Progressive Web Apps','pwa',               'Web applications, native app-like experience, offline support, push notifications, installable, service workers, manifest file, HTTPS, reliable, fast, engaging',                                                                                         'ee7cbb2f-2694-46b0-a024-ecae6927522e');`
	DB.Exec(frontendSubcategories)

	// Step 3: Insert first-level subcategories for backend
	backendSubcategories := `INSERT INTO categories (
  id,
  created_at,
  updated_at,
  title,
  slug,
  description,
  parent_id
) VALUES
('546cff2a-f067-441a-a4c7-fd5b2ff1817b', NOW(), NOW(), 'nodejs',            'nodejs',            'JavaScript runtime, scalable backend, Express.js, NestJS, event-driven, asynchronous programming, non-blocking I/O, V8 engine, serverless, WebSockets, microservices, API development, package management (npm, yarn)',                                           '66adbf73-a8d2-46dd-ad03-5e7e69ed3367'),
('df0f4c62-2c7b-4055-a31e-32837fe84432', NOW(), NOW(), 'microservices',     'microservices',     'Distributed services, architecture, containerization, Kubernetes, Docker, service discovery, API gateway, inter-service communication, CQRS, event sourcing, decentralized architecture, domain-driven design (DDD), fault tolerance',                                                            '66adbf73-a8d2-46dd-ad03-5e7e69ed3367'),
('5f640a87-1eb0-4da8-aa96-bbf7ae30c640', NOW(), NOW(), 'performance optimization','performance-optimization','Speed, latency, efficiency, caching strategies, database indexing, load balancing, CDN, query optimization, profiling, memory management, parallel processing, multithreading, performance monitoring, response time, request throttling',                               '66adbf73-a8d2-46dd-ad03-5e7e69ed3367'),
('6fd3a8ba-dc4e-4f6a-a8ba-b3c4d5e6f7a9', NOW(), NOW(), 'Python',            'python',            'General-purpose language, backend development, data science, machine learning, scripting, web frameworks (Django, Flask), readability, large community, libraries (NumPy, Pandas), asyncio',                                                           '66adbf73-a8d2-46dd-ad03-5e7e69ed3367'),
('8bf5cadc-fe6a-4b8c-cadc-d5e6f7a8b9cb', NOW(), NOW(), 'Java',              'java',              'Object-oriented programming, enterprise applications, Android development, robust, platform-independent (JVM), large ecosystem, Spring Framework, Jakarta EE, concurrent programming, performance',                                                    '66adbf73-a8d2-46dd-ad03-5e7e69ed3367'),
('ad07ecee-bf8c-4dae-ecee-a7b9cacbedf5', NOW(), NOW(), 'PHP',               'php',               'Server-side scripting, web development, widely used, content management systems (WordPress, Drupal), frameworks (Laravel, Symfony), LAMP stack, easy deployment, large community',                                                                      '66adbf73-a8d2-46dd-ad03-5e7e69ed3367'),
('cf29ab00-d1a0-4fc0-ab00-c9dbdce0af07', NOW(), NOW(), 'Databases',         'databases',         'Data storage, persistence, information retrieval, DBMS, relational (SQL), non-relational (NoSQL), data modeling, indexing, transactions, ACID, database design',                                                                                       '66adbf73-a8d2-46dd-ad03-5e7e69ed3367'),
('f05cdf33-04d3-42f3-df33-fcfeafb1da00', NOW(), NOW(), 'APIs',              'apis',              'Application Programming Interface, service communication, contract, REST, GraphQL, SOAP, gRPC, API design, documentation (Swagger, OpenAPI), security (OAuth, JWT), API gateway, webhooks',                                                            '66adbf73-a8d2-46dd-ad03-5e7e69ed3367'),
('016de044-15e4-4304-e044-0d0fb0c2eb01', NOW(), NOW(), 'Authentication',    'authentication',    'User identity verification, login, passwords, multi-factor authentication (MFA), OAuth, OpenID Connect, JWT, session management, security best practices, identity providers (IdP)',                                                                    '66adbf73-a8d2-46dd-ad03-5e7e69ed3367'),
('127ef155-26f5-4415-f155-1e1ac1d3fc02', NOW(), NOW(), 'Serverless',        'serverless',        'Backend as a Service (BaaS), Function as a Service (FaaS), pay-per-use, auto-scaling, event-driven architecture, AWS Lambda, Google Cloud Functions, Azure Functions, reduced operational overhead',                                                   '66adbf73-a8d2-46dd-ad03-5e7e69ed3367');`
	DB.Exec(backendSubcategories)

	// Step 4: Insert subcategories for specific backend technologies
	backendTechSubcategories := `INSERT INTO categories (
  id,
  created_at,
  updated_at,
  title,
  slug,
  description,
  parent_id
) VALUES
-- under Python
('7ae4b9cb-ed5f-4a7b-b9cb-c4d5e6f7a8ba', NOW(), NOW(), 'Django',            'django',            'Python web framework, high-level, batteries-included, ORM, admin interface, MVC (MTV), rapid development, security features, scalability, REST framework (DRF)',                                                                       '6fd3a8ba-dc4e-4f6a-a8ba-b3c4d5e6f7a9'),
-- under Java
('9ca6dbdd-af7b-4c9d-dbdd-f6a8b9cadce4', NOW(), NOW(), 'Spring Boot',       'spring-boot',       'Java framework, microservices, convention over configuration, dependency injection, rapid application development, Spring ecosystem (Data, Security, Cloud), embedded servers (Tomcat, Jetty), REST APIs',                                         '8bf5cadc-fe6a-4b8c-cadc-d5e6f7a8b9cb'),
-- under PHP
('be18faff-c09d-4ebf-faff-b8cadcedfe06', NOW(), NOW(), 'Laravel',           'laravel',           'PHP web framework, elegant syntax, MVC architecture, Eloquent ORM, Blade templating engine, Artisan console, routing, middleware, queue management, Vapor (serverless)',                                                                      'ad07ecee-bf8c-4dae-ecee-a7b9cacbedf5'),
-- under Databases
('de3abd11-e2b1-40d1-bd11-dadcedf0bf08', NOW(), NOW(), 'SQL',               'sql',               'Structured Query Language, relational databases, DDL, DML, DCL, querying, joins, indexes, stored procedures, MySQL, PostgreSQL, SQL Server, Oracle',                                                                                       'cf29ab00-d1a0-4fc0-ab00-c9dbdce0af07'),
('ef4bce22-f3c2-41e2-ce22-ebedfea0cf09', NOW(), NOW(), 'NoSQL',             'nosql',             'Non-relational databases, flexible schema, scalability, high availability, document (MongoDB), key-value (Redis), column-family (Cassandra), graph (Neo4j), BASE properties',                                                                 'cf29ab00-d1a0-4fc0-ab00-c9dbdce0af07');`
	DB.Exec(backendTechSubcategories)

	// Step 5: Insert subcategories for tech
	techSubcategories := `INSERT INTO categories (
  id,
  created_at,
  updated_at,
  title,
  slug,
  description,
  parent_id
) VALUES
('c1e710fa-e04b-4405-b0a2-13aa07bfd268', NOW(), NOW(), 'ai',                'ai',                'Artificial intelligence, machine learning, deep learning, neural networks, NLP, computer vision, reinforcement learning, generative AI, transformer models, large language models (LLM), AI ethics, data science, predictive analytics, AI-driven automation', '106f419f-4fa8-4d11-947d-bb0f1aa58dbd');`
	DB.Exec(techSubcategories)

	// Step 6: Insert subcategories for AI
	aiSubcategories := `INSERT INTO categories (
  id,
  created_at,
  updated_at,
  title,
  slug,
  description,
  parent_id
) VALUES
('1a8b3c5d-e7f9-4a1b-b3c5-d8e0f1a2b4c6', NOW(), NOW(), 'Machine Learning', 'machine-learning',    'Algorithms, supervised learning, unsupervised learning, reinforcement learning, predictive modeling, classification, regression, clustering, feature engineering, model training, evaluation metrics, scikit-learn, TensorFlow, PyTorch','c1e710fa-e04b-4405-b0a2-13aa07bfd268'),
('2b9c4d6e-f8a0-4b2c-c4d6-e9f1a2b3c5d7', NOW(), NOW(), 'Natural Language Processing','nlp','Text analysis, sentiment analysis, topic modeling, named entity recognition (NER), chatbots, language translation, speech recognition, text generation, spaCy, NLTK, Hugging Face Transformers','c1e710fa-e04b-4405-b0a2-13aa07bfd268'),
('3ca0d5e7-a9b1-4c3d-d5e7-f0a2b3c4d6e8', NOW(), NOW(), 'Computer Vision',  'computer-vision',   'Image recognition, object detection, image segmentation, facial recognition, video analysis, optical character recognition (OCR), image processing, OpenCV, YOLO, CNNs (Convolutional Neural Networks)',                                   'c1e710fa-e04b-4405-b0a2-13aa07bfd268'),
('4db1e6f8-ba2c-4d4e-e6f8-a1b3c4d5e7f9', NOW(), NOW(), 'Generative AI',     'generative-ai',     'Content creation, image generation (Stable Diffusion, Midjourney), text generation (GPT, Llama), code generation, deepfakes, prompt engineering, foundation models, diffusion models, GANs',                                               'c1e710fa-e04b-4405-b0a2-13aa07bfd268'),
('5ec2f7a9-cb3d-4e5f-f7a9-b2c4d5e6f8a0', NOW(), NOW(), 'Data Science',      'data-science',      'Data analysis, data visualization, statistics, data mining, big data, business intelligence, Jupyter notebooks, Pandas, NumPy, Matplotlib, Seaborn, data storytelling, ETL',                                                                            'c1e710fa-e04b-4405-b0a2-13aa07bfd268');`
	DB.Exec(aiSubcategories)

	// Step 7: Insert missing DevOps categories and subcategories
	devOpsCategories := `INSERT INTO categories (
  id,
  created_at,
  updated_at,
  title,
  slug,
  description,
  parent_id
) VALUES
('3490b377-4817-4637-b377-303ce3f5be04', NOW(), NOW(), 'DevOps',            'devops',            'Development and Operations integration, automation, continuous delivery, deployment pipeline, infrastructure as code (IaC), collaboration, monitoring, site reliability engineering (SRE), observability',                                        NULL);`
	DB.Exec(devOpsCategories)

	devOpsSubcategories := `INSERT INTO categories (
  id,
  created_at,
  updated_at,
  title,
  slug,
  description,
  parent_id
) VALUES
('45a1c488-5928-4748-c488-414df406cf05', NOW(), NOW(), 'CI/CD',             'ci-cd',             'Continuous Integration, Continuous Delivery, Continuous Deployment, automated builds, automated testing, automated deployment, Jenkins, GitLab CI, GitHub Actions, deployment pipeline',                                                          '3490b377-4817-4637-b377-303ce3f5be04'),
('56b2d599-6a39-4859-d599-525ea517da06', NOW(), NOW(), 'Docker',            'docker',            'Containerization, application packaging, isolated environments, lightweight virtualization, Docker Hub, Docker Compose, microservices deployment, consistent environments, portability',                                                         '3490b377-4817-4637-b377-303ce3f5be04'),
('67c3e6aa-7b4a-496a-e6aa-636fb628eb07', NOW(), NOW(), 'Kubernetes',        'kubernetes',        'Container orchestration, K8s, automated deployment, scaling, management of containerized applications, service discovery, load balancing, self-healing, GKE, EKS, AKS',                                                                         '3490b377-4817-4637-b377-303ce3f5be04');`
	DB.Exec(devOpsSubcategories)

	fmt.Println("Categories initialized successfully.")
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
('550e8400-e29b-41d4-a716-446655440039', NOW(), NOW(), 'product40', 'product40', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 15.00),
-- Sản phẩm mới tương ứng với 37 danh mục mới
('550e8400-e29b-41d4-a716-446655440040', NOW(), NOW(), 'product41', 'product41', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 20.00),
('550e8400-e29b-41d4-a716-446655440041', NOW(), NOW(), 'product42', 'product42', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 21.50),
('550e8400-e29b-41d4-a716-446655440042', NOW(), NOW(), 'product43', 'product43', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 23.00),
('550e8400-e29b-41d4-a716-446655440043', NOW(), NOW(), 'product44', 'product44', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 24.50),
('550e8400-e29b-41d4-a716-446655440044', NOW(), NOW(), 'product45', 'product45', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 26.00),
('550e8400-e29b-41d4-a716-446655440045', NOW(), NOW(), 'product46', 'product46', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 27.50),
('550e8400-e29b-41d4-a716-446655440046', NOW(), NOW(), 'product47', 'product47', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 29.00),
('550e8400-e29b-41d4-a716-446655440047', NOW(), NOW(), 'product48', 'product48', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 30.50),
('550e8400-e29b-41d4-a716-446655440048', NOW(), NOW(), 'product49', 'product49', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 32.00),
('550e8400-e29b-41d4-a716-446655440049', NOW(), NOW(), 'product50', 'product50', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 33.50),
('550e8400-e29b-41d4-a716-446655440050', NOW(), NOW(), 'product51', 'product51', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 35.00),
('550e8400-e29b-41d4-a716-446655440051', NOW(), NOW(), 'product52', 'product52', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 36.50),
('550e8400-e29b-41d4-a716-446655440052', NOW(), NOW(), 'product53', 'product53', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 38.00),
('550e8400-e29b-41d4-a716-446655440053', NOW(), NOW(), 'product54', 'product54', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 39.50),
('550e8400-e29b-41d4-a716-446655440054', NOW(), NOW(), 'product55', 'product55', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 41.00),
('550e8400-e29b-41d4-a716-446655440055', NOW(), NOW(), 'product56', 'product56', NULL, NULL, 'ee1b2de9-c8b6-436b-8527-011a036f5fbb', 42.50);
`
	DB.Exec(sql)
}

// hàm khỏi tạo media cho product
func initMediaForProducts() {
	// 1. Kiểm tra xem media cho product đã tồn tại chưa
	var count int64
	// Sử dụng models.Media để kiểm tra bảng media
	// Chỉ đếm media liên quan đến 'products' để chính xác hơn
	if err := DB.Model(&models.Media{}).Where("related_type = ?", "products").Count(&count).Error; err != nil {
		fmt.Printf("Error checking media count: %v\n", err)
		return
	}

	if count > 0 {
		fmt.Println("Media data for products seems to exist already. Skipping media initialization.")
		return
	}

	fmt.Println("Initializing media data for products...")

	// 2. Lấy danh sách Product ID (sử dụng lại từ initProduct)
	var productIDs []string
	DB.Model(models.Product{}).Select("id").Find(&productIDs)
	// 3. Chuẩn bị các giá trị cho câu lệnh INSERT
	var valueStrings []string
	// Template cho VALUES: (created_at, updated_at, file, file_type, related_id, related_type)
	// ID là autoIncrement nên không cần thêm. Status có default 'using'.
	// Lưu ý: Sử dụng tên cột trong DB (thường là snake_case)
	sqlTemplate := "(NOW(), NOW(), '%s', '%s', '%s', 'products')"

	imageFile := "products/cac_mau_website_du_lich_dep_2_1744627245292.webp"
	downloadFile := "products/schedule_1744451078410.zip"

	for _, productID := range productIDs {
		// Thêm media kiểu image
		valueStrings = append(valueStrings, fmt.Sprintf(sqlTemplate, imageFile, "image", productID))
		// Thêm media kiểu file_download
		valueStrings = append(valueStrings, fmt.Sprintf(sqlTemplate, downloadFile, "file_download", productID))
	}

	// 4. Xây dựng câu lệnh SQL hoàn chỉnh
	// Đảm bảo tên bảng 'media' và các cột là chính xác
	sql := fmt.Sprintf("INSERT INTO media (created_at, updated_at, file, file_type, related_id, related_type) VALUES %s",
		strings.Join(valueStrings, ",\n")) // Nối các VALUES lại, thêm newline cho dễ đọc

	// 5. Thực thi câu lệnh SQL
	if err := DB.Exec(sql).Error; err != nil {
		fmt.Printf("Error initializing media data: %v\n", err)
	} else {
		fmt.Printf("Initialized %d media records for products successfully.\n", len(valueStrings))
	}
}

func initProductCategories() {
	// Kiểm tra nếu đã có dữ liệu
	var count int64
	if err := DB.Model(&models.ProductCategory{}).Count(&count).Error; err != nil {
		fmt.Printf("Error checking product_categories count: %v\n", err)
		return
	}
	if count > 0 {
		fmt.Println("Product category data already exists. Skipping initialization.")
		return
	}

	// Lấy danh sách category và product
	var categories []string
	if err := DB.Model(&models.Category{}).Select("id").Find(&categories).Error; err != nil {
		fmt.Printf("Error fetching categories: %v\n", err)
		return
	}
	if len(categories) == 0 {
		fmt.Println("No categories found. Skipping initialization.")
		return
	}

	var products []string
	if err := DB.Model(&models.Product{}).Select("id").Find(&products).Error; err != nil {
		fmt.Printf("Error fetching products: %v\n", err)
		return
	}
	if len(products) == 0 {
		fmt.Println("No products found. Skipping initialization.")
		return
	}

	// Kiểm tra số lượng categories và products bằng nhau
	if len(categories) != len(products) {
		fmt.Printf("Error: Number of categories (%d) and products (%d) must be equal.\n", len(categories), len(products))
		return
	}

	fmt.Printf("Initializing product categories... (Categories: %d, Products: %d)\n", len(categories), len(products))

	// Khởi tạo seed cho random
	rand.Seed(time.Now().UnixNano())

	// Map để theo dõi các cặp (product, category) đã gán
	productCategorySet := make(map[string]struct{})
	// Slice để lưu các bản ghi ProductCategory
	var productCategories []models.ProductCategory

	// Bước 1: Tạo cặp ProductCategory theo chỉ số để đảm bảo mỗi category và product có ít nhất 1 liên kết
	for i := 0; i < len(categories); i++ {
		productID := products[i]
		categoryID := categories[i]
		key := productID + ":" + categoryID
		productCategorySet[key] = struct{}{}
		productCategories = append(productCategories, models.ProductCategory{
			ProductID:  uuid.MustParse(productID),
			CategoryID: uuid.MustParse(categoryID),
		})
	}

	// Bước 2: Gán ngẫu nhiên thêm category cho product
	for _, product := range products {
		// Gán thêm tối đa 3 category ngẫu nhiên (có thể điều chỉnh)
		numCategories := rand.Intn(4) // 0 đến 3 category bổ sung
		if numCategories == 0 {
			continue
		}
		// Lấy danh sách category chưa được gán cho product này
		availableCategories := make([]string, 0, len(categories))
		for _, category := range categories {
			key := product + ":" + category
			if _, exists := productCategorySet[key]; !exists {
				availableCategories = append(availableCategories, category)
			}
		}
		// Chọn ngẫu nhiên từ các category còn lại
		selectedCategories := randomSample(availableCategories, numCategories)
		for _, category := range selectedCategories {
			key := product + ":" + category
			productCategorySet[key] = struct{}{}
			productCategories = append(productCategories, models.ProductCategory{
				ProductID:  uuid.MustParse(product),
				CategoryID: uuid.MustParse(category),
			})
		}
	}

	// Bước 3: Chèn dữ liệu vào bảng product_categories
	if len(productCategories) > 0 {
		if err := DB.Create(&productCategories).Error; err != nil {
			fmt.Printf("Error inserting product categories: %v\n", err)
		} else {
			fmt.Printf("Initialized %d product category records successfully.\n", len(productCategories))
		}
	} else {
		fmt.Println("No product category records to insert.")
	}
}

// Hàm randomSample để chọn ngẫu nhiên k phần tử từ slice
func randomSample(items []string, k int) []string {
	if k > len(items) {
		k = len(items)
	}
	if k <= 0 {
		return nil
	}

	// Xáo trộn và lấy k phần tử đầu
	indices := make([]int, len(items))
	for i := range items {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	result := make([]string, k)
	for i := 0; i < k; i++ {
		result[i] = items[indices[i]]
	}
	return result
}
