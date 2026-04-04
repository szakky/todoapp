package main

import (
	"net/http"
	"log"
	"html/template"
	"fmt"
	"strconv"

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

	http.Redirect(w, r, "/"+roomID, http.StatusSeeOther)
}

func add(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	categorize := r.URL.Query().Get("categorize")
	memo := r.URL.Query().Get("memo")
	doneGet := r.URL.Query().Get("done")
	roomID := r.URL.Query().Get("room_id")

	done := false
	if doneGet == "true" {
		done = true
	}

	_, err := conn.Exec("INSERT INTO tasks (title, categorize, memo,done) VALUES (?, ?, ?, ?)", title, categorize, memo, done)
	if err != nil {
		fmt.Printf("Added failed: %v\n", err)
		http.Error(w,"Added failed", http.StatusInternalServerError)
		return
	}
	
	if roomID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w,r,"/"+roomID, http.StatusSeeOther)
}

func list(w http.ResponseWriter, r *http.Request) {
	rows,err := conn.Query("SELECT id, title, categorize, memo, done FROM tasks WHERE done = 0")
	if err != nil {
		fmt.Fprintf(w,"Loading error: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Fprintln(w, "--All Tasks--:")

	for rows.Next() {
		var id int
		var title string
		var categorize string
		var memo string
		var done bool
		
		if err := rows.Scan(&id, &title, &categorize, &memo, &done); err != nil {
			fmt.Fprintf(w, "Loading error: %v\n", err)
			return
		}

		fmt.Fprintf(w, "%d: %s (Categorize: %s, Memo: %s, Done: %v)\n", id, title, categorize, memo, done)
	}

	if err := rows.Err(); err != nil {
		fmt.Fprintf(w, "Loading error: %v\n", err)
	}
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

	_, err = conn.Exec("UPDATE tasks SET memo = ?, categorize = ? WHERE id = ?", memo, categorize, id)
	if err != nil {
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+roomID, http.StatusSeeOther)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	roomID := r.URL.Query().Get("room_id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = conn.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+roomID, http.StatusSeeOther)
}

func deleteAll(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")

	_, err := conn.Exec("TRUNCATE TABLE tasks")
	if err != nil {
		fmt.Printf("Delete failed: %v\n", err)
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/"+roomID, http.StatusSeeOther)
}