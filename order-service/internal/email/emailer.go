package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"order-service/config"
	"order-service/internal/db"
	"order-service/internal/models"
	"path/filepath"
)

func SendPaymentCheckout(user models.User, productIDS []string, order models.Order) {
	fmt.Println("order in mail:", order)
	products := make([]models.Product, len(productIDS))
	if err := db.DB.Where("id IN ?", productIDS).Find(&products).Error; err != nil {
		fmt.Println("Error happpend when senpayment checkout with filter products:", err.Error())
		return
	}
	// SMTP Server Info
	smtpServer := config.Config.SMTPServer
	smtpPort := config.Config.SMTPPort
	smtpUser := config.Config.SMTPUsername
	smtpPass := config.Config.SMTPPassword
	basePath := config.Config.BasePath

	// Load template
	htmlFilePath := filepath.Join(basePath, "internal", "templates", "email", "payment.html")
	tmpl, err := template.ParseFiles(htmlFilePath)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Dữ liệu truyền vào template
	data := map[string]interface{}{
		"UserName":      user.Name,
		"UserEmail":     user.Email,
		"PaymentDate":   order.CreatedAt.Format("2006-01-02 15:04:05"), // Format chuẩn
		"Products":      products,
		"AmountPaid":    order.AmountPaid,
		"PaymentMethod": order.PaymentMethod,
		"TransactionID": order.TransactionID,
	}

	// Render template
	var body bytes.Buffer
	err = tmpl.Execute(&body, data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	// Email Headers
	subject := "Subject: Payment Confirmation\r\n"
	msg := []byte(subject + "Content-Type: text/html; charset=UTF-8\r\n\r\n" + body.String())

	// Setup TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer,
	}

	// Connect to SMTP server using TLS
	conn, err := tls.Dial("tcp", smtpServer+":"+smtpPort, tlsConfig)
	if err != nil {
		fmt.Println("Error connecting to SMTP server:", err)
		return
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		fmt.Println("Error creating SMTP client:", err)
		return
	}
	defer client.Quit()

	// Authenticate
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpServer)
	if err = client.Auth(auth); err != nil {
		fmt.Println("Error authenticating:", err)
		return
	}

	// Set sender and recipient
	if err = client.Mail(smtpUser); err != nil {
		fmt.Println("Error setting sender:", err)
		return
	}

	if err = client.Rcpt(user.Email); err != nil {
		fmt.Println("Error setting recipient:", err)
		return
	}

	// Send email
	w, err := client.Data()
	if err != nil {
		fmt.Println("Error opening data:", err)
		return
	}

	_, err = w.Write(msg)
	if err != nil {
		fmt.Println("Error writing message:", err)
		return
	}

	err = w.Close()
	if err != nil {
		fmt.Println("Error closing write:", err)
		return
	}

	fmt.Println("Email sent successfully to", user.Email)
}
