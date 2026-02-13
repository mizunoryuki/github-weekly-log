package email

import (
	"bytes"
	"fmt"
	"github-weekly-log/internal/github"
	"os"
	"text/template"

	"github.com/resend/resend-go/v3"
)

// templateを読み込み、ファイルにデータを埋め込む
func LoadTemplate(stats github.WeeklyStats) (string, error) {
	tmpl, err := template.ParseFiles("templates/dist/weekly.html")
	if err != nil {
		return "", err
	}

	// データをテンプレートに埋め込む
	var buf bytes.Buffer
	// Executeでデータを埋め込む
	if err := tmpl.Execute(&buf, stats); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// メール送信
func SendWeeklyReport(apiKey string, htmlContent string, imagePath string, emailDomain string) error {
	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "Acme <" + emailDomain + ">",
		To:      []string{"delivered@resend.dev"},
		Html:    htmlContent,
		Subject: "週間コミットレポート",
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println(sent.Id)

	return nil
}

func TestSend() error {
	htmlByte, err := os.ReadFile("templates/dist/test.html")
	if err != nil {
		return err
	}
	apiKey := os.Getenv("RESEND_API_KEY")
	emailDomain := os.Getenv("RESEND_EMAIL_DOMAIN")
	emailFrom := "Acme <" + emailDomain + ">"
	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    emailFrom,
		To:      []string{"kifimya@instmail.uk"},
		Subject: "【Test】Weekly Log System Connection Check",
		Html:    string(htmlByte),
	}

	_, err = client.Emails.Send(params)
	return err
}
