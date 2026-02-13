package github

import (
	"math"
	"testing"
	"time"
)

// テスト: 変化率計算（正常系・エッジケース）
func TestCommitsChangeRateCalculation(t *testing.T) {
	tests := []struct {
		name                string
		currentWeekCommits  int
		previousWeekCommits int
		expectedDiff        int
		expectedChangeRate  int
	}{
		{
			name:                "通常の増加（前週30 → 今週42）",
			currentWeekCommits:  42,
			previousWeekCommits: 30,
			expectedDiff:        12,
			expectedChangeRate:  40, // (12 / 30) * 100 = 40%
		},
		{
			name:                "通常の減少（前週100 → 今週50）",
			currentWeekCommits:  50,
			previousWeekCommits: 100,
			expectedDiff:        -50,
			expectedChangeRate:  -50, // (-50 / 100) * 100 = -50%
		},
		{
			name:                "前週0コミット → 今週10（100%増）",
			currentWeekCommits:  10,
			previousWeekCommits: 0,
			expectedDiff:        10,
			expectedChangeRate:  100, // 0から増加した場合は100%
		},
		{
			name:                "両週0コミット（変化なし）",
			currentWeekCommits:  0,
			previousWeekCommits: 0,
			expectedDiff:        0,
			expectedChangeRate:  0, // 変化率も0
		},
		{
			name:                "前週多い、今週0（-100%）",
			currentWeekCommits:  0,
			previousWeekCommits: 50,
			expectedDiff:        -50,
			expectedChangeRate:  -100, // (-50 / 50) * 100 = -100%
		},
		{
			name:                "四捨五入テスト（33.3% → 33%）",
			currentWeekCommits:  40,
			previousWeekCommits: 120,
			expectedDiff:        -80,
			expectedChangeRate:  -67, // (-80 / 120) * 100 = -66.666... → -67%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モック WeeklyComparison を作成
			comparison := &WeeklyComparison{
				CurrentWeek: &WeeklyStats{
					TotalCommits: tt.currentWeekCommits,
				},
				PreviousWeek: &WeeklyStats{
					TotalCommits: tt.previousWeekCommits,
				},
				CommitsDiff: tt.currentWeekCommits - tt.previousWeekCommits,
			}

			// 変化率を計算
			if tt.previousWeekCommits > 0 {
				rate := float64(comparison.CommitsDiff) / float64(tt.previousWeekCommits) * 100
				comparison.CommitsChangeRate = int(math.Round(rate))
			} else if tt.currentWeekCommits > 0 {
				comparison.CommitsChangeRate = 100
			}

			// 検証
			if comparison.CommitsDiff != tt.expectedDiff {
				t.Errorf("CommitsDiff: expected %d, got %d", tt.expectedDiff, comparison.CommitsDiff)
			}
			if comparison.CommitsChangeRate != tt.expectedChangeRate {
				t.Errorf("CommitsChangeRate: expected %d, got %d", tt.expectedChangeRate, comparison.CommitsChangeRate)
			}
		})
	}
}

