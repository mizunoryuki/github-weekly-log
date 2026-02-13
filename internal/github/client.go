package github

import (
	"context"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
)

type Client struct {
	ghClient *github.Client
}

// 日次コミットデータ
type DailyCommit struct {
	Date    time.Time // 日付
	DateStr string    // "1/2" 形式の日付文字列
	Weekday string    // "月", "火" など
	Count   int       // コミット数
}

// 週間コミットデータ構造体
type WeeklyStats struct {
	TotalCommits    int            // 累計コミット数
	CommitDays      map[string]int // 日付ごとのコミット数
	DailyCommits    []DailyCommit  // 7日分の日次データ（順序保証）
	HourlyActivity  [24]int        // 時間帯ごとのコミット数
	RepoCommits     map[string]int // リポジトリごとのコミット数
	LanguageCommits map[string]int // 言語ごとのコミット数
	MainLanguages   map[string]int // 主要言語ごとのコミット数
	StartDate       time.Time      // 週間開始日
	EndDate         time.Time      // 週間終了日
	ActiveDays      int            // コミットがあった日数
}

// 前週比較データ構造体
type WeeklyComparison struct {
	CurrentWeek       *WeeklyStats
	PreviousWeek      *WeeklyStats
	CommitsDiff       int     // コミット数の差分
	CommitsChangeRate float64 // コミット数の変化率（%）
}

// クライアントの生成
func NewClient(token string) *Client {
	return &Client{
		ghClient: github.NewClient(nil).WithAuthToken(token),
	}
}

// データ取得ロジック
func (c *Client) FetchWeeklyCommits(ctx context.Context, username string) (*WeeklyStats, error) {
	// 週間の開始日と終了日を取得
	startDate, endDate := getTargetRange()
	return c.fetchWeeklyCommitsInRange(ctx, username, startDate, endDate)
}

// 前週比を含むデータ取得
func (c *Client) FetchWeeklyCommitsWithComparison(ctx context.Context, username string) (*WeeklyComparison, error) {
	// 今週のデータを取得
	currentStart, currentEnd := getTargetRange()
	currentWeek, err := c.fetchWeeklyCommitsInRange(ctx, username, currentStart, currentEnd)
	if err != nil {
		return nil, fmt.Errorf("error fetching current week data: %v", err)
	}

	// 先週のデータを取得
	previousStart := currentStart.AddDate(0, 0, -7)
	previousEnd := currentEnd.AddDate(0, 0, -7)
	previousWeek, err := c.fetchWeeklyCommitsInRange(ctx, username, previousStart, previousEnd)
	if err != nil {
		return nil, fmt.Errorf("error fetching previous week data: %v", err)
	}

	// 比較データを計算
	comparison := &WeeklyComparison{
		CurrentWeek:  currentWeek,
		PreviousWeek: previousWeek,
		CommitsDiff:  currentWeek.TotalCommits - previousWeek.TotalCommits,
	}

	// 変化率を計算
	if previousWeek.TotalCommits > 0 {
		comparison.CommitsChangeRate = float64(comparison.CommitsDiff) / float64(previousWeek.TotalCommits) * 100
	} else if currentWeek.TotalCommits > 0 {
		comparison.CommitsChangeRate = 100 // 0から増加した場合は100%とする
	}

	return comparison, nil
}

