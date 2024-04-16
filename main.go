package main

import (
	"crypto/tls"
	"log"
	"os"
	"sync"
	"time"

	gomail "gopkg.in/mail.v2"
	"gopkg.in/yaml.v2"

	"fmt"
)

func main() {
	list, err := loadEmailList("emails.yaml")
	if err != nil {
		log.Fatalf("Error loading email list: %v", err)
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(list.Emails))

	for i := 0; i < len(list.Emails); i += 10 {
		end := i + 10
		if end > len(list.Emails) {
			end = len(list.Emails)
		}
		for _, email := range list.Emails[i:end] {
			wg.Add(1)
			go func(email string) {
				defer wg.Done()
				if err := sendEmail(email); err != nil {
					errCh <- err
				}
			}(email)
		}
		fmt.Println("Batch Ran successfully")
		wg.Wait()
	}
	// Check for errors
	close(errCh)
	for err := range errCh {
		log.Printf("Error: %v", err)
	}

}

type EmailList struct {
	Emails []string `yaml:"emails"`
}

// loadEmailList loads the list of emails from the YAML file
func loadEmailList(filePath string) (*EmailList, error) {
	var list EmailList
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &list)
	return &list, err
}

func sendEmail(recipient string) error {

	// Read the contents of the HTML file
	content, err := os.ReadFile("mail.html") // Ensure the file path is correct
	if err != nil {
		log.Fatal(err)
	}

	// Convert the content to a string and store it in the variable
	mailContent := string(content)

	m := gomail.NewMessage()
	// Set E-Mail sender
	m.SetAddressHeader("From", "support@influenza.com", "Influenza")

	// Set E-Mail receivers
	m.SetHeader("To", recipient)

	// Set E-Mail subject
	m.SetHeader("Subject", "subject")

	cid := "image-cid" // The Content-ID for the image
	mailContent = fmt.Sprintf(mailContent, cid)

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/html", mailContent)
	// Attach the image with the corresponding Content-ID
	m.Embed("<IMAGE>", gomail.SetHeader(map[string][]string{"Content-ID": {"<" + cid + ">"}}))

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "<usermail>", "<password>")
	d.Timeout = 100 * time.Second
	d.TLSConfig = &tls.Config{ServerName: "smtp.gmail.com"}

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		return (err)
	}
	return nil
}
