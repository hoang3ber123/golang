import spacy

# Danh sách categories với description tối ưu
categories = [
    {
        'id': '1', 
        'title': 'backend', 
        'description': 'Server-side, APIs, databases, RESTful, GraphQL, authentication, authorization, MVC, ORM, backend development, middleware, caching, scalability, backend architecture, load balancing, API gateway'
    },
    {
        'id': '3', 
        'title': 'nodejs', 
        'description': 'JavaScript runtime, scalable backend, Express.js, NestJS, event-driven, asynchronous programming, non-blocking I/O, V8 engine, serverless, WebSockets, microservices, API development, package management (npm, yarn)'
    },
    {
        'id': '5', 
        'title': 'microservices', 
        'description': 'Distributed services, architecture, containerization, Kubernetes, Docker, service discovery, API gateway, inter-service communication, CQRS, event sourcing, decentralized architecture, domain-driven design (DDD), fault tolerance'
    },
    {
        'id': '6', 
        'title': 'performance optimization', 
        'description': 'Speed, latency, efficiency, caching strategies, database indexing, load balancing, CDN, query optimization, profiling, memory management, parallel processing, multithreading, performance monitoring, response time, request throttling'
    },
    {
        'id': '7', 
        'title': 'budget-friendly', 
        'description': 'Low cost, affordable solutions, cost optimization, open-source tools, free-tier services, cloud cost management, pricing models, SaaS alternatives, resource efficiency, cost-effective hosting, serverless pricing, FinOps'
    },
    {
        'id': '2', 
        'title': 'frontend', 
        'description': 'UI, client-side, responsive design, HTML, CSS, JavaScript, TypeScript, CSS frameworks (Bootstrap, Tailwind, Material UI), UI components, animations, DOM manipulation, progressive web apps (PWA), single-page applications (SPA), WebAssembly, frontend optimization'
    },
    {
        'id': '4', 
        'title': 'react', 
        'description': 'Interactive UI, JavaScript library, React.js, React Native, JSX, virtual DOM, component-based architecture, hooks, state management (Redux, Recoil, Zustand), SSR (Next.js), hydration, reconciliation, frontend framework'
    },
    {
        'id': '8', 
        'title': 'ai', 
        'description': 'Artificial intelligence, machine learning, deep learning, neural networks, NLP, computer vision, reinforcement learning, generative AI, transformer models, large language models (LLM), AI ethics, data science, predictive analytics, AI-driven automation'
    },
    {
        'id': '9', 
        'title': 'ecommerce', 
        'description': 'Online stores, shopping, transactions, payment gateways (Stripe, PayPal, VNPay), shopping cart, checkout flow, order management, product catalog, customer reviews, dropshipping, marketplace, subscription model, SEO for ecommerce, user experience (UX), conversion rate optimization (CRO)'
    },
    {
        'id': '10', 
        'title': 'saas', 
        'description': 'Software as a Service, cloud solutions, multi-tenant architecture, subscription-based, SaaS pricing models, API-first development, microservices for SaaS, customer onboarding, usage analytics, scalability, CI/CD, DevOps, cloud hosting (AWS, Azure, GCP), security compliance (SOC 2, GDPR)'
    },
    {
        'id': '11', 
        'title': 'portfolio', 
        'description': 'Showcase projects, personal branding, web portfolio, design portfolio, developer portfolio, case studies, UI/UX presentation, interactive resume, testimonials, online presence, custom domain, SEO optimization, responsive design'
    },
    {
        'id': '12', 
        'title': 'blog', 
        'description': 'Content writing, publishing, news, CMS (WordPress, Ghost, Strapi), Markdown, SEO, social media integration, email newsletters, blog monetization, affiliate marketing, audience engagement, blog analytics, content strategy, editorial workflow'
    },
    {
        'id': '13', 
        'title': 'landing-page', 
        'description': 'Marketing, conversions, lead generation, sales funnel, call-to-action (CTA), A/B testing, copywriting, UI/UX design, high-converting pages, one-page websites, performance tracking, Google Ads, Facebook Pixel, SEO optimization'
    },
    {
        'id': '14', 
        'title': 'news', 
        'description': 'Media, articles, latest updates, journalism, online magazines, breaking news, press releases, RSS feeds, news aggregation, real-time updates, media coverage, social media trends, digital publishing, fact-checking'
    },
    {
        'id': '15', 
        'title': 'real-estate', 
        'description': 'Property listings, real estate solutions, rental properties, commercial real estate, mortgage calculators, house valuation, property management, real estate CRM, MLS (Multiple Listing Service), home-buying process, real estate investments, virtual tours'
    },
    {
        'id': '16', 
        'title': 'web3', 
        'description': 'Decentralized applications, blockchain, smart contracts, Ethereum, NFTs, DeFi (Decentralized Finance), DAOs (Decentralized Autonomous Organizations), tokenomics, crypto wallets, metaverse, on-chain governance, Web3 authentication, Layer 2 scaling solutions'
    },
    {
        'id': '17', 
        'title': 'startup', 
        'description': 'Entrepreneurship, business growth, startup funding, venture capital, bootstrapping, business model canvas, go-to-market strategy, pitch decks, MVP (Minimum Viable Product), customer acquisition, product-market fit, accelerator programs, startup scaling'
    },
    {
        'id': '18', 
        'title': 'tech', 
        'description': 'Technology, innovations, IT, artificial intelligence, cloud computing, cybersecurity, data science, IoT (Internet of Things), big data, quantum computing, 5G networks, emerging technologies, IT infrastructure, digital transformation'
    },
    {
        'id': '19', 
        'title': 'modern', 
        'description': 'Contemporary design, latest trends, minimalism, futuristic UI, neomorphic design, glassmorphism, dark mode, responsive layouts, creative direction, user-centric design, modern typography, digital aesthetics, web trends'
    },
    {
        'id': '20', 
        'title': 'animated', 
        'description': 'Motion graphics, interactive UI, CSS animations, Lottie animations, SVG animations, microinteractions, transitions, parallax effects, 3D animations, WebGL, After Effects, real-time rendering, immersive user experience'
    },]


