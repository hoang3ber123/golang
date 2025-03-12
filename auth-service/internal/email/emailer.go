package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"path/filepath"
	"text/template"

	"auth-service/config"
	"auth-service/internal/services"

	"github.com/google/uuid"
)

func SendVerifyMail(user_id uuid.UUID, receiver string) {
	// SMTP Server Info
	smtpServer := config.Config.SMTPServer
	smtpPort := config.Config.SMTPPort
	smtpUser := config.Config.SMTPUsername
	smtpPass := config.Config.SMTPPassword
	basePath := config.Config.BasePath

	// Generate token
	token, _ := services.GenerateTokenVerifyEmailJWT(user_id)
	verifyURL := "http://localhost:3000/verify-email/" + token

	htmlFilePath := filepath.Join(basePath, "internal", "templates", "email", "verify_email.html")
	htmlContent, err := ioutil.ReadFile(htmlFilePath)
	if err != nil {
		fmt.Println("Error reading HTML file:", err)
		return
	}

	tmpl, err := template.New("verify_email").Parse(string(htmlContent))
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]interface{}{
		"VerificationLink": verifyURL,
	})
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	// Email Headers
	subject := "Subject: Verify your email address\r\n"
	msg := []byte(subject + "Content-Type: text/html; charset=UTF-8\r\n\r\n" + body.String())

	// Setup TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Chỉ dùng nếu gặp lỗi TLS (tạm thời)
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

	if err = client.Rcpt(receiver); err != nil {
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

	fmt.Println("Email sent successfully to", receiver)
}
