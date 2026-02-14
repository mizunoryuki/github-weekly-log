package database

import (
	"context"
	"fmt"
	"github-weekly-log/internal/github"
	"log"
	"strconv"
	"time"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

func InitD1(apiToken string, accountID string) *cloudflare.Client {
	cfClient := cloudflare.NewClient(
		option.WithAPIToken(apiToken),
	)

	return cfClient
}

// D1に週間コミットデータを保存する関数
func SaveWeeklyStatsToD1WithTransaction(ctx context.Context, client *cloudflare.Client, accountID, databaseID string, stats *github.WeeklyStats) error {
	log.Println("[INFO] データ保存処理を開始します")

	// weekly_stats を挿入（UPSERT）
	log.Println("[INFO] weekly_stats テーブルへの挿入を開始")
	weeklyStatsID, err := insertWeeklyStats(ctx, client, accountID, databaseID, stats)
	if err != nil {
		log.Printf("[ERROR] weekly_stats の挿入に失敗しました: %v", err)
		return fmt.Errorf("weekly_stats挿入エラー: %w", err)
	}
	log.Printf("[INFO] weekly_stats を挿入しました (ID: %s)", weeklyStatsID)

	// 既存の子データを削除
	log.Println("[INFO] 既存の子データを削除中")
	err = deleteChildData(ctx, client, accountID, databaseID, weeklyStatsID)
	if err != nil {
		log.Printf("[WARN] 既存データの削除に失敗しました: %v", err)
	}

	// 子データを一括挿入
	log.Println("[INFO] 子データの挿入を開始")
	err = insertChildData(ctx, client, accountID, databaseID, weeklyStatsID, stats)
	if err != nil {
		log.Printf("[ERROR] 子データの挿入に失敗しました: %v", err)
		return fmt.Errorf("子データ挿入エラー: %w", err)
	}

	log.Printf("[INFO] データ保存が完了しました (ID: %s, commits: %d, active_days: %d)",
		weeklyStatsID, stats.TotalCommits, stats.ActiveDays)
	return nil
}

// 既存の子データを削除
func deleteChildData(ctx context.Context, client *cloudflare.Client, accountID, databaseID, weeklyStatsID string) error {
	var batch []d1.DatabaseQueryParamsBodyMultipleQueriesBatch

	batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
		Sql:    cloudflare.F(`DELETE FROM language_commits WHERE weekly_stats_id = ?`),
		Params: cloudflare.F([]string{weeklyStatsID}),
	})
	batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
		Sql:    cloudflare.F(`DELETE FROM repo_details WHERE weekly_stats_id = ?`),
		Params: cloudflare.F([]string{weeklyStatsID}),
	})
	batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
		Sql:    cloudflare.F(`DELETE FROM hourly_activity WHERE weekly_stats_id = ?`),
		Params: cloudflare.F([]string{weeklyStatsID}),
	})
	batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
		Sql:    cloudflare.F(`DELETE FROM daily_commits WHERE weekly_stats_id = ?`),
		Params: cloudflare.F([]string{weeklyStatsID}),
	})

	result, err := client.D1.Database.Query(ctx, databaseID, d1.DatabaseQueryParams{
		AccountID: cloudflare.F(accountID),
		Body:      d1.DatabaseQueryParamsBodyMultipleQueries{Batch: cloudflare.F(batch)},
	})
	if err != nil {
		return err
	}

	totalDeleted := 0
	for _, queryResult := range result.Result {
		if queryResult.Meta.Changes > 0 {
			totalDeleted += int(queryResult.Meta.Changes)
		}
	}

	if totalDeleted > 0 {
		log.Printf("[INFO] 既存の子データを削除しました (%d件)", totalDeleted)
	}

	return nil
}

