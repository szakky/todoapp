package main

import (
	"database/sql"
	"log"
	"net/http"
	"todo-api/db"

	_ "github.com/go-sql-driver/mysql"
)

var conn *sql.DB

func main() {
	var err error
	conn, err = db.Connect()
	if err != nil {
		log.Fatal("db error:", err)
	}
	defer conn.Close()

	if err = conn.Ping(); err != nil {
		log.Fatal("db error:", err)
	}
	log.Println("db connected")
	log.Println("ready")

	http.HandleFunc("/", topPage)
	http.HandleFunc("/login", enterRoom)
	http.HandleFunc("/room/", roomPage)
	http.HandleFunc("/add", add)
	http.HandleFunc("/update", updateTask)
	http.HandleFunc("/delete", deleteTask)
	http.HandleFunc("/delete-all", deleteAll)
	log.Println("waiting for requests...")
	http.ListenAndServe(":8080", nil)
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