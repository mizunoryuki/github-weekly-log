package email

import (
	"bytes"
	"github-weekly-log/internal/github"
	"text/template"
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
