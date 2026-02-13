package email

import (
	"bytes"
	"fmt"
	"github-weekly-log/internal/github"
	"os"
	"text/template"
	"time"

	"github.com/resend/resend-go/v3"
)

// templateを読み込み、ファイルにデータを埋め込む
func LoadTemplate(stats github.WeeklyStats) (string, error) {
	tmpl, err := template.ParseFiles("templates/dist/weekly.html")
	if err != nil {
		return "", fmt.Errorf("テンプレート読み込み失敗: %w", err)
	}

	// データをテンプレートに埋め込む
	var buf bytes.Buffer
	// Executeでデータを埋め込む
	if err := tmpl.Execute(&buf, stats); err != nil {
		return "", fmt.Errorf("テンプレート実行失敗: %w", err)
	}

	return buf.String(), nil
}

// メール送信
func SendWeeklyReport(apiKey string, htmlContent string, imagePath string, emailDomain string, emailTo string) error {
	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "Acme <" + emailDomain + ">",
		To:      []string{emailTo},
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
	emailTo := os.Getenv("TEST_RESEND_EMAIL_TO")
	emailFrom := "Acme <" + emailDomain + ">"
	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    emailFrom,
		To:      []string{emailTo},
		Subject: "【Test】Weekly Log System Connection Check",
		Html:    string(htmlByte),
	}

	_, err = client.Emails.Send(params)
	return err
}

func TestWeeklyMailSend() error {

	now := time.Now()

	// モックデータ
	stats := github.WeeklyStats{
		TotalCommits: 42,
		ActiveDays:   5,
		StartDate:    now.AddDate(0, 0, -7),
		EndDate:      now,
		RepoCommits: map[string]int{
			"awesome-project": 25,
			"go-utils":        12,
			"dotfiles":        5,
		},
		MainLanguages: map[string]int{
			"Go":         120,
			"TypeScript": 85,
			"Python":     30,
		},
	}
	comparison := github.WeeklyComparison{
		CurrentWeek: &stats,
		PreviousWeek: &github.WeeklyStats{
			TotalCommits: 30,
			ActiveDays:   4,
			StartDate:    now.AddDate(0, 0, -14),
			EndDate:      now.AddDate(0, 0, -7),
			RepoCommits: map[string]int{
				"awesome-project": 18,
				"go-utils":        8,
				"dotfiles":        4,
			},
			MainLanguages: map[string]int{
				"Go":         90,
				"TypeScript": 60,
				"Python":     20,
			},
		},
		CommitsDiff:       12,
		CommitsChangeRate: 40.0,
	}

	tmepl, err := template.ParseFiles("templates/dist/weekly.html")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmepl.Execute(&buf, comparison)
	if err != nil {
		return err
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	emailDomain := os.Getenv("RESEND_EMAIL_DOMAIN")
	emailTo := os.Getenv("TEST_RESEND_EMAIL_TO")

	client := resend.NewClient(apiKey)
	params := &resend.SendEmailRequest{
		From:    "Acme <" + emailDomain + ">",
		To:      []string{emailTo},
		Html:    buf.String(),
		Subject: "週間コミットレポート（テスト送信）",
	}
	fmt.Println(params)
	if err != nil {
		return err
	}
	_, err = client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}