// 指定期間のコミットデータを取得（内部用）
func (c *Client) fetchWeeklyCommitsInRange(ctx context.Context, username string, startDate, endDate time.Time) (*WeeklyStats, error) {
	stats := &WeeklyStats{
		CommitDays:      make(map[string]int),
		RepoCommits:     make(map[string]int),
		LanguageCommits: make(map[string]int),
		MainLanguages:   make(map[string]int),
		StartDate:       startDate,
		EndDate:         endDate,
	}

	// ユーザーのリポジトリ一覧を取得
	opts := &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "pushed",
		Direction:   "desc",
		Visibility:  "all",
		Affiliation: "owner", // 所有しているリポジトリのみ
	}

	var allRepos []*github.Repository

	// 全リポジトリを取得
	for {
		repos, resp, err := c.ghClient.Repositories.ListByAuthenticatedUser(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("error fetching repositories: %v", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// 各リポジトリのコミットを取得
	for _, repo := range allRepos {
		// 最近プッシュされていないリポジトリはスキップ
		// 先週のデータも取得するため、2週間前までチェック
		twoWeeksAgo := startDate.AddDate(0, 0, -7)
		if repo.GetPushedAt().Before(twoWeeksAgo) {
			continue
		}

		commitOpts := &github.CommitsListOptions{
			Author:      username,
			Since:       startDate,
			Until:       endDate,
			ListOptions: github.ListOptions{PerPage: 100},
		}

		// 全コミットを取得
		commits, _, err := c.ghClient.Repositories.ListCommits(
			ctx, username, repo.GetName(), commitOpts)
		if err != nil {
			fmt.Printf("Error fetching commits for %s: %v\n", repo.GetName(), err)
			continue
		}

		// 期間内のコミットを集計
		for _, commit := range commits {
			if commit.Commit == nil || commit.Commit.Author == nil {
				continue
			}

			commitDate := commit.Commit.Author.GetDate()
			if commitDate.Before(startDate) || commitDate.After(endDate) {
				continue
			}

			stats.TotalCommits++

			jst := commitDate.In(time.FixedZone("Asia/Tokyo", 9*60*60))
			dateStr := jst.Format("2006-01-02")
			stats.CommitDays[dateStr]++
			stats.HourlyActivity[jst.Hour()]++

			repoName := repo.GetName()
			stats.RepoCommits[repoName]++

			// コミットの言語集計
			// ファイル情報を取得
			commitDetail, _, err := c.ghClient.Repositories.GetCommit(
				ctx, username, repoName, commit.GetSHA(), nil)
			if err != nil {
				fmt.Printf("Error fetching commit details for %s: %v\n", repoName, err)
				continue
			}
			// 変更されたファイルごとに言語を集計
			for _, file := range commitDetail.Files {
				filename := file.GetFilename()
				language := getLanguageFromFilename(filename)
				if language != "" {
					stats.LanguageCommits[language]++
				}
			}
		}
	}

	// CommitDaysを日付の昇順にソート
	dateKeys := slices.Sorted(maps.Keys(stats.CommitDays))
	sortedCommitDays := make(map[string]int)
	for _, date := range dateKeys {
		sortedCommitDays[date] = stats.CommitDays[date]
	}
	stats.CommitDays = sortedCommitDays
	stats.ActiveDays = len(stats.CommitDays)
	// 回数0の日付を補完
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if _, exists := stats.CommitDays[dateStr]; !exists {
			stats.CommitDays[dateStr] = 0
		}
	}

	// 主要言語のみフィルタリング
	stats.MainLanguages = filterMainLanguages(stats.LanguageCommits)

	// 7日分のDailyCommitsを生成（コントリビュートグラフ用）
	stats.DailyCommits = generateDailyCommits(startDate, endDate, stats.CommitDays)

	return stats, nil
}

// 週間の開始日と終了日を取得
func getTargetRange() (time.Time, time.Time) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().In(jst)

	// 今週の金曜日21時を終了日時とする
	// 現在の曜日から金曜日までの日数を計算
	offset := int(time.Friday - now.Weekday())

	endDate := time.Date(now.Year(), now.Month(), now.Day()+offset, 21, 0, 0, 0, jst)
	startDate := endDate.AddDate(0, 0, -7)
	return startDate, endDate
}

func getLanguageFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// 拡張子を確認する
	if lang, exists := LANGUAGE_MAP[ext]; exists {
		return lang
	}

	baseName := filepath.Base(filename)
	// 特殊なファイル名を確認する
	if lang, exists := SPECIAL_LANGUAGE_MAP[baseName]; exists {
		return lang
	}

	return "Other"
}

// 主要言語をフィルタリング
func filterMainLanguages(langMap map[string]int) map[string]int {
	filtered := make(map[string]int)
	for lang, count := range langMap {
		if MAIN_LANGUAGES_SET[lang] {
			filtered[lang] = count
		}
	}
	return filtered
}

// 7日分のDailyCommitsを生成（コントリビュートグラフ用）
func generateDailyCommits(startDate, endDate time.Time, commitDays map[string]int) []DailyCommit {
	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}
	dailyCommits := make([]DailyCommit, 0, 7)

	// 7日分のデータを生成
	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		dateKey := date.Format("2006-01-02")
		count := commitDays[dateKey] // コミットがない日は0

		dailyCommits = append(dailyCommits, DailyCommit{
			Date:    date,
			DateStr: date.Format("1/2"),
			Weekday: weekdays[date.Weekday()],
			Count:   count,
		})
	}

	return dailyCommits
}
