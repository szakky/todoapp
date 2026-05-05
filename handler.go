package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func topPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/top.html")
	if err != nil {
		http.Error(w, "parse Error", http.StatusInternalServerError)
		log.Println("parse error:", err)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "execute Error", http.StatusInternalServerError)
		log.Println("execute error:", err)
		return
	}
}

func enterRoom(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	if roomID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/room/?room_id="+roomID, http.StatusSeeOther)
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

func add(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	categorize := r.URL.Query().Get("categorize")
	memo := r.URL.Query().Get("memo")
	roomID := r.URL.Query().Get("room_id")

	_, err := conn.Exec("INSERT INTO tasks (title, categorize, memo, room_id) VALUES (?, ?, ?, ?)", title, categorize, memo, roomID)
	if err != nil {
		log.Printf("Added failed: %v\n", err)
		http.Error(w, "Added failed", http.StatusInternalServerError)
		return
	}

	if roomID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/room/?room_id="+roomID, http.StatusSeeOther)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	memo := r.URL.Query().Get("memo")
	categorize := r.URL.Query().Get("categorize")
	roomID := r.URL.Query().Get("room_id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = conn.Exec("UPDATE tasks SET memo = ?, categorize = ? WHERE id = ? AND room_id = ?", memo, categorize, id, roomID)
	if err != nil {
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/room/?room_id="+roomID, http.StatusSeeOther)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	roomID := r.URL.Query().Get("room_id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = conn.Exec("DELETE FROM tasks WHERE id = ? AND room_id = ?", id, roomID)
	if err != nil {
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/room/?room_id="+roomID, http.StatusSeeOther)
}

func deleteAll(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")

	_, err := conn.Exec("DELETE FROM tasks WHERE room_id = ?", roomID)
	if err != nil {
		log.Printf("Delete failed: %v\n", err)
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/room/?room_id="+roomID, http.StatusSeeOther)
}
