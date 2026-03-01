package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	var err error
	db,err = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/todo_app?parseTime=true")
	if err != nil {
		log.Fatal("DB接続エラー:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("DB接続エラー:", err)
	}
	fmt.Println("MySQLに接続成功")

	todoTableSQL := `
    CREATE TABLE IF NOT EXISTS tasks (
        id INT AUTO_INCREMENT PRIMARY KEY,
        title VARCHAR(255) NOT NULL
    );`
    _, err = db.Exec(todoTableSQL)
    if err != nil {
        log.Fatal("テーブル作成エラー:", err)
    }
    fmt.Println("tasksテーブルの準備もOKです！")

	http.HandleFunc("/add", add)
	fmt.Println("サーバー起動: 待機中です...")
	http.ListenAndServe(":8080", nil)
}

func add(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")

	_, err := db.Exec("INSERT INTO tasks (title) VALUES (?)", title)
	if err != nil {
		fmt.Printf("タスク追加できなかった: %v\n", err)
		fmt.Fprintf(w, "タスク追加できなかった: %v\n", err)
		return
	}
	fmt.Printf("タスクを追加できた: %s\n", title)
	fmt.Fprintf(w, "タスクを追加できた: %s\n", title)
}