# Tải mô hình ngôn ngữ lớn hơn
nlp = spacy.load("en_core_web_md")

def predict_categories(text, threshold=0.60,rank=10):
    """
    Dự đoán danh mục dựa trên độ tương đồng ngữ nghĩa, tối ưu với mô tả ngắn gọn.
    
    Args:
        text (str): Văn bản đầu vào từ người dùng
        threshold (float): Ngưỡng độ tương đồng tối thiểu
    
    Returns:
        list: Danh sách các danh mục phù hợp
    """
    # Xử lý văn bản đầu vào, loại từ dừng
    input_doc = nlp(text.lower())
    input_text = " ".join([token.text for token in input_doc if not token.is_stop])
    input_doc = nlp(input_text)
    
    # Tính độ tương đồng với từng danh mục
    matched_categories = []
    for category in categories:
        # Kết hợp title và description, loại từ dừng
        category_text = f"{category['title']} {category['description']}".lower()
        category_doc = nlp(category_text)
        
        # Tính độ tương đồng
        similarity = input_doc.similarity(category_doc)
        if similarity >= threshold:
            matched_categories.append((category['title'], similarity))
    
    # Sắp xếp theo độ tương đồng và trả về
    matched_categories.sort(key=lambda x: x[1], reverse=True)
    return [cat[0] for cat in matched_categories]

def main():
    test_inputs = [
        "recommend for me product with low cost",
        "i want some prod optimizing speed and using microservice architecture",
        "give me something for building user interfaces"
    ]
    for test_input in test_inputs:
        print(f">>> predict_categories(\"{test_input}\")")
        result = predict_categories(test_input)
        print(result)

if __name__ == "__main__":
    main()