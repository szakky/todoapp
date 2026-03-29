package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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

	todoTableSQL := `
    CREATE TABLE IF NOT EXISTS tasks (
        id INT AUTO_INCREMENT PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
		categorize VARCHAR(255) NOT NULL,
		done BOOLEAN NOT NULL DEFAULT FALSE,
		memo TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

    _, err = db.Exec(todoTableSQL)
    if err != nil {
        log.Fatal("error:", err)
    }
    fmt.Println("ready")

	http.HandleFunc("/add", add)
	http.HandleFunc("/list", list)
	http.HandleFunc("/update-task", updateTask)
	http.HandleFunc("/delete", deleteTask)
	http.HandleFunc("/delete-all", deleteAll)
	http.HandleFunc("/", front)
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

func updateTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	memo := r.URL.Query().Get("memo")
	categorize := r.URL.Query().Get("categorize") 

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE tasks SET memo = ?, categorize = ? WHERE id = ?", memo, categorize, id)
	if err != nil {
		http.Error(w, "データベースの更新に失敗しました", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		http.Error(w, "削除に失敗しました", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
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

// タグの文字から自動的に色を決定する関数
func getColorForTag(tag string) string {
	// 🍎 カッコいい色のパレット（8色）を用意
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

	// タスクの文字を1文字ずつ足し算して、独自の「背番号」を作る
	var hash int
	for _, char := range tag {
		hash += int(char)
	}

	// 背番号を色の数(8)で割った「余り」を使って色を選ぶ！
	return colors[hash%len(colors)]
}

func front(w http.ResponseWriter, r *http.Request) {
	// 1. 継続日数を取得
	streakCount := getStreak(db)

	// 2. 今日の日付を取得
	jst := time.FixedZone("JST", 9*60*60)
	todayStr := time.Now().In(jst).Format("2006-01-02")

	// 3. 今日のタスクだけをDBから取得
	rows, err := db.Query("SELECT id, title, categorize, memo FROM tasks WHERE done = 0 AND DATE(created_at) = ?", todayStr)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close() // ← ダブっていた部分は1つにまとめました！

	var tasksHTML string
	for rows.Next() {
		var id int
		var title, categorize, memo string

		if err := rows.Scan(&id, &title, &categorize, &memo); err == nil {
			tagHTML := ""
			if categorize != "" {
				tagColor := getColorForTag(categorize)
				tagHTML = fmt.Sprintf(`<span style="background-color: %s; color: white; padding: 3px 8px; border-radius: 12px; font-size: 12px; margin-left: 10px; vertical-align: middle;">%s</span>`, tagColor, categorize)
			}

			tasksHTML += fmt.Sprintf(`
            <div class="task-item" onclick="toggleDetails(%d)">
                <input type="checkbox" name="task" value="%d" onclick="event.stopPropagation()">
                <div class="task-content" style="width: 100%%;"> 
                    <div class="task-title" style="display: flex; align-items: center; justify-content: space-between;">
                        <div style="display: flex; align-items: center;">
                            %s %s
                        </div>
                        <a href="/delete?id=%d" onclick="event.stopPropagation(); return confirm('このタスクを削除しますか？');" style="color: #ff3b30; font-size: 14px; font-weight: bold; text-decoration: none; padding: 4px 8px; border: 1px solid #ff3b30; border-radius: 4px;">Delete</a>
                    </div>
                    
                    <div id="details-%d" class="task-details" style="display: none;" onclick="event.stopPropagation()">
                        <form action="/update-task" method="GET" style="margin-top: 10px; display: flex; flex-direction: column; gap: 10px;">
                            <input type="hidden" name="id" value="%d">
                            
                            <div style="display: flex; align-items: center; gap: 5px;">
                                <span class="label" style="width: 50px;">TAG:</span>
                                <input type="text" name="categorize" value="%s" placeholder="タグ (例: english)" style="flex-grow: 1; padding: 5px; border: 1px solid #ccc; border-radius: 4px;">
                            </div>
                            
                            <div style="display: flex; align-items: center; gap: 5px;">
                                <span class="label" style="width: 50px;">MEMO:</span>
                                <input type="text" name="memo" value="%s" placeholder="メモを入力..." style="flex-grow: 1; padding: 5px; border: 1px solid #ccc; border-radius: 4px;">
                            </div>

                            <button type="submit" style="align-self: flex-end; padding: 5px 15px; background-color: #007aff; color: white; border: none; border-radius: 4px; cursor: pointer;">保存</button>
                        </form>
                    </div>
                </div>
            </div>`, id, id, title, tagHTML, id, id, id, categorize, memo)
		}
	}

	if tasksHTML == "" {
		tasksHTML = `<p style="text-align:center; color:#888;">タスクはありません</p>`
	}

	// ※ここにあった 2回目の streakCount := getStreak(db) は削除しました！

	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html lang="ja">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Todo App</title>
        <style>
            body { font-family: sans-serif; background-color: #f5f5f7; padding: 20px; display: flex; justify-content: center; }
            .app-container { width: 100%%; max-width: 400px; }
            .streak-banner {
                background: linear-gradient(135deg, #ff9500, #ff3b30);
                color: white;
                text-align: center;
                padding: 10px;
                border-radius: 8px;
                font-size: 20px;
                font-weight: bold;
                margin-bottom: 20px;
                box-shadow: 0 4px 6px rgba(255, 59, 48, 0.3);
            }
            .input-form { display: flex; gap: 10px; margin-bottom: 30px; }
            .input-form input[type="text"] { flex-grow: 1; padding: 12px; border: 1px solid #ddd; border-radius: 5px; font-size: 16px; }
            .input-form button { padding: 12px 20px; background-color: #e5e5ea; color: #007aff; border: none; border-radius: 5px; font-size: 16px; font-weight: bold; cursor: pointer; }
            .task-list { margin-bottom: 20px; }
            .task-item { background-color: white; padding: 15px; margin-bottom: 10px; border-radius: 12px; box-shadow: 0 2px 5px rgba(0,0,0,0.05); display: flex; align-items: flex-start; font-size: 18px; cursor: pointer; transition: background-color 0.2s; }
            .task-item:hover { background-color: #f9f9fb; }
            .task-item input[type="checkbox"] { transform: scale(1.5); margin-right: 15px; margin-top: 5px; cursor: pointer; }
            .task-content { flex-grow: 1; }
            .task-title { font-weight: bold; }
            .task-details { margin-top: 10px; font-size: 14px; color: #555; background-color: #f0f0f5; padding: 10px; border-radius: 8px; border-left: 4px solid #007aff; }
            .task-details p { margin: 5px 0; }
            .label { font-weight: bold; color: #333; font-size: 12px; text-transform: uppercase; }
            .btn-delete { background-color: #e5e5ea; color: #007aff; border: none; padding: 10px 20px; border-radius: 5px; font-size: 18px; cursor: pointer; margin-bottom: 20px; }
        </style>
    </head>
    <body>
        <div class="app-container">
            
            <div class="streak-banner">🔥 %d日継続中!! 🔥</div>

            <form action="/add" method="GET" class="input-form">
                <input type="text" name="title" placeholder="Enter a task" required style="flex-grow: 1; padding: 12px; border: 1px solid #ddd; border-radius: 5px; font-size: 16px;">
                <input type="hidden" name="categorize" value=""> <button type="submit">Add</button>
            </form>

            <div class="task-list">
                %s
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
        </div>
    </body>
    </html>
    `, streakCount, tasksHTML)

	fmt.Fprint(w, html)
}