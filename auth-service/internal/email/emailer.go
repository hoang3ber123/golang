package email

import (
	"auth-service/config"
	"auth-service/internal/services"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"text/template"

	"github.com/google/uuid"
)

func SendVerifyMail(user_id uuid.UUID, receiver string) {
	auth := smtp.PlainAuth(
		"",
		config.Config.SMTPUsername,
		config.Config.SMTPPassword,
		config.Config.SMTPHost,
	)

	// generate token
	token, _ := services.GenerateTokenVerifyEmailJWT(user_id)
	// liên kết đến đường dẫn để verify email
	verifyURL := "http://localhost:8080/verify-email/" + token

	// Đọc file HTML chứa template email
	htmlContent, err := ioutil.ReadFile("internal/templates/email/verify_email.html")
	if err != nil {
		fmt.Println("Error reading HTML file:", err)
		return
	}

	// Tạo template từ nội dung của file HTML
	tmpl, err := template.New("verify_email").Parse(string(htmlContent))
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Render template với đường dẫn xác thực
	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]interface{}{
		"VerificationLink": verifyURL,
	})
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	// Đặt tiêu đề cho email
	subject := "Subject: Verify your email address"

	// Gửi email với template HTML
	msg := []byte(subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" + body.String())

	err = smtp.SendMail(
		config.Config.SMTPHost+":587",
		auth,
		config.Config.SMTPUsername,
		[]string{receiver},
		msg,
	)

	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}
	fmt.Println("Email sent successfully to", receiver)
}
