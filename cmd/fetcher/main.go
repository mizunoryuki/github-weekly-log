package main

import (
	"context"
	"fmt"
	"github-weekly-log/internal/database"
	"github-weekly-log/internal/email"
	"github-weekly-log/internal/github"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	if len(os.Args) > 1 && os.Args[1] == "test-email" {
		err := email.TestWeeklyMailSend()
		if err != nil {
			panic(err)
		}
		return
	}

	GITHUB_TOKEN := os.Getenv("GITHUB_TOKEN")
	GITHUB_USER := os.Getenv("GITHUB_USER")
	EMAIL_API_KEY := os.Getenv("RESEND_API_KEY")
	EMAIL_DOMAIN := os.Getenv("RESEND_EMAIL_DOMAIN")
	EMAIL_TO := os.Getenv("RESEND_EMAIL_TO")
	D1_API_TOKEN := os.Getenv("D1_API_TOKEN")
	D1_DATABASE_ID := os.Getenv("D1_DATABASE_ID")
	D1_ACCOUNT_ID := os.Getenv("D1_ACCOUNT_ID")

	client := github.NewClient(GITHUB_TOKEN)

	fmt.Println("Start scanning")
	comparison, err := client.FetchWeeklyCommitsWithComparison(context.Background(), GITHUB_USER)
	if err != nil {
		panic(err)
	}
	fmt.Println("Finished scanning")

	// 結果表示
	printWeeklyComparison(comparison)

	// D1に保存
	cfClient := database.InitD1(D1_API_TOKEN, D1_ACCOUNT_ID)

	fmt.Println("Save to D1")
	err = database.SaveWeeklyStatsToD1WithTransaction(context.Background(), cfClient, D1_ACCOUNT_ID, D1_DATABASE_ID, comparison.CurrentWeek)
	if err != nil {
		panic(err)
	}

	fmt.Println("Load HTML template")
	// HTMLテンプレート読み込み
	htmlContent, err := email.LoadTemplate(*comparison)
	if err != nil {
		panic(err)
	}

	fmt.Println("Send weekly report email")
	// メール送信
	err = email.SendWeeklyReport(EMAIL_API_KEY, htmlContent, "", EMAIL_DOMAIN, EMAIL_TO)
	if err != nil {
		panic(err)
	}

}

func printWeeklyComparison(comp *github.WeeklyComparison) {
	current := comp.CurrentWeek
	previous := comp.PreviousWeek

	fmt.Println("========================================")
	fmt.Println("週間コミットレポート（前週比）")
	fmt.Println("========================================")

	// 今週の期間
	fmt.Printf("\n📅 今週: %s 〜 %s\n",
		current.StartDate.Format("2006-01-02"),
		current.EndDate.Format("2006-01-02"))
	fmt.Printf("📅 先週: %s 〜 %s\n",
		previous.StartDate.Format("2006-01-02"),
		previous.EndDate.Format("2006-01-02"))

	// コミット数比較
	fmt.Println("\n📊 総コミット数:")
	fmt.Printf("  今週: %d\n", current.TotalCommits)
	fmt.Printf("  先週: %d\n", previous.TotalCommits)

	// 差分と変化率を表示
	if comp.CommitsDiff > 0 {
		fmt.Printf("  📈 %+d (%d%% 増加)\n", comp.CommitsDiff, comp.CommitsChangeRate)
	} else if comp.CommitsDiff < 0 {
		fmt.Printf("  📉 %d (%d%% 減少)\n", comp.CommitsDiff, -comp.CommitsChangeRate)
	} else {
		fmt.Printf("  ➡️  変化なし\n")
	}

	// リポジトリ別比較
	fmt.Println("\n📁 リポジトリ別コミット数:")
	fmt.Println("  リポジトリ名          今週  先週  差分")
	fmt.Println("  " + strings.Repeat("-", 45))

	// RepoDetails から map を生成
	currentRepos := make(map[string]int)
	for _, repo := range current.RepoDetails {
		currentRepos[repo.Name] = repo.Count
	}

	previousRepos := make(map[string]int)
	for _, repo := range previous.RepoDetails {
		previousRepos[repo.Name] = repo.Count
	}

	allRepos := make(map[string]bool)
	for repo := range currentRepos {
		allRepos[repo] = true
	}
	for repo := range previousRepos {
		allRepos[repo] = true
	}

	for repo := range allRepos {
		currentCount := currentRepos[repo]
		previousCount := previousRepos[repo]
		diff := currentCount - previousCount

		diffStr := ""
		if diff > 0 {
			diffStr = fmt.Sprintf("+%d", diff)
		} else if diff < 0 {
			diffStr = fmt.Sprintf("%d", diff)
		} else {
			diffStr = "0"
		}

		fmt.Printf("  %-20s %4d  %4d  %s\n", repo, currentCount, previousCount, diffStr)
	}

	// 言語別比較
	fmt.Println("\n💻 言語別変更ファイル数:")
	fmt.Println("  言語                  今週  先週  差分")
	fmt.Println("  " + strings.Repeat("-", 45))

	allLangs := make(map[string]bool)
	for lang := range current.LanguageCommits {
		allLangs[lang] = true
	}
	for lang := range previous.LanguageCommits {
		allLangs[lang] = true
	}

	for lang := range allLangs {
		currentCount := current.LanguageCommits[lang]
		previousCount := previous.LanguageCommits[lang]
		diff := currentCount - previousCount

		diffStr := ""
		if diff > 0 {
			diffStr = fmt.Sprintf("+%d", diff)
		} else if diff < 0 {
			diffStr = fmt.Sprintf("%d", diff)
		} else {
			diffStr = "0"
		}

		fmt.Printf("  %-20s %4d  %4d  %s\n", lang, currentCount, previousCount, diffStr)
	}
}
