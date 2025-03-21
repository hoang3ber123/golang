package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/db"
	"order-service/internal/email"
	grpcclient "order-service/internal/grpc_client"
	"order-service/internal/models"
	"order-service/internal/responses"
	"order-service/internal/serializers"
	"order-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

func PaymentCreate(c *fiber.Ctx) error {
	// Kiểm tra danh sách product
	serializer := new(serializers.CreateOrderSerializer)

	// Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}
	// Gửi grpc để kiểm tra thử danh sách product
	products, err := grpcclient.GetProductsInCartRequest(serializer)
	if err != nil {
		return err.Send(c)
	}
	// Lấy thông tin user từ context
	user := c.Locals("user").(*models.User) // Giả sử user đã được middleware xác thực và lưu vào context
	// Tạo danh sách Line Items cho Stripe Checkout
	lineItems := make([]*stripe.CheckoutSessionLineItemParams, len(products))
	var amountPaid float64
	for prod_index, product := range products {
		image := product.Image
		images := []*string{&image}
		if image == "" {
			images = nil
		}
		// Tạo slice Images với URL từ product.Image
		lineItem := &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("usd"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name:   stripe.String(fmt.Sprintf("Name:%s", product.Title)),
					Images: images,
				},
				UnitAmount: stripe.Int64(int64(product.Price * 100)), // Convert float64 to int64 (cents)
			},
			Quantity: stripe.Int64(1),
		}
		lineItems[prod_index] = lineItem
		amountPaid += product.Price
	}

	// Tạo session với Stripe
	// Chuyển đổi dữ liệu thành json để đưa vào metadata
	userJsonData, _ := json.Marshal(user)
	userStringData := string(userJsonData)
	productsJsonData, _ := json.Marshal(products)
	productStringData := string(productsJsonData)
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Metadata: map[string]string{
			"user":     userStringData,
			"products": productStringData,
		},
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String("http://localhost:8082/v1/payment/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("http://localhost:8082/v1/payment/cancel?session_id={CHECKOUT_SESSION_ID}"),
	}

	s, errSesion := session.New(params)
	if errSesion != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, errSesion.Error()).Send(c)
	}
	// Lưu lại order
	services.CreateOrder(user.ID.String(), products, "stripe", s.ID, amountPaid)
	// Trả về url front end
	fmt.Println("URL checkout:", s.URL)
	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"payment_link": s.URL,
	}).Send(c)
}

func PaymentSuccess(c *fiber.Ctx) error {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "session is required").Send(c)
	}

	// Thêm tham số Expand để lấy đầy đủ thông tin PaymentIntent
	expand := "payment_intent"
	params := &stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Expand: []*string{&expand}, // Expand PaymentIntent
		},
	}
	// Lấy thông tin hóa đơn thanh toán
	s, err := session.Get(sessionID, params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	// Kiểm tra trạng thái thông tin hóa đơn để cập nhật
	// Tìm đơn hàng theo TransactionID
	var order models.Order
	if err := db.DB.Where("transaction_id = ?", sessionID).First(&order).Error; err != nil {
		log.Println("can't not find order:", err)
	}

	// Nếu thanh toán thành công, cập nhật trạng thái
	if s.PaymentIntent.Status == "succeeded" {
		if err := db.DB.Model(&order).
			Where("transaction_id = ?", sessionID).
			Update("payment_status", "success").Error; err != nil {
			log.Println("Error when update payment_status:", err)
		} else {
			log.Println("update successfully!")
		}
	}
	// Lấy thông tin
	response := map[string]interface{}{
		"payment_intent_id": s.PaymentIntent.ID,
		"amount":            s.PaymentIntent.Amount,
		"currency":          s.PaymentIntent.Currency,
		"status":            s.PaymentIntent.Status,
		"metadata":          s.Metadata,
	}
	fmt.Println("responsedataa:", response)
	// đổi dữ liệu user và products trong metadata thành struct rồi gửi mail
	var user models.User
	var products []models.Product
	json.Unmarshal([]byte(s.Metadata["user"]), &user)
	json.Unmarshal([]byte(s.Metadata["products"]), &products)
	// gọi hàm gửi mail thông báo payment trả thành công
	fmt.Println("order sau khi lưu:", order)
	go email.SendPaymentCheckout(user, products, order)
	// trả về dashboard của front end
	return c.Redirect("http://localhost:3000/dashboard", fiber.StatusSeeOther)
}

func PaymentCancel(c *fiber.Ctx) error {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "session is required").Send(c)
	}

	// Thêm tham số Expand để lấy đầy đủ thông tin PaymentIntent
	expand := "payment_intent"
	params := &stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Expand: []*string{&expand}, // Expand PaymentIntent
		},
	}
	// Lấy thông tin hóa đơn thanh toán
	s, err := session.Get(sessionID, params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	// Kiểm tra trạng thái thông tin hóa đơn để cập nhật
	// Tìm đơn hàng theo TransactionID
	var order models.Order
	if err := db.DB.Where("transaction_id = ?", sessionID).First(&order).Error; err != nil {
		log.Println("can't not find order:", err)
	}

	// Nếu thanh toán thành công, cập nhật trạng thái
	if s.PaymentIntent.Status == "cancel" {
		if err := db.DB.Model(&order).
			Where("transaction_id = ?", sessionID).
			Update("payment_status", "cancel").Error; err != nil {
			log.Println("Error when update payment_status:", err)
		} else {
			log.Println("update successfully!")
		}
	}
	// Lấy thông tin
	response := map[string]interface{}{
		"payment_intent_id": s.PaymentIntent.ID,
		"amount":            s.PaymentIntent.Amount,
		"currency":          s.PaymentIntent.Currency,
		"status":            s.PaymentIntent.Status,
		"metadata":          s.Metadata,
	}
	fmt.Println("responsedataa:", response)
	// trả về dashboard của front end
	return c.Redirect("http://localhost:3000/dashboard", fiber.StatusSeeOther)
}
