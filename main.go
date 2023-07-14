package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"sync"

	"context"
	"os"
	"os/signal"
	"time"

	"dev.azure.com/tanla/tanlamvp/_git/common.git/helper/config"
	"dev.azure.com/tanla/tanlamvp/_git/common.git/helper/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type InsertEmployeeRequest struct {
	EmployeeID int    `json:"EID"`
	LastName   string `json:"lastname"`
	FirstName  string `json:"firstname"`
	City       string `json:"city"`
}
type GetEmployeeResponseStruct struct {
	ID         int    `json:"ID"`
	EmployeeID int    `json:"EID"`
	LastName   string `json:"lastname"`
	FirstName  string `json:"firstname"`
	City       string `json:"city"`
}

type GetEmpDetailsRequest struct {
	EmployeeID int `json:"EID"`
}

type DeleteEmpDetailsRequest struct {
	EmployeeID int `json:"EID"`
}

const (
	//define all endpoints
	insertEmpDetailsRoute = "/insertEmpDetails"
	getEmployeebyIDRoute  = "/getEmployee"
	getAllEmployeesRoute  = "/getallemployees"
	deleterecord          = "/deleteemployee"
	loginroute            = "/login"
)

var DB *sql.DB

var wg = sync.WaitGroup{}

func init() {
	//initializing configuration and Database
	config.SetConfiguration("config.json", "json")
	log.Initialize()
	DB = getDB()
}

func main() {

	Handlerfunc()

}

func Handlerfunc() {
	log.Info("starting application")

	r := mux.NewRouter()

	//define Server
	Srv := &http.Server{
		Addr:         ":8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r}

	//initialise routes
	r.HandleFunc(loginroute, Login).Methods("POST")
	r.HandleFunc(insertEmpDetailsRoute, InsertEmpDetails).Methods("POST")
	r.HandleFunc(getEmployeebyIDRoute, GetEmpDetails).Methods("GET")
	r.HandleFunc(getAllEmployeesRoute, GetAllEmployeeDetails).Methods("GET")
	r.HandleFunc(deleterecord, DeleteEmployeeRecord).Methods("DELETE")

	log.Info(fmt.Sprintf("Application is listening at port %s", Srv.Addr))
	// go WaitShutdown(Srv)
	Srv.ListenAndServe()
}

func getDB() *sql.DB {

	//"root:Pass@123@tcp(localhost:3306)/newdb?parseTime=true"
	password := viper.GetString("mysql.Password")
	host := viper.GetString("mysql.Host")
	database := viper.GetString("mysql.Database")
	connstring := fmt.Sprintf("root:%s@tcp(%s)/%s?parseTime=true", password, host, database)

	db, err := sql.Open("mysql", connstring)
	if err != nil {
		panic(err.Error())
	}
	return db
}
func WaitShutdown(srv *http.Server) {

	fmt.Println("Inside WaitShutdown")
	// var wait time.Duration
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	wg.Wait()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Error("%v", err.Error())
	}
	// to finalize based on context cancellation.
	log.Info("shutting down")
	time.Sleep(60 * time.Second)
	os.Exit(0)
	// wg.Done()
}
