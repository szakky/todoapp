package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	var err error
	db,err = sql.Open("mysql", "root:password!@tcp(127.0.0.1:3306)/todo_app?parseTime=true")
	if err != nil {
		log.Fatal("DB接続エラー:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("DB接続エラー:", err)
	}
	fmt.Println("MySQLに接続成功")

}