// テスト: 7日分データ生成
func TestGenerateDailyCommits(t *testing.T) {
	// テスト用の開始日付（2026年2月7日 土曜日）
	startDate := time.Date(2026, 2, 7, 0, 0, 0, 0, time.UTC)

	// テスト用のcommitDayデータ
	commitDays := map[string]int{
		"2026-02-07": 5,
		"2026-02-08": 12,
		"2026-02-09": 0,
		"2026-02-10": 8,
		"2026-02-11": 15,
		"2026-02-12": 2,
		"2026-02-13": 0,
	}

	dailyCommits := generateDailyCommits(startDate, commitDays)

	// 検証 1: 7日分のデータが生成されているか
	if len(dailyCommits) != 7 {
		t.Errorf("Expected 7 days, got %d", len(dailyCommits))
	}

	// 検証 2: 日付が正しく計算されているか
	expectedDates := []string{
		"2/7", "2/8", "2/9", "2/10", "2/11", "2/12", "2/13",
	}
	for i, expected := range expectedDates {
		if dailyCommits[i].DateStr != expected {
			t.Errorf("Day %d: expected date %s, got %s", i, expected, dailyCommits[i].DateStr)
		}
	}

	// 検証 3: コミット数が正しく割り当てられているか
	expectedCounts := []int{5, 12, 0, 8, 15, 2, 0}
	for i, expected := range expectedCounts {
		if dailyCommits[i].Count != expected {
			t.Errorf("Day %d count: expected %d, got %d", i, expected, dailyCommits[i].Count)
		}
	}

	// 検証 4: 曜日が正しいか（2026-02-07は土曜日から）
	expectedWeekdays := []string{"土", "日", "月", "火", "水", "木", "金"}
	for i, expected := range expectedWeekdays {
		if dailyCommits[i].Weekday != expected {
			t.Errorf("Day %d weekday: expected %s, got %s", i, expected, dailyCommits[i].Weekday)
		}
	}
}

// テスト: リポジトリバー幅計算
func TestGenerateRepoDetails(t *testing.T) {
	tests := []struct {
		name              string
		repoCommits       map[string]int
		expectedMaxCount  int
		expectedRepoCount int
	}{
		{
			name: "通常のリポジトリデータ",
			repoCommits: map[string]int{
				"awesome-project": 25,
				"go-utils":        12,
				"dotfiles":        5,
			},
			expectedMaxCount:  25,
			expectedRepoCount: 3,
		},
		{
			name: "単一リポジトリ",
			repoCommits: map[string]int{
				"single-repo": 100,
			},
			expectedMaxCount:  100,
			expectedRepoCount: 1,
		},
		{
			name:              "空のリポジトリ",
			repoCommits:       map[string]int{},
			expectedMaxCount:  0,
			expectedRepoCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details := generateRepoDetails(tt.repoCommits)

			// リポジトリ数の検証
			if len(details) != tt.expectedRepoCount {
				t.Errorf("Repository count: expected %d, got %d", tt.expectedRepoCount, len(details))
			}

			// 空のリポジトリはスキップ
			if len(details) == 0 {
				return
			}

			// 最大値の検証
			if details[0].Count != tt.expectedMaxCount {
				t.Errorf("Max repo commits: expected %d, got %d", tt.expectedMaxCount, details[0].Count)
			}

			// バー幅パーセンテージの検証
			if details[0].BarPercent != 100.0 {
				t.Errorf("First repo bar percent: expected 100.0, got %f", details[0].BarPercent)
			}

			// コミット数の多い順でソートされているか確認
			for i := 1; i < len(details); i++ {
				if details[i].Count > details[i-1].Count {
					t.Errorf("Not sorted by count descending: [%d]=%d > [%d]=%d",
						i, details[i].Count, i-1, details[i-1].Count)
				}
			}

			// バー幅が0-100の範囲か
			for i, detail := range details {
				if detail.BarPercent < 0 || detail.BarPercent > 100 {
					t.Errorf("Repo %d (%s) bar percent out of range: %f", i, detail.Name, detail.BarPercent)
				}
			}
		})
	}
}

