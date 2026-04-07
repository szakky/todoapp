package main

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func getStreak(db *sql.DB) int {
	rows,err := db.Query("SELECT DISTINCT DATE(created_at) FROM tasks WHERE created_at IS NOT NULL ORDER BY DATE(created_at) DESC")
	if err != nil {
		return 0
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var dateStr string
		if err := rows.Scan(&dateStr); err != nil {
			dates = append(dates, dateStr)
		}
	}
	if len(dates) == 0 {
		return 0
	}

	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")
	yesterdayStr := today.AddDate(0, 0, -1).Format("2006-01-02")

	if dates[0] != todayStr && dates[0] != yesterdayStr {
		return 0
	}

	streak := 0
	checkDate := today
	if dates[0] == yesterdayStr {
		checkDate = checkDate.AddDate(0, 0, -1)
	}

	for _, d := range dates {
		if d == checkDate.Format("2006-01-02") {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak

	
}

func getColorForTag(tag string) string {
	colors := []string{
		"#007aff", // ブルー
		"#34c759", // グリーン
		"#ff9500", // オレンジ
		"#ff3b30", // レッド
		"#af52de", // パープル
		"#5856d6", // インディゴ
		"#ff2d55", // ピンク
		"#00c7be", // ティール（青緑）
	}

	var hash int
	for _, char := range tag {
		hash += int(char)
	}
	
	return colors[hash%len(colors)]
}