// weekly_stats を挿入して ID を返す
func insertWeeklyStats(ctx context.Context, client *cloudflare.Client, accountID, databaseID string, stats *github.WeeklyStats) (string, error) {
	result, err := client.D1.Database.Query(ctx, databaseID, d1.DatabaseQueryParams{
		AccountID: cloudflare.F(accountID),
		Body: d1.DatabaseQueryParamsBodyD1SingleQuery{
			Sql: cloudflare.F(`
				INSERT INTO weekly_stats (start_date, end_date, total_commits, active_days, created_at)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(start_date) DO UPDATE SET
					end_date = excluded.end_date,
					total_commits = excluded.total_commits,
					active_days = excluded.active_days,
					created_at = excluded.created_at
				RETURNING id
			`),
			Params: cloudflare.F([]string{
				stats.StartDate.Format("2006-01-02"),
				stats.EndDate.Format("2006-01-02"),
				strconv.Itoa(stats.TotalCommits),
				strconv.Itoa(stats.ActiveDays),
				time.Now().Format(time.RFC3339),
			}),
		},
	})
	if err != nil {
		return "", err
	}

	if len(result.Result) == 0 {
		return "", fmt.Errorf("weekly_stats の挿入結果が空です")
	}

	queryResult := result.Result[0]
	if len(queryResult.Results) == 0 {
		return "", fmt.Errorf("RETURNING id の結果が空です")
	}

	firstRowInterface := queryResult.Results[0]
	firstRow, ok := firstRowInterface.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("結果の型変換に失敗しました: %T", firstRowInterface)
	}

	if id, exists := firstRow["id"]; exists {
		switch v := id.(type) {
		case float64:
			return strconv.Itoa(int(v)), nil
		case int64:
			return strconv.FormatInt(v, 10), nil
		case int:
			return strconv.Itoa(v), nil
		case string:
			return v, nil
		default:
			return "", fmt.Errorf("予期しない id の型: %T", id)
		}
	}

	return "", fmt.Errorf("id フィールドが見つかりません")
}

// 子データを一括挿入
func insertChildData(ctx context.Context, client *cloudflare.Client, accountID, databaseID, weeklyStatsID string, stats *github.WeeklyStats) error {
	var batch []d1.DatabaseQueryParamsBodyMultipleQueriesBatch

	// daily_commits
	for _, daily := range stats.DailyCommits {
		batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
			Sql: cloudflare.F(`INSERT INTO daily_commits (weekly_stats_id, date, commits) VALUES (?, ?, ?)`),
			Params: cloudflare.F([]string{
				weeklyStatsID,
				daily.Date.Format("2006-01-02"),
				strconv.Itoa(daily.Count),
			}),
		})
	}

	// hourly_activity
	hourlyCount := 0
	for hour, commits := range stats.HourlyActivity {
		if commits > 0 {
			hourlyCount++
			batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
				Sql: cloudflare.F(`INSERT INTO hourly_activity (weekly_stats_id, hour, commits) VALUES (?, ?, ?)`),
				Params: cloudflare.F([]string{
					weeklyStatsID,
					strconv.Itoa(hour),
					strconv.Itoa(commits),
				}),
			})
		}
	}

	// repo_details
	for _, repo := range stats.RepoDetails {
		batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
			Sql: cloudflare.F(`INSERT INTO repo_details (weekly_stats_id, repo_name, commits, bar_width) VALUES (?, ?, ?, ?)`),
			Params: cloudflare.F([]string{
				weeklyStatsID,
				repo.Name,
				strconv.Itoa(repo.Count),
				strconv.Itoa(int(repo.BarPercent)),
			}),
		})
	}

	// language_commits
	for lang, commits := range stats.LanguageCommits {
		isMain := "0"
		if _, exists := stats.MainLanguages[lang]; exists {
			isMain = "1"
		}
		batch = append(batch, d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
			Sql: cloudflare.F(`INSERT INTO language_commits (weekly_stats_id, language, commits, is_main) VALUES (?, ?, ?, ?)`),
			Params: cloudflare.F([]string{
				weeklyStatsID,
				lang,
				strconv.Itoa(commits),
				isMain,
			}),
		})
	}

	if len(batch) == 0 {
		log.Println("[WARN] 挿入する子データがありません")
		return nil
	}

	log.Printf("[INFO] バッチ処理を実行します (daily: %d, hourly: %d, repos: %d, langs: %d, total: %d)",
		len(stats.DailyCommits), hourlyCount, len(stats.RepoDetails), len(stats.LanguageCommits), len(batch))

	result, err := client.D1.Database.Query(ctx, databaseID, d1.DatabaseQueryParams{
		AccountID: cloudflare.F(accountID),
		Body:      d1.DatabaseQueryParamsBodyMultipleQueries{Batch: cloudflare.F(batch)},
	})
	if err != nil {
		return err
	}

	// 結果確認
	for i, queryResult := range result.Result {
		if !queryResult.Success {
			log.Printf("[ERROR] バッチ処理 #%d が失敗しました", i)
			return fmt.Errorf("バッチ #%d 実行エラー", i)
		}
	}

	log.Printf("[INFO] 子データの挿入が完了しました (%d件)", len(result.Result))
	return nil
}