// テスト: ActiveDays計算
func TestActiveDaysCalculation(t *testing.T) {
	tests := []struct {
		name         string
		dailyCommits []DailyCommit
		expectedDays int
	}{
		{
			name: "5日間活動",
			dailyCommits: []DailyCommit{
				{Count: 5}, {Count: 12}, {Count: 0},
				{Count: 8}, {Count: 15}, {Count: 2}, {Count: 0},
			},
			expectedDays: 5,
		},
		{
			name: "全日活動",
			dailyCommits: []DailyCommit{
				{Count: 1}, {Count: 2}, {Count: 3},
				{Count: 4}, {Count: 5}, {Count: 6}, {Count: 7},
			},
			expectedDays: 7,
		},
		{
			name: "全日非活動",
			dailyCommits: []DailyCommit{
				{Count: 0}, {Count: 0}, {Count: 0},
				{Count: 0}, {Count: 0}, {Count: 0}, {Count: 0},
			},
			expectedDays: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activeDays := 0
			for _, day := range tt.dailyCommits {
				if day.Count > 0 {
					activeDays++
				}
			}

			if activeDays != tt.expectedDays {
				t.Errorf("ActiveDays: expected %d, got %d", tt.expectedDays, activeDays)
			}
		})
	}
}

// テスト: ターゲット日付範囲計算
func TestGetTargetRange(t *testing.T) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)

	tests := []struct {
		name                 string
		testTime             time.Time // 実行時刻（JST）
		expectedStartDateStr string    // "2026-02-07" 形式
		expectedEndDateStr   string
	}{
		{
			name:                 "2月7日（土曜日）に実行",
			testTime:             time.Date(2026, 2, 7, 10, 0, 0, 0, jst),
			expectedStartDateStr: "2026-01-31",
			expectedEndDateStr:   "2026-02-06",
		},
		{
			name:                 "2月8日（日曜日）に実行",
			testTime:             time.Date(2026, 2, 8, 10, 0, 0, 0, jst),
			expectedStartDateStr: "2026-02-07",
			expectedEndDateStr:   "2026-02-13",
		},
		{
			name:                 "2月14日（土曜日）に実行",
			testTime:             time.Date(2026, 2, 14, 10, 0, 0, 0, jst),
			expectedStartDateStr: "2026-02-07",
			expectedEndDateStr:   "2026-02-13",
		},
		{
			name:                 "2月15日（日曜日）に実行",
			testTime:             time.Date(2026, 2, 15, 10, 0, 0, 0, jst),
			expectedStartDateStr: "2026-02-14",
			expectedEndDateStr:   "2026-02-20",
		},
		{
			name:                 "2月20日（金曜日）に実行",
			testTime:             time.Date(2026, 2, 20, 10, 0, 0, 0, jst),
			expectedStartDateStr: "2026-02-14",
			expectedEndDateStr:   "2026-02-20",
		},
		{
			name:                 "2月21日（土曜日）に実行",
			testTime:             time.Date(2026, 2, 21, 10, 0, 0, 0, jst),
			expectedStartDateStr: "2026-02-14",
			expectedEndDateStr:   "2026-02-20",
		},
		{
			name:                 "2月22日（日曜日）に実行",
			testTime:             time.Date(2026, 2, 22, 10, 0, 0, 0, jst),
			expectedStartDateStr: "2026-02-21",
			expectedEndDateStr:   "2026-02-27",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := getTargetRangeAt(tt.testTime)

			startStr := start.Format("2006-01-02")
			endStr := end.Format("2006-01-02")

			if startStr != tt.expectedStartDateStr {
				t.Errorf("Start date: expected %s, got %s", tt.expectedStartDateStr, startStr)
			}
			if endStr != tt.expectedEndDateStr {
				t.Errorf("End date: expected %s, got %s", tt.expectedEndDateStr, endStr)
			}

			// 検証: 常に7日間の範囲であること
			expectedDays := 7
			actualDays := int(end.Sub(start).Hours()/24) + 1
			if actualDays != expectedDays {
				t.Errorf("Expected %d days range, got %d days", expectedDays, actualDays)
			}

			// 検証: StartDate は必ず土曜日であること
			if start.Weekday() != time.Saturday {
				t.Errorf("Start date must be Saturday, got %v", start.Weekday())
			}

			// 検証: EndDate は必ず金曜日であること
			if end.Weekday() != time.Friday {
				t.Errorf("End date must be Friday, got %v", end.Weekday())
			}
		})
	}
}
