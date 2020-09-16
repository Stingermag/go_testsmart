package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
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

func deloldlogs(logfile *os.File){
	fileScanner := bufio.NewScanner(logfile)
	lineCount := 0
	for fileScanner.Scan() {
		lineCount++
	}
	fmt.Println("number of lines logs:", lineCount)
	if(lineCount > 20){

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

func insertToBD(t reqestt, db *sql.DB){
	//создание строки запроса
	log.Printf("%v\n", "Попытка обращения к базе данных для записи таблицы " + t.Table)
	var querry = "insert into " + t.Table + " ( " + t.Into[0].Column
	for i := 1; i < len(t.Into); i++{
		querry = querry + ", " + t.Into[i].Column
	}
	querry = querry + " ) values ( \"" + t.Values[0].Value + "\" "
	for i := 1; i < len(t.Values); i++{
		querry = querry + ", \"" + t.Values[i].Value+ "\""
	}
	querry = querry + ")"

	//инсерт в бд
	rows, err := db.Query(querry)
	if err != nil {
		log.Printf("%v\n", "Запись не удалась в таблицу "+  t.Table + " ERROR: " + err.Error())
		return
	}
	defer rows.Close()
	log.Printf("%v\n", "Запись успешна в таблицу "+  t.Table )
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


	//запись тела запроса в стракт
	var re reqestt
	json.NewDecoder(req.Body).Decode(&re)


	//подключение к бд
	log.Printf("%v\n", "Попытка подключения к базе данных")
	db, err := sql.Open("mysql", "root:password@/go_testsmart_user")
	if err != nil {
		panic(err)
	}


	//проверка на успешное подключение
	if db.Ping() != nil {
		log.Printf("%v\n", "Ошибка при подключении к базе данных ")
		timer1 := time.NewTimer(time.Second*10)
		go func(){
			<-timer1.C
			db.Close()
			log.Printf("%v\n", "Попытка переподключения к базе данных")
			//переподключение к бд
			db, err = sql.Open("mysql", "root:password@/go_testsmart_user")
			if err != nil {
				panic(err)
			}

			//проверка на успешное подключение
			if db.Ping() != nil {
				log.Printf("%v\n", "Не удалось повторно записать строку в бд go_testsmart_user ")
				return
			}else {
				defer db.Close()
				log.Printf("%v\n", "Подключение успешно")

				//функция обработки входящих запросов
				insertToBD(re, db)
			}
		}()
	} else {
		defer db.Close()
		log.Printf("%v\n", "Подключение успешно")

		//функция обработки входящих запросов
		insertToBD(re, db)
	}

}

func main() {

	http.HandleFunc("/", test)
	log.Fatal(http.ListenAndServe(":3001", nil))
}