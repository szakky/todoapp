package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"
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

type TaskView struct {
	ID         int
	Title      string
	Categorize string
	Memo       string
	TagColor   string
}

func roomPage(w http.ResponseWriter, r *http.Request) {

	jst := time.FixedZone("JST", 9*60*60)
	todayStr := time.Now().In(jst).Format("2006-01-02")
	roomID := r.URL.Query().Get("room_id")

	rows, err := conn.Query("SELECT id, title, categorize, COALESCE(memo, '') FROM tasks WHERE done = 0 AND DATE(created_at) = ? AND room_id = ?", todayStr, roomID)
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

	tmpl, err := template.ParseFiles("templates/room.html")
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		log.Println("template parse error:", err)
		return
	}

	data := struct {
		Tasks  []TaskView
		RoomID string
	}{
		Tasks:  tasks,
		RoomID: roomID,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "template execute error", http.StatusInternalServerError)
		log.Println("template execute error:", err)
		return
	}
}
