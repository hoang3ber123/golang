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
	"github.com/google/uuid"
	proto_product "github.com/hoang3ber123/proto-golang/product"
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
	// duyệt mảng tạo proto gửi vào
	productsInfoRequest := make([]*proto_product.ProductsInfoRequest, len(serializer.Cart))
	for index, item := range serializer.Cart {
		productsInfoRequest[index] = &proto_product.ProductsInfoRequest{
			RelatedId:   item.RelatedID,
			RelatedType: item.RelatedType,
		}
	}
	// Gửi grpc để kiểm tra thử danh sách product
	products, err := grpcclient.GetProductsInfo(productsInfoRequest)
	if err != nil {
		return err.Send(c)
	}
	// Lấy thông tin user từ context
	user := c.Locals("user").(*models.User) // Giả sử user đã được middleware xác thực và lưu vào context
	// Tạo danh sách Line Items cho Stripe Checkout
	if len(products) == 0 {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "product is empty").Send(c)
	}

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
	productsJsonData, _ := json.Marshal(productsInfoRequest)
	productStringData := string(productsJsonData)
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Metadata: map[string]string{
			"user":     userStringData,
			"products": productStringData,
		},
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String("http://127.0.0.1:8082/v1/payment/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("http://127.0.0.1:8082/v1/payment/cancel?session_id={CHECKOUT_SESSION_ID}"),
	}

	s, errSesion := session.New(params)
	if errSesion != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, errSesion.Error()).Send(c)
	}
	// Lưu lại order
	services.CreateOrder(user.ID.String(), products, "stripe", s.ID, amountPaid)
	// Trả về url front end
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
	var productsInfoRequest []*proto_product.ProductsInfoRequest
	json.Unmarshal([]byte(s.Metadata["user"]), &user)
	json.Unmarshal([]byte(s.Metadata["products"]), &productsInfoRequest)
	// lọc mảng id từ productsInfoRequest
	productIDS := make([]string, len(productsInfoRequest))
	for index, item := range productsInfoRequest {
		productIDS[index] = item.RelatedId
	}
	// Xóa product khỏi cart
	go grpcclient.ClearCartAfterCheckout(productsInfoRequest, user.ID.String()) // gọi hàm gửi mail thông báo payment trả thành công
	go email.SendPaymentCheckout(user, productIDS, order)
	// trả về dashboard của front end
	return c.Redirect("http://127.0.0.1:3000/customer/orders", fiber.StatusSeeOther)
}

func PaymentCancel(c *fiber.Ctx) error {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Session ID is required").Send(c)
	}

	// Thêm tham số Expand để lấy đầy đủ thông tin PaymentIntent
	expand := "payment_intent"
	params := &stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Expand: []*string{&expand},
		},
	}

	// Lấy thông tin phiên Checkout
	s, err := session.Get(sessionID, params)
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to retrieve session: "+err.Error()).Send(c)
	}

	// Tìm đơn hàng theo TransactionID
	var order models.Order
	if err := db.DB.Where("transaction_id = ?", sessionID).First(&order).Error; err != nil {
		log.Printf("Cannot find order with transaction_id %s: %v", sessionID, err)
	}

	// Kiểm tra trạng thái phiên Checkout
	if s.Status == "open" || s.PaymentIntent == nil || (s.PaymentIntent != nil && s.PaymentIntent.Status != "succeeded") {
		// Khách hàng đã hủy hoặc thanh toán chưa hoàn tất
		if order.ID != uuid.Nil { // Chỉ cập nhật nếu đơn hàng tồn tại
			if err := db.DB.Model(&models.Order{}).
				Where("transaction_id = ?", sessionID).
				Updates(map[string]interface{}{
					"payment_status": "cancel", // Trạng thái hủy
				}).Error; err != nil {
				log.Printf("Error updating payment_status for order %s: %v", sessionID, err)
			} else {
				log.Printf("Payment status updated to 'canceled' for order %s", sessionID)
			}
		}
	}

	// Chuyển hướng về dashboard của frontend
	return c.Redirect("http://127.0.0.1:3000/customer/orders", fiber.StatusSeeOther)
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
