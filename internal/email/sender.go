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
func LoadTemplate(comparison github.WeeklyComparison) (string, error) {
	tmpl, err := template.ParseFiles("templates/dist/weekly.html")
	if err != nil {
		return "", fmt.Errorf("テンプレート読み込み失敗: %w", err)
	}

	// データをテンプレートに埋め込む
	var buf bytes.Buffer
	// Executeでデータを埋め込む
	if err := tmpl.Execute(&buf, comparison); err != nil {
		return "", fmt.Errorf("テンプレート実行失敗: %w", err)
	}

	return buf.String(), nil
}

// メール送信
func SendWeeklyReport(apiKey string, htmlContent string, imagePath string, emailDomain string, emailTo string) error {
	client := resend.NewClient(apiKey)
	subject := fmt.Sprintf("週間コミットレポート (%s)", time.Now().Format("2006/01/02"))

	params := &resend.SendEmailRequest{
		From:    "お疲れ様委員会 <" + emailDomain + ">",
		To:      []string{emailTo},
		Html:    htmlContent,
		Subject: subject,
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
	startDate := now.AddDate(0, 0, -7)
	prevStartDate := now.AddDate(0, 0, -14)
	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}

	// モックデータ
	stats := github.WeeklyStats{
		TotalCommits: 42,
		ActiveDays:   5,
		StartDate:    startDate,
		EndDate:      now,
		DailyCommits: []github.DailyCommit{
			{Date: startDate, DateStr: startDate.Format("1/2"), Weekday: weekdays[startDate.Weekday()], Count: 5},
			{Date: startDate.AddDate(0, 0, 1), DateStr: startDate.AddDate(0, 0, 1).Format("1/2"), Weekday: weekdays[startDate.AddDate(0, 0, 1).Weekday()], Count: 12},
			{Date: startDate.AddDate(0, 0, 2), DateStr: startDate.AddDate(0, 0, 2).Format("1/2"), Weekday: weekdays[startDate.AddDate(0, 0, 2).Weekday()], Count: 0},
			{Date: startDate.AddDate(0, 0, 3), DateStr: startDate.AddDate(0, 0, 3).Format("1/2"), Weekday: weekdays[startDate.AddDate(0, 0, 3).Weekday()], Count: 8},
			{Date: startDate.AddDate(0, 0, 4), DateStr: startDate.AddDate(0, 0, 4).Format("1/2"), Weekday: weekdays[startDate.AddDate(0, 0, 4).Weekday()], Count: 15},
			{Date: startDate.AddDate(0, 0, 5), DateStr: startDate.AddDate(0, 0, 5).Format("1/2"), Weekday: weekdays[startDate.AddDate(0, 0, 5).Weekday()], Count: 2},
			{Date: startDate.AddDate(0, 0, 6), DateStr: startDate.AddDate(0, 0, 6).Format("1/2"), Weekday: weekdays[startDate.AddDate(0, 0, 6).Weekday()], Count: 0},
		},
		RepoDetails: []github.RepoDetail{
			{Name: "awesome-project", Count: 25, BarPercent: 100.0},
			{Name: "go-utils", Count: 12, BarPercent: 48.0},
			{Name: "dotfiles", Count: 5, BarPercent: 20.0},
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
			StartDate:    prevStartDate,
			EndDate:      now.AddDate(0, 0, -7),
			DailyCommits: []github.DailyCommit{
				{Date: prevStartDate, DateStr: prevStartDate.Format("1/2"), Weekday: weekdays[prevStartDate.Weekday()], Count: 3},
				{Date: prevStartDate.AddDate(0, 0, 1), DateStr: prevStartDate.AddDate(0, 0, 1).Format("1/2"), Weekday: weekdays[prevStartDate.AddDate(0, 0, 1).Weekday()], Count: 8},
				{Date: prevStartDate.AddDate(0, 0, 2), DateStr: prevStartDate.AddDate(0, 0, 2).Format("1/2"), Weekday: weekdays[prevStartDate.AddDate(0, 0, 2).Weekday()], Count: 0},
				{Date: prevStartDate.AddDate(0, 0, 3), DateStr: prevStartDate.AddDate(0, 0, 3).Format("1/2"), Weekday: weekdays[prevStartDate.AddDate(0, 0, 3).Weekday()], Count: 12},
				{Date: prevStartDate.AddDate(0, 0, 4), DateStr: prevStartDate.AddDate(0, 0, 4).Format("1/2"), Weekday: weekdays[prevStartDate.AddDate(0, 0, 4).Weekday()], Count: 7},
				{Date: prevStartDate.AddDate(0, 0, 5), DateStr: prevStartDate.AddDate(0, 0, 5).Format("1/2"), Weekday: weekdays[prevStartDate.AddDate(0, 0, 5).Weekday()], Count: 0},
				{Date: prevStartDate.AddDate(0, 0, 6), DateStr: prevStartDate.AddDate(0, 0, 6).Format("1/2"), Weekday: weekdays[prevStartDate.AddDate(0, 0, 6).Weekday()], Count: 0},
			},
			RepoDetails: []github.RepoDetail{
				{Name: "awesome-project", Count: 18, BarPercent: 100.0},
				{Name: "go-utils", Count: 8, BarPercent: 44.4},
				{Name: "dotfiles", Count: 4, BarPercent: 22.2},
			},
			MainLanguages: map[string]int{
				"Go":         90,
				"TypeScript": 60,
				"Python":     20,
			},
		},
		CommitsDiff:       12,
		CommitsChangeRate: 40,
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
	emailDomain := os.Getenv("RESEND_EMAIL_DOMAIN_DEV")
	emailTo := os.Getenv("TEST_RESEND_EMAIL_TO")

	client := resend.NewClient(apiKey)
	subject := fmt.Sprintf("[テスト]週間コミットレポート (%s)", time.Now().Format("2006/01/02"))
	params := &resend.SendEmailRequest{
		From:    "[テスト]お疲れ様委員会 <" + emailDomain + ">",
		To:      []string{emailTo},
		Html:    buf.String(),
		Subject: subject,
	}
	_, err = client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}
