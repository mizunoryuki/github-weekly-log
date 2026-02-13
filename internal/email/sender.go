package email

import (
	"bytes"
	"fmt"
	"github-weekly-log/internal/github"
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
func SendWeeklyReport(apiKey string, htmlContent string, imagePath string) error {
	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "Acme <onboarding@resend.dev>",
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
