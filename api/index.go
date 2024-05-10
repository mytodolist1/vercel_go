package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mytodolist1/todolist_be/config"
	. "github.com/mytodolist1/todolist_be/handler"
	"github.com/mytodolist1/todolist_be/model"
)

var (
	datauser model.User
	// datatodo      model.Todo
	// datatodoclear model.TodoClear
	// responseData  bson.M
)

var mconn = config.MongoConnect("MONGOSTRING", "mytodolist")

func Homes(w http.ResponseWriter, r *http.Request) {
	StatusOK(w, "Welcome to MyTodoList API")
}

func init() {
	router := mux.NewRouter()
	router.Use(config.CorsMiddleware)
	router.HandleFunc("/", Homes).Methods("GET")

	http.ListenAndServe(":8080", router)
}
