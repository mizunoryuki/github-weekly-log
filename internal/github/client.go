package github

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
)

type Client struct {
	ghClient *github.Client
}

// 週間コミットデータ構造体
type WeeklyStats struct {
	TotalCommits    int            // 累計コミット数
	CommitDays      map[string]int // 日付ごとのコミット数
	HourlyActivity  [24]int        // 時間帯ごとのコミット数
	RepoCommits     map[string]int // リポジトリごとのコミット数
	LanguageCommits map[string]int // 言語ごとのコミット数
	StartDate       time.Time      // 週間開始日
	EndDate         time.Time      // 週間終了日
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

	stats := &WeeklyStats{
		CommitDays:      make(map[string]int),
		RepoCommits:     make(map[string]int),
		LanguageCommits: make(map[string]int),
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
		if repo.GetPushedAt().Before(startDate) {
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
