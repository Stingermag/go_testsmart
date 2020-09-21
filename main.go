package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type reqeststruct struct {
	Into []struct {
		Column string `json:"column"`
		Size   string `json:"size"`
	} `json:"into"`
	Table  string `json:"table"`
	Values []struct {
		Value string `json:"value"`
	} `json:"values"`
}

//функция удаляет старые логи
func deloldlogs(logfile *os.File) {
	fileScanner := bufio.NewScanner(logfile)
	lineCount := 0
	for fileScanner.Scan() {
		lineCount++
	}
	//количество сохраняемых строк логов
	if lineCount > 40 {

		input, err := ioutil.ReadFile("test.log")
		if err != nil {
			fmt.Println(err)

		}
		err = ioutil.WriteFile("test_old.log", input, 0644)
		if err != nil {
			fmt.Println("Error creating", "test_old.log")
			fmt.Println(err)
		}
		logfile.Close()
		os.Create("test.log")
	}
}

//валидация размерности и количества аргументов
func validatesize(requesrobj reqeststruct) bool {

	if len(requesrobj.Into) == 0 || len(requesrobj.Values) == 0 {
		log.Print("ERROR \t", "Number of arguments for insert to bata base is null")
		return true
	}
	if len(requesrobj.Into) != len(requesrobj.Values) {
		log.Print("ERROR \t", "Number of arguments ("+strconv.Itoa(len(requesrobj.Into))+") not equal to the number of filled table fields ("+strconv.Itoa(len(requesrobj.Values))+")")
		return true
	}
	for i := 0; i < len(requesrobj.Into); i++ {
		strsize, _ := strconv.Atoi(requesrobj.Into[i].Size)
		if strsize < len(requesrobj.Values[i].Value) {
			log.Print("ERROR \t", "The received data from the request is not valid. Argument ["+requesrobj.Values[i].Value+"] ("+strconv.Itoa(len(requesrobj.Values[i].Value))+") is outside the field size "+requesrobj.Into[i].Column+"("+requesrobj.Into[i].Size+")")
			return true
		}
	}
	return false
}

//создвние запроса и ввод в бд
func insertToBD(requesrobj reqeststruct, db *sql.DB) {
	//создание строки запроса
	log.Print("INFO \t", "Try to access the database to write a table "+requesrobj.Table)

	var querry = "insert into " + requesrobj.Table + " ( " + requesrobj.Into[0].Column
	for i := 1; i < len(requesrobj.Into); i++ {
		querry = querry + ", " + requesrobj.Into[i].Column
	}
	querry = querry + " ) values ( \"" + requesrobj.Values[0].Value + "\" "
	for i := 1; i < len(requesrobj.Values); i++ {
		querry = querry + ", \"" + requesrobj.Values[i].Value + "\""
	}
	querry = querry + ")"
	log.Print("INFO \t", "Database query prepared: ["+querry+"]")

	//инсерт в бд
	rows, err := db.Query(querry)
	if err != nil {
		log.Print("ERROR \t", "Insert to the table "+requesrobj.Table+" failed "+err.Error())
		return
	}
	defer rows.Close()
	log.Print("INFO \t", "The row was successfully added to the table "+requesrobj.Table)
}

//функция с подключением к бд и записью тела запроса
func test(rw http.ResponseWriter, req *http.Request) {

	//указание вывода лого в файл
	logfile, err := os.OpenFile("test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(logfile)
	defer logfile.Close()
	deloldlogs(logfile)
	logfile, err = os.OpenFile("test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(logfile)

	//запись тела запроса в структуру
	var requesrobj reqeststruct
	json.NewDecoder(req.Body).Decode(&requesrobj)
	if validatesize(requesrobj) {
		return
	}

	//подключение к бд
	log.Print("INFO \t", "Trying to connect to the database ")
	db, err := sql.Open("mysql", "root:password@/go_testsmart_user")
	if err != nil {
		panic(err)
	}

	//проверка на успешное подключение
	if err = db.Ping(); err != nil {
		log.Print("ERROR \t", "Error while connecting to database "+err.Error())
		//период переподключения
		timer1 := time.NewTimer(time.Second * 3)
		go func() {
			//кличество попыток переподключения
			timescount := 5
			for i := 0; i < timescount; i++ {
				<-timer1.C
				db.Close()
				log.Print("INFO \t", "Trying to reconnect to the database")
				//переподключение к бд
				db, err = sql.Open("mysql", "root:password@/go_testsmart_user")
				if err != nil {
					panic(err)
				}

				//проверка на успешное подключение
				if err = db.Ping(); err != nil {
					log.Print("ERROR \t", "Error while connecting to database "+err.Error())
					{
						timer1 = time.NewTimer(time.Second * 3)
						continue
					}
				} else {
					defer db.Close()
					log.Print("INFO \t", "Database connection was successful ")

					//функция ввода в базу данных
					insertToBD(requesrobj, db)
					return

				}
			}
		}()
	} else {
		defer db.Close()
		log.Print("INFO \t", "Database connection was successful")

		//функция ввода в базу данных
		insertToBD(requesrobj, db)
	}
}

func main() {

	http.HandleFunc("/", test)
	log.Fatal(http.ListenAndServe(":3001", nil))
}
