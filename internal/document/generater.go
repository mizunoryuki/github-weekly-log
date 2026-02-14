package document

import (
	"encoding/json"
	"fmt"
	"github-weekly-log/internal/github"
	"os"
)

// JSONファイルを生成する関数
func GenerateJSONData(data *github.WeeklyComparison) error {
	fileName := fmt.Sprintf("%s.json", data.CurrentWeek.EndDate.Format("2006-01-02"))

	//JSON書き出し
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON変換失敗: %w", err)
	}

	fmt.Println(fileName)

	return os.WriteFile(fileName, file, 0644)
}
