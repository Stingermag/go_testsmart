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
	Table string `json:"table"`
	Into struct {
		Table1 string `json:"table1"`
		Table2 string `json:"table2"`
		Table3 string `json:"table3"`
	}
	Values struct {
		Name string `json:"name"`
		Surname string `json:"surname"`
		Age int `json:"age"`
	}
}

func dellogs(logfile *os.File){
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

	log.Printf("%v\n", "Попытка обращения к бд")

	var querry = "insert into " + t.Table + " ( " + t.Into.Table1 +" , " + t.Into.Table2 +" , " + t.Into.Table3 + " ) values (?,?,?)"
	//инсерт в бд
	res, err := db.Prepare(querry)
	if err != nil {
		log.Printf("%v\n", "Запись не удалась "+  t.Values.Name +" "+ t.Values.Surname + " ERROR: " + err.Error())
	//	panic(err)
		return
	}
	_, er := res.Exec(t.Values.Name,t.Values.Surname,t.Values.Age)
	if er != nil {
		log.Printf("%v\n", "Запись не удалась "+  t.Values.Name +" "+ t.Values.Surname + " ERROR: " + er.Error())
		return
	//	panic(er.Error())
	}

	log.Printf("%v\n", "Запись успешна "+  t.Values.Name +" "+ t.Values.Surname)
}

func test(rw http.ResponseWriter, req *http.Request) {

	//указание вывода лого в файл
	logfile, err := os.OpenFile("test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(logfile)
	defer logfile.Close()
	dellogs(logfile)
	logfile, err = os.OpenFile("test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(logfile)
	//взятие тела запроса
	dec, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		log.Fatalln(err)
		return
	}
	var re reqestt
	err = json.Unmarshal(dec,&re)
	if err != nil {
		log.Fatalln(err)
		return
	}


	log.Printf("%v\n", "Попытка подключения к бд")

	//подключение к бд
	db, err := sql.Open("mysql", "root:password@/go_testsmart_user")
	if err != nil {
		panic(err)
	}

	//проверка на успешное подключение
	if db.Ping() != nil {
		log.Printf("%v\n", "Ошибка при подключении к бд go_testsmart_user ")
		connect := 0
		timer1 := time.NewTimer(time.Second*10)
		{
			<-timer1.C
			db.Close()
			log.Printf("%v\n", "Попытка переподключения к бд")
			//подключение к бд
			db, err = sql.Open("mysql", "root:password@/go_testsmart_user")
			if err != nil {
				panic(err)
			}
			log.Printf("%v\n", "Попытка переподключения к бд2")
			//проверка на успешное подключение
			if db.Ping() != nil {
				log.Printf("%v\n", "Ошибка при повторном подключении к бд go_testsmart_user ")
				return
			}
			fmt.Print("успешно")
			log.Printf("%v\n", "Попытка переподключения к бд3")
			connect = 1
		}
		if(connect == 0){
			fmt.Print("неуспешно")
			return

		}
		fmt.Print("успешно")

	}
	fmt.Print("324567")
	defer db.Close()
	fmt.Print("324567")
	log.Printf("%v\n", "Подключение успешно")

	//обработка входящих запросов
	insertToBD(re, db)
}

func main() {

	http.HandleFunc("/", test)
	log.Fatal(http.ListenAndServe(":3001", nil))
}