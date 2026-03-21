package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" {
		log.Fatal("error: DB_USER is not set")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("db error:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("db error:", err)
	}
	fmt.Println("db connected")

	db.Exec("DROP TABLE IF EXISTS tasks")

	todoTableSQL := `
    CREATE TABLE IF NOT EXISTS tasks (
        id INT AUTO_INCREMENT PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
		categorize VARCHAR(255) NOT NULL,
		done BOOLEAN NOT NULL DEFAULT FALSE,
		memo TEXT
    );`
    _, err = db.Exec(todoTableSQL)
    if err != nil {
        log.Fatal("error:", err)
    }
    fmt.Println("ready")

	http.HandleFunc("/add", add)
	http.HandleFunc("/list", list)
	http.HandleFunc("/", front)
	http.HandleFunc("/delete-all", deleteAll)
	fmt.Println("waiting for requests...")
	http.ListenAndServe(":8080", nil)
}

func add(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	categorize := r.URL.Query().Get("categorize")
	memo := r.URL.Query().Get("memo")
	doneGet := r.URL.Query().Get("done")

	done := false
	if doneGet == "true" {
		done = true
	}

	_, err := db.Exec("INSERT INTO tasks (title, categorize, memo,done) VALUES (?, ?, ?, ?)", title, categorize, memo, done)
	if err != nil {
		fmt.Printf("Added failed: %v\n", err)
		http.Error(w,"Added failed", http.StatusInternalServerError)
		return
	}
	
	http.Redirect(w,r,"/", http.StatusSeeOther)
}

func list(w http.ResponseWriter, r *http.Request) {
	rows,err := db.Query("SELECT id, title, categorize, memo, done FROM tasks WHERE done = 0")
	if err != nil {
		fmt.Fprintf(w,"Loading error: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Fprintln(w, "--タスク一覧--:")

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

func deleteAll(w http.ResponseWriter, r *http.Request) {
	_, err := db.Exec("TRUNCATE TABLE tasks")
	if err != nil {
		fmt.Printf("Delete failed: %v\n", err)
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func front(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, categorize, memo FROM tasks WHERE done = 0")
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasksHTML string
	for rows.Next() {
		var id int
		var title, categorize, memo string
		
		if err := rows.Scan(&id, &title, &categorize, &memo); err == nil {
			tasksHTML += fmt.Sprintf(`
			<div class="task-item" onclick="toggleDetails(%d)">
				<input type="checkbox" name="task" value="%d" onclick="event.stopPropagation()">
				<div class="task-content">
					<div class="task-title">%s</div>
					<div id="details-%d" class="task-details" style="display: none;">
						<p><span class="label">Category:</span> %s</p>
						<p><span class="label">Memo:</span> %s</p>
					</div>
				</div>
			</div>`, id, id, title, id, categorize, memo)
		}
	}

	if tasksHTML == "" {
		tasksHTML = `<p style="text-align:center; color:#888;">タスクはありません</p>`
	}

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="ja">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Todo App</title>
		<style>
			body {
				font-family: sans-serif;
				background-color: #f5f5f7;
				padding: 20px;
				display: flex;
				justify-content: center;
			}
			.app-container { width: 100%%; max-width: 400px; }
			.input-form { display: flex; gap: 10px; margin-bottom: 30px; }
			.input-form input[type="text"] { flex-grow: 1; padding: 12px; border: 1px solid #ddd; border-radius: 5px; font-size: 16px; }
			.input-form button { padding: 12px 20px; background-color: #e5e5ea; color: #007aff; border: none; border-radius: 5px; font-size: 16px; font-weight: bold; cursor: pointer; }
			
			.task-list { margin-bottom: 20px; }
			
			/* デザインの調整: フレックスボックスで上部揃えに変更 */
			.task-item { background-color: white; padding: 15px; margin-bottom: 10px; border-radius: 12px; box-shadow: 0 2px 5px rgba(0,0,0,0.05); display: flex; align-items: flex-start; font-size: 18px; cursor: pointer; transition: background-color 0.2s; }
			.task-item:hover { background-color: #f9f9fb; }
			.task-item input[type="checkbox"] { transform: scale(1.5); margin-right: 15px; margin-top: 5px; cursor: pointer; }
			
			.task-content { flex-grow: 1; }
			.task-title { font-weight: bold; }
			
			/* 詳細表示用のスタイルを追加 */
			.task-details { margin-top: 10px; font-size: 14px; color: #555; background-color: #f0f0f5; padding: 10px; border-radius: 8px; border-left: 4px solid #007aff; }
			.task-details p { margin: 5px 0; }
			.label { font-weight: bold; color: #333; font-size: 12px; text-transform: uppercase; }

			.btn-delete { background-color: #e5e5ea; color: #007aff; border: none; padding: 10px 20px; border-radius: 5px; font-size: 18px; cursor: pointer; margin-bottom: 20px; }
			.btn-sm { background-color: #ff0000; color: #ffff00; border: none; padding: 15px 40px; border-radius: 5px; font-size: 18px; font-weight: bold; display: block; margin-bottom: 10px; width: 100%%; cursor: pointer; }
			.bottom-buttons { display: flex; gap: 10px; }
			.btn-victory { background-color: #ffff99; color: #555; border: none; padding: 15px 30px; border-radius: 5px; font-size: 18px; flex: 1; cursor: pointer; font-weight: bold; }
			.btn-defeat { background-color: #1c1c1e; color: #ff3b30; border: none; padding: 15px 30px; border-radius: 5px; font-size: 18px; flex: 1; cursor: pointer; font-weight: bold; }
		</style>
	</head>
	<body>
		<div class="app-container">
			
			<form action="/add" method="GET" class="input-form">
				<input type="text" name="title" placeholder="Enter a task" required>
				<input type="hidden" name="categorize" value="general">
				<input type="hidden" name="memo" value="No memo provided.">
				<button type="submit">Add</button>
			</form>

			<div class="task-list">
				%s
			</div>

			<form action="/delete-all" method="GET" onsubmit="return confirm('本当に全てのタスクを削除しますか？');">
				<button type="submit" class="btn-delete">Delete All</button>
			</form>

			<button class="btn-sm" onclick="location.href='https://www.youtube.com/watch?v=wBf47hGMch0'">SM</button>
			<div class="bottom-buttons">
				<button class="btn-victory" onclick="location.href='https://www.youtube.com/watch?v=joQPqK1W45A&list=LL&index=2'">Victory</button>
				<button class="btn-defeat" onclick="location.href='https://www.youtube.com/watch?v=n31lW0BcZ1Q'">Defeat</button>
			</div>

		</div>

		<script>
			function toggleDetails(taskId) {
				var detailsDiv = document.getElementById('details-' + taskId);
				if (detailsDiv.style.display === 'none' || detailsDiv.style.display === '') {
					detailsDiv.style.display = 'block';
				} else {
					detailsDiv.style.display = 'none';
				}
			}
		</script>
	</body>
	</html>
	`, tasksHTML)

	fmt.Fprint(w, html)
}