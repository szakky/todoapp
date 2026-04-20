package main

import (

	_ "github.com/go-sql-driver/mysql"
)

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