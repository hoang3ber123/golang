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
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
	"gorm.io/gorm"
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
	if len(products) == 0 {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "product is empty").Send(c)
	}
	fmt.Println("products:", products)
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
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
	}

	// Lấy PaymentIntent từ session
	piParams := &stripe.PaymentIntentParams{}
	pi, err := paymentintent.Get(s.PaymentIntent.ID, piParams)
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to retrieve payment intent: "+err.Error()).Send(c)
	}

	// Lấy charge_id từ LatestCharge
	chargeID := pi.LatestCharge.ID
	if chargeID == "" {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "No charge found for this payment intent").Send(c)
	}

	// Kiểm tra trạng thái thông tin hóa đơn để cập nhật
	// Tìm đơn hàng theo TransactionID
	var order models.Order
	if err := db.DB.Where("transaction_id = ?", sessionID).First(&order).Error; err != nil {
		log.Println("can't not find order:", err)
	}

	// Nếu thanh toán thành công, cập nhật trạng thái
	if s.PaymentIntent != nil && s.PaymentIntent.Status == "succeeded" {
		if err := db.DB.Model(&models.Order{}).
			Where("transaction_id = ?", s.ID).
			Updates(map[string]interface{}{
				"payment_status": "success",
				"charge_id":      chargeID,
			}).Error; err != nil {
			log.Println("Error when updating payment_status and charge_id:", err)
		} else {
			log.Println("Payment status and charge_id updated successfully!")
		}
	}

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

// PaymentRefund xử lý yêu cầu hoàn tiền cho một giao dịch
func PaymentRefund(c *fiber.Ctx) error {
	// Lấy payment_id từ query params hoặc body (tùy bạn thiết kế API)
	orderID := c.Params("id")
	if orderID == "" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "id is required").Send(c)
	}
	serializer := new(serializers.RefundPaymentSerializer)
	// Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	var order models.Order

	err := db.DB.First(&order, "id = ?", orderID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Order not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusNotFound, "Database error: "+err.Error()).Send(c)
	}

	// kiểm tra trạng thái order có thành công trước khi refund
	if order.PaymentStatus != "success" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "This order was not success paid for refund").Send(c)
	}

	// Tạo request refund tới Stripe
	// Sử dụng StripeChargeID từ database để xác định giao dịch cần hoàn tiền
	refundParams := &stripe.RefundParams{
		Charge: stripe.String(order.ChargeID),
		// Amount: stripe.Int64(1000), // Nếu muốn hoàn tiền một phần (tính bằng cents)
		Reason: stripe.String(serializer.Reason), // Lý do hoàn tiền, tùy chọn
	}

	// Gọi Stripe API để thực hiện refund
	if _, err := refund.New(refundParams); err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Refund failed: "+err.Error()).Send(c)
	}

	// (Tùy chọn) Cập nhật trạng thái trong database nếu cần
	// Ví dụ: thêm cột 'refunded' vào bảng payments và cập nhật nó
	err = db.DB.Model(&order).Update("payment_status", "refunded").Error
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to update refund status: "+err.Error()).Send(c)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, "Refund successfully").Send(c)
}
