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
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type reqestt struct {
	Into []struct {
		Column string `json:"column"`
	} `json:"into"`
	Table  string `json:"table"`
	Values []struct {
		Value string `json:"value"`
	} `json:"values"`
}

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
		fmt.Println("vases:")
		logfile.Close()
		os.Create("test.log")
	}
}

func insertToBD(t reqestt, db *sql.DB) {
	//создание строки запроса
	log.Print("INFO \t", "Попытка обращения к базе данных для записи таблицы "+t.Table)
	var querry = "insert into " + t.Table + " ( " + t.Into[0].Column
	for i := 1; i < len(t.Into); i++ {
		querry = querry + ", " + t.Into[i].Column
	}
	querry = querry + " ) values ( \"" + t.Values[0].Value + "\" "
	for i := 1; i < len(t.Values); i++ {
		querry = querry + ", \"" + t.Values[i].Value + "\""
	}
	querry = querry + ")"

	//инсерт в бд
	rows, err := db.Query(querry)
	if err != nil {
		log.Print("ERROR \t", "Запись не удалась в таблицу "+t.Table+" ERROR: "+err.Error())
		return
	}
	defer rows.Close()
	log.Print("INFO \t", "Запись успешна в таблицу "+t.Table)
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
	var re reqestt
	json.NewDecoder(req.Body).Decode(&re)

	//подключение к бд
	log.Print("INFO \t", "Попытка подключения к базе данных")
	db, err := sql.Open("mysql", "root:password@/go_testsmart_user")
	if err != nil {
		panic(err)
	}

	//проверка на успешное подключение
	if db.Ping() != nil {
		log.Print("ERROR \t", "Ошибка при подключении к базе данных. База данных выключена.")

		timer1 := time.NewTimer(time.Second * 10)
		go func() {
			<-timer1.C
			db.Close()
			log.Print("INFO \t", "Попытка повторного подключения к базе данных")
			//переподключение к бд
			db, err = sql.Open("mysql", "root:password@/go_testsmart_user")
			if err != nil {
				panic(err)
			}

			//проверка на успешное подключение
			if db.Ping() != nil {
				log.Print("ERROR \t", "Ошибка при подключении к базе данных. База данных выключена.")
				return
			} else {
				defer db.Close()
				log.Print("INFO \t", "Подключение к базе данных прошло успешно")

				//функция обработки входящих запросов
				insertToBD(re, db)
			}
		}()
	} else {
		defer db.Close()
		log.Print("INFO \t", "Подключение к базе данных прошло успешно")

		//функция обработки входящих запросов
		insertToBD(re, db)
	}

}

func main() {

	http.HandleFunc("/", test)
	log.Fatal(http.ListenAndServe(":3001", nil))
}
