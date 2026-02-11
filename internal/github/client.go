package github

import "github.com/google/go-github/v60/github"

type Client struct {
	ghClient *github.Client
}

// クライアントの生成
func NewClient(token string) *Client {
	return &Client{
		ghClient: github.NewClient(nil).WithAuthToken(token),
	}
}

// データ取得ロジック
func (c *Client) FetchWeeklyEvents(username string) ([]*github.Event, error) {
	return nil, nil
}
