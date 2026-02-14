package database

import (
	"context"
	"fmt"
	"github-weekly-log/internal/github"
	"strconv"
	"time"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zones"
)

func InitD1(apiToken string, accountID string) *cloudflare.Client {
	cfClient := cloudflare.NewClient(
		option.WithAPIToken(apiToken),
	)

	zone, err := cfClient.Zones.New(context.TODO(), zones.ZoneNewParams{
		Account: cloudflare.F(zones.ZoneNewParamsAccount{
			ID: cloudflare.F(accountID),
		}),
	})
	if err != nil {
		panic(err)
	}
	println("Zone created with ID:", zone.ID)

	return cfClient
}

// D1に週間コミットデータを保存する関数
func SaveWeeklyStatsToD1WithTransaction(ctx context.Context, client *cloudflare.Client, accountID, databaseID string, stats *github.WeeklyStats) error {
	var batch []d1.DatabaseQueryParamsBodyMultipleQueriesBatch

	// weekly_stats
	batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
		Sql: cloudflare.F(`
			INSERT INTO weekly_stats (start_date, end_date, total_commits, active_days, created_at)
			VALUES (?, ?, ?, ?, ?)
		`),
		Params: cloudflare.F([]string{
			stats.StartDate.Format("2006-01-02"),
			stats.EndDate.Format("2006-01-02"),
			strconv.Itoa(stats.TotalCommits),
			strconv.Itoa(stats.ActiveDays),
			time.Now().Format(time.RFC3339),
		}),
	})

	// daily_commits
	for _, daily := range stats.DailyCommits {
		batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
			Sql:    cloudflare.F(`INSERT INTO daily_commits (weekly_stats_id, date, commits) VALUES (last_insert_rowid(), ?, ?)`),
			Params: cloudflare.F([]string{daily.Date.Format("2006-01-02"), strconv.Itoa(daily.Count)}),
		})
	}

	// hourly_activity
	for hour, commits := range stats.HourlyActivity {
		batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
			Sql:    cloudflare.F(`INSERT INTO hourly_activity (weekly_stats_id, hour, commits) VALUES (last_insert_rowid(), ?, ?)`),
			Params: cloudflare.F([]string{strconv.Itoa(hour), strconv.Itoa(commits)}),
		})
	}

	// repo_details
	for _, repo := range stats.RepoDetails {
		batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
			Sql:    cloudflare.F(`INSERT INTO repo_details (weekly_stats_id, repo_name, commits, bar_width) VALUES (last_insert_rowid(), ?, ?, ?)`),
			Params: cloudflare.F([]string{repo.Name, strconv.Itoa(repo.Count), strconv.Itoa(int(repo.BarPercent))}),
		})
	}

	// language_commits
	for lang, commits := range stats.LanguageCommits {
		isMain := "0"
		if _, exists := stats.MainLanguages[lang]; exists {
			isMain = "1"
		}
		batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
			Sql:    cloudflare.F(`INSERT INTO language_commits (weekly_stats_id, language, commits, is_main) VALUES (last_insert_rowid(), ?, ?, ?)`),
			Params: cloudflare.F([]string{lang, strconv.Itoa(commits), isMain}),
		})
	}

	result, err := client.D1.Database.Query(ctx, databaseID, d1.DatabaseQueryParams{
		AccountID: cloudflare.F(accountID),
		Body:      d1.DatabaseQueryParamsBodyMultipleQueries{Batch: cloudflare.F(batch)},
	})
	if err != nil {
		return fmt.Errorf("トランザクション実行エラー: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, item := range result.Result {
		if item.Success {
			successCount++
		} else {
			errorCount++
		}
	}

	fmt.Printf("✅ %d個のSQL文を実行しました (成功: %d, 失敗: %d)\n", len(batch), successCount, errorCount)
	return nil
}
