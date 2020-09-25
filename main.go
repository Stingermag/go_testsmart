package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
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
type responsestruct struct {
	Body  string `json:"bady"`
	Error bool   `json:"error"`
}

var timer1 = time.NewTimer(time.Second*10)
func reconnect()  {
//	timer1 := time.NewTimer(time.Second * 10)
	go func() {
		//цикл с переподелючением
		for {
			fmt.Print("так так")
			<-timer1.C
			errorrequests = nil
			var errorreqs []reqeststruct //в этой полученные из файла
			file, err := os.OpenFile("errrequests.json", os.O_RDWR, 0666)
			var reqDecoder *json.Decoder = json.NewDecoder(file)
			if err != nil {
				log.Fatal(err)
			}
			err = reqDecoder.Decode(&errorreqs)
			if err != nil {
				log.Fatal(err)
			}
			for i,_ := range errorreqs{
				errorrequests = append(errorrequests, errorreqs[i])
			}
			file.Close()


			if len(errorrequests) == 0 {
				timer1 = time.NewTimer(time.Second * 10)
				fmt.Print("Файл пуст")
				continue
			}
				log.Print("INFO \t", "Trying to reconnect to the database")
			//переподключение к бд
			db, err := sql.Open("mysql", "root:password@/go_testsmart_user")
			if err != nil {
				panic(err)
			}

			//проверка на успешное подключение
			if err = db.Ping(); err != nil {
				log.Print("ERROR \t", "Error while connecting to database "+err.Error())
				{
					fmt.Print("")
					timer1 = time.NewTimer(time.Second * 10)
					db.Close()
					continue
				}
			} else {

				log.Print("INFO \t", "Database connection was successful ")

				//функция ввода в базу данных
				for i :=0;i< len(errorrequests);i++ {
					//insertToBD(errorrequests[i], db)
				}
				fmt.Print("Файл очищен, записано все в бд")
				//остановить таймер, очистить файл
				os.Create("errrequests.json")
				file, err = os.OpenFile("errrequests.json", os.O_RDWR, 0666)
				io.WriteString(file, "[]")
				file.Close()
				db.Close()
				timer1 = time.NewTimer(time.Second * 10)
				//timer1.Stop()
			}

		}
	}()

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
func validatesize(requestobj reqeststruct) bool {

	if len(requestobj.Into) == 0 || len(requestobj.Values) == 0 {
		log.Print("ERROR \t", "Number of arguments for insert to bata base is null")
		return true
	}
	if len(requestobj.Into) != len(requestobj.Values) {
		log.Print("ERROR \t", "Number of arguments ("+strconv.Itoa(len(requestobj.Into))+") not equal to the number of filled table fields ("+strconv.Itoa(len(requestobj.Values))+")")
		return true
	}
	for i := 0; i < len(requestobj.Into); i++ {
		strsize, _ := strconv.Atoi(requestobj.Into[i].Size)
		if strsize < len(requestobj.Values[i].Value) {
			log.Print("ERROR \t", "The received data from the request is not valid. Argument ["+requestobj.Values[i].Value+"] ("+strconv.Itoa(len(requestobj.Values[i].Value))+") is outside the field size "+requestobj.Into[i].Column+"("+requestobj.Into[i].Size+")")
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
func test(w http.ResponseWriter, req *http.Request) {



	//формирование ответа сервера
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	resp := responsestruct{
		Body:  "Data accepted for processing",
		Error: false}

	//запись тела запроса в структуру
	var requestobj reqeststruct

	json.NewDecoder(req.Body).Decode(&requestobj)

	if validatesize(requestobj) {
		resp.Error = true
		resp.Body = "Data not correct"
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			panic(err)
		}
		return
	}

	//апи ответ
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}

	//подключение к бд
	log.Print("INFO \t", "Trying to connect to the database ")
	db, err := sql.Open("mysql", "root:password@/go_testsmart_user")
	if err != nil {
		panic(err)
	}

	//проверка на успешное подключение
	if err = db.Ping(); err != nil {
		errorrequests = nil
		var errorreqs []reqeststruct //в этой полученные из файла
		log.Print("ERROR \t", "Error while connecting to database "+err.Error())
		errorrequests = append(errorrequests, requestobj)

		//fmt.Print(string(result))
		file, err := os.OpenFile("errrequests.json", os.O_RDWR, 0666)
		//file, err := os.OpenFile("errrequests.json", os.O_CREATE, os.ModePerm)
		var reqDecoder *json.Decoder = json.NewDecoder(file)
		if err != nil {
			log.Fatal(err)
		}
		err = reqDecoder.Decode(&errorreqs)
		if err != nil {
			log.Fatal(err)
		}
		for i,_ := range errorreqs{
			errorrequests = append(errorrequests, errorreqs[i])
		}
		result, error := json.Marshal(errorrequests)
		if error != nil {
			log.Fatalln(err)
		}
		file.Close()
		os.Create("errrequests.json")
		file, err = os.OpenFile("errrequests.json", os.O_RDWR, 0666)
		n, err := io.WriteString(file, string(result))
		if err != nil {
			fmt.Println(n, err)
		}
		file.Close()
	//	timer1 = time.NewTimer(time.Second * 10)
		//функция переподключения для записи незаписанных данных
		//timer1 := time.NewTimer(time.Second * 10)
		//go func() {
		//	//цикл с переподелючением
		//	for {
		//		<-timer1.C
		//		db.Close()
		//		log.Print("INFO \t", "Trying to reconnect to the database")
		//		//переподключение к бд
		//		db, err = sql.Open("mysql", "root:password@/go_testsmart_user")
		//		if err != nil {
		//			panic(err)
		//		}
		//
		//		//проверка на успешное подключение
		//		if err = db.Ping(); err != nil {
		//			log.Print("ERROR \t", "Error while connecting to database "+err.Error())
		//			{
		//				timer1 = time.NewTimer(time.Second * 10)
		//				continue
		//			}
		//		} else {
		//			defer db.Close()
		//			log.Print("INFO \t", "Database connection was successful ")
		//
		//			//функция ввода в базу данных
		//			insertToBD(requestobj, db)
		//			return
		//
		//		}
		//	}
		//}()
	} else {
		defer db.Close()
		log.Print("INFO \t", "Database connection was successful")


		//функция ввода в базу данных
		insertToBD(requestobj, db)
	}

}
var errorrequests []reqeststruct //в этой  все запросы
func main() {
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
	reconnect()
	http.HandleFunc("/", test)
	log.Fatal(http.ListenAndServe(":3001", nil))
}
