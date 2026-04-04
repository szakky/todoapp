package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
	"todo-api/db"
	"html/template"

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
	fmt.Println("db connected")
    fmt.Println("ready")

	http.HandleFunc("/top", topPage)
	http.HandleFunc("/login", enterRoom)
	http.HandleFunc("/", roomPage)
	http.HandleFunc("/add", add)
	http.HandleFunc("/list", list)
	http.HandleFunc("/update", updateTask)
	http.HandleFunc("/delete", deleteTask)
	http.HandleFunc("/delete-all", deleteAll)
	fmt.Println("waiting for requests...")
	http.ListenAndServe(":8080", nil)
}

type TaskView struct {
	ID         int
	Title      string
	Categorize string
	Memo       string
	TagColor   string
}

type FrontPageData struct {
	StreakCount int
	Tasks       []TaskView
}

func roomPage(w http.ResponseWriter, r *http.Request) {
	streakCount := getStreak(conn)

	jst := time.FixedZone("JST", 9*60*60)
	todayStr := time.Now().In(jst).Format("2006-01-02")

	rows, err := conn.Query("SELECT id, title, categorize, memo FROM tasks WHERE done = 0 AND DATE(created_at) = ?", todayStr)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []TaskView
	for rows.Next() {
		var t TaskView
		if err := rows.Scan(&t.ID, &t.Title, &t.Categorize, &t.Memo); err != nil {
			http.Error(w, "DB Error", http.StatusInternalServerError)
			return
		}

		if t.Categorize != "" {
			t.TagColor = getColorForTag(t.Categorize)
		}

		tasks = append(tasks, t)
	}

	data := FrontPageData{
		StreakCount: streakCount,
		Tasks:       tasks,
	}

	tmpl, err := template.ParseFiles("templates/room.html")
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		log.Println("template parse error:", err)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "template execute error", http.StatusInternalServerError)
		log.Println("template execute error:", err)
		return
	}
}