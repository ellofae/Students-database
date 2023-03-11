package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Entry struct {
	FullName string
	Phone    string
	Age      string
	Major    string
}

var DATA []Entry
var tFile string
var dataFile string = "./data.txt"

func myHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Host: %s Path: %s\n", r.Host, r.URL.Path)
	myT := template.Must(template.ParseGlob(tFile))
	myT.ExecuteTemplate(w, tFile, DATA)
}

func main() {
	arguments := os.Args
	if len(arguments) != 3 {
		fmt.Println("Database file + Template file")
		return
	}

	database := arguments[1]
	tFile = arguments[2]

	db, err := sql.Open("sqlite3", database)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Emptying database")
	_, err = db.Exec("DELETE FROM data")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Populating", database)
	stmt, _ := db.Prepare("INSERT INTO data(fullname, phone, age, major) values(?,?,?,?)")

	f, err := os.Open(dataFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
			return
		}

		if strings.Contains(line, "//") {
			continue
		}

		lineData := strings.Split(line, "|")
		for _, value := range lineData {
			value = strings.Trim(value, " ")
		}

		_, _ = stmt.Exec(lineData[0], lineData[1], lineData[2], lineData[3])
	}

	rows, err := db.Query("SELECT * FROM data")
	if err != nil {
		fmt.Println(err)
		return
	}

	var fullName string
	var phone string
	var age string
	var major string
	for rows.Next() {
		err = rows.Scan(&fullName, &phone, &age, &major)
		temp := Entry{FullName: fullName, Phone: phone, Age: age, Major: major}
		DATA = append(DATA, temp)
	}

	http.HandleFunc("/", myHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
