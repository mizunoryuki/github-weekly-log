package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
)

type Client struct {
	ghClient *github.Client
}

// 週間コミットデータ構造体
type WeeklyStats struct {
	TotalCommits   int            // 累計コミット数
	CommitDays     map[string]int // 日付ごとのコミット数
	HourlyActivity [24]int        // 時間帯ごとのコミット数
	RepoCommits    map[string]int // リポジトリごとのコミット数
	StartDate      time.Time      // 週間開始日
	EndDate        time.Time      // 週間終了日
}

// クライアントの生成
func NewClient(token string) *Client {
	return &Client{
		ghClient: github.NewClient(nil).WithAuthToken(token),
	}
}

// データ取得ロジック
func (c *Client) FetchWeeklyCommits(ctx context.Context, username string) (*WeeklyStats, error) {
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)

	stats := &WeeklyStats{
		CommitDays:  make(map[string]int),
		RepoCommits: make(map[string]int),
		StartDate:   oneWeekAgo,
		EndDate:     now,
	}

	// ユーザーのリポジトリ一覧を取得
	opts := &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "pushed",
		Direction:   "desc",
		Visibility:  "all",
		Affiliation: "owner",
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
		if repo.GetPushedAt().Before(oneWeekAgo) {
			continue
		}

		commitOpts := &github.CommitsListOptions{
			Author:      username,
			Since:       oneWeekAgo,
			Until:       now,
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
			if commitDate.Before(oneWeekAgo) || commitDate.After(now) {
				continue
			}

			stats.TotalCommits++

			jst := commitDate.In(time.FixedZone("Asia/Tokyo", 9*60*60))
			dateStr := jst.Format("2006-01-02")
			stats.CommitDays[dateStr]++
			stats.HourlyActivity[jst.Hour()]++

			repoName := repo.GetName()
			stats.RepoCommits[repoName]++
		}
	}

	return stats, nil
}
