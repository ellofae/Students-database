package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Entry struct {
	FullName string
	Phone    string
	Age      string
	Major    string
	IP       string
	DateTime string
}

type Attend struct {
	IP   string
	Time string
}

var DATA []Entry
var IP_TIME []Attend
var tFile string
var dataFile string = "./data.txt"
var ipTimeFile string = "./ip_date.txt"

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

	// Open database
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Modify the base
	fmt.Println("Emptying database")
	_, err = db.Exec("DELETE FROM data")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Populating", database)
	stmt, _ := db.Prepare("INSERT INTO data(fullname, phone, age, major, ip, datetime) values(?,?,?,?,?,?)")

	// Open file with data on students
	studFile, err := os.Open(dataFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer studFile.Close()

	// Open file with ip and time info
	attendFile, err := os.Open(ipTimeFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer attendFile.Close()

	studReader := bufio.NewReader(studFile)
	attendReader := bufio.NewReader(attendFile)
	for {
		lineData, err := ReadAndCheck(studReader)
		if err == 1 {
			break
		} else if err == 2 {
			fmt.Println("Error during reading the file")
			return
		} else if err == 3 {
			continue
		}

		attendData, err := ReadAndCheck(attendReader)
		if err == 1 {
			break
		} else if err == 2 {
			fmt.Println("Error during reading the file")
			return
		} else if err == 3 {
			continue
		}

		check := checkDate(attendData)
		if check != nil {
			continue
		}

		check = checkIPv4(attendData)
		if check != nil {
			continue
		}

		// Fill data into a table
		_, _ = stmt.Exec(lineData[0], lineData[1], lineData[2], lineData[3], attendData[0], attendData[1])
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
	var ip string
	var timeDate string
	for rows.Next() {
		err = rows.Scan(&fullName, &phone, &age, &major, &ip, &timeDate)
		temp := Entry{FullName: fullName, Phone: phone, Age: age, Major: major, IP: ip, DateTime: timeDate}
		DATA = append(DATA, temp)
	}

	http.HandleFunc("/", myHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Read and format the file
func ReadAndCheck(reader *bufio.Reader) ([]string, int) {
	line, err := reader.ReadString('\n')
	if err == io.EOF {
		return nil, 1
	} else if err != nil {
		fmt.Println(err)
		return nil, 2
	}

	if strings.Contains(line, "//") {
		return nil, 3
	}

	lineData := strings.Split(line, "|")
	for _, value := range lineData {
		value = strings.Trim(value, " ")
	}

	return lineData, 0
}

// Validate date and time format
func checkDate(line []string) error {
	r := regexp.MustCompile(`.*\[(\d\d\/\w+/\d\d\d\d:\d\d:\d\d:\d\d.*)\].*`)
	if r.MatchString(line[1]) {
		match := r.FindStringSubmatch(line[1])

		dt, err := time.Parse("02/Jan/2006:15:04:05 -0700", match[1])
		if err == nil {
			newFormat := dt.Format(time.RFC850)
			line[1] = newFormat

			//userLog.Printf("Attend date: %s ", line[1])
		} else {
			//errLog.Println("Invalid date format!")
			line[1] = "None"
		}

		return nil
	} else {
		return errors.New("Not a valid date format")
	}
}

// Validate IPv4 address format
func checkIPv4(line []string) error {
	partIP := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	grammar := partIP + "\\." + partIP + "\\." + partIP + "\\." + partIP
	r := regexp.MustCompile(grammar)

	ip := r.FindString(line[0])
	trial := net.ParseIP(ip)
	if trial.To4() == nil {
		//userLog.Println(" Invalid ip address: %s\n", line[0])
		line[0] = "None"
		return errors.New("Not a valid ip address format")
	}
	//userLog.Printf(" IPv4: %s \n", line[0])

	return nil
}
