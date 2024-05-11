package api

import (
	"net/http"
	"os"

	"github.com/mytodolist1/todolist_be/config"
	h "github.com/mytodolist1/todolist_be/handler"
	"github.com/mytodolist1/todolist_be/model"
	"github.com/mytodolist1/todolist_be/modul"
	"github.com/mytodolist1/todolist_be/paseto"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	datauser      model.User
	datatodo      model.Todo
	datatodoclear model.TodoClear
)

var mconn = config.MongoConnect("MONGOSTRING", "mytodolist")

func Handler(w http.ResponseWriter, r *http.Request) {
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	handler := corsMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			if r.Method == "GET" {
				h.StatusOK(w, "Welcome to My To Do List API")
				return
			}

		case "/login":
			if r.Method == "POST" {
				if r.ContentLength == 0 {
					h.StatusMethodNotAllowed(w, "Request body is empty")
					return
				}
				err := h.JDecoder(w, r, &datauser)
				if err != nil {
					h.StatusBadRequest(w, "error parsing application/json: "+err.Error())
					return
				}
				user, err := modul.LogIn(mconn, "user", datauser)
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				tokenstring, err := paseto.Encode(user.ID.Hex(), user.Role, os.Getenv("PRIVATE_KEY"))
				if err != nil {
					h.StatusBadRequest(w, "Gagal Encode Token : "+err.Error())
					return
				}
				h.StatusOK(w, "User "+user.Username+" has been logged in", "token", tokenstring, "data", user)
				return
			}

		case "/register":
			if r.Method == "POST" {
				err := h.JDecoder(w, r, &datauser)
				if err != nil {
					h.StatusBadRequest(w, "error parsing application/json: "+err.Error())
					return
				}
				err = modul.Register(mconn, "user", datauser)
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				h.StatusCreated(w, "User "+datauser.Username+" has been created")
				return
			}

		case "/user":
			username := r.URL.Query().Get("username")
			id := r.URL.Query().Get("_id")

			switch r.Method {
			case "DELETE":
				_, err := h.PasetoDecode(w, r, "Authorization")
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				if username == "" {
					h.StatusBadRequest(w, "Missing 'username' parameter in the URL")
					return
				}
				datauser.Username = username
				err = h.JDecoder(w, r, &datauser)
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				status, err := modul.DeleteUser(mconn, "user", username)
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				if !status {
					h.StatusConflict(w, "User "+username+" cannot be deleted because it is already deleted or does not exist")
					return
				}
				h.StatusNoContent(w, "User "+username+" has been deleted")
				return

			case "PUT":
				_, err := h.PasetoDecode(w, r, "Authorization")
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				if id != "" {
					if id == "" {
						h.StatusBadRequest(w, "Missing '_id' parameter in the URL")
						return
					}
					ID, err := primitive.ObjectIDFromHex(id)
					if err != nil {
						h.StatusBadRequest(w, "Invalid '_id' parameter in the URL")
						return
					}
					datauser.ID = ID
					err = h.JDecoder(w, r, &datauser)
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					user, _, err := modul.UpdateUser(mconn, "user", datauser)
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					h.StatusOK(w, "User "+user.Username+" has been updated")
					return

				} else if username != "" {
					if username == "" {
						h.StatusBadRequest(w, "Missing 'username' parameter in the URL")
						return
					}
					datauser.Username = username
					err = h.JDecoder(w, r, &datauser)
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					user, err := modul.ChangePassword(mconn, "user", datauser)
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					h.StatusOK(w, "User "+user.Username+" has been updated")
					return
				}

			case "GET":
				header := r.Header.Get("AuthorizationA")
				if header != "" {
					payload, err := h.PasetoDecode(w, r, "AuthorizationA")
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					if payload.Role == "admin" {
						users, err := modul.GetUserFromRole(mconn, "user", "user")
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "All User has been found", "data", users)
						return

					} else {
						h.StatusUnauthorized(w, "You are not authorized to access this data")
						return
					}

				} else {
					if username != "" {
						if username == "" {
							h.StatusBadRequest(w, "Missing 'username' parameter in the URL")
							return
						}
						datauser.Username = username
						user, err := modul.GetUserFromUsername(mconn, "user", username)
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "User "+user.Username+" has been found", "data", user)
						return

					} else if id != "" {
						if id == "" {
							h.StatusBadRequest(w, "Missing '_id' parameter in the URL")
							return
						}
						ID, err := primitive.ObjectIDFromHex(id)
						if err != nil {
							h.StatusBadRequest(w, "Invalid '_id' parameter in the URL")
							return
						}
						datauser.ID = ID
						user, err := modul.GetUserFromID(mconn, "user", ID)
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "User "+user.Username+" has been found", "data", user)
						return

					} else {
						payload, err := h.PasetoDecode(w, r, "Authorization")
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						user, err := modul.GetUserFromID(mconn, "user", payload.Id)
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "User "+user.Username+" has been found", "data", user)
						return
					}
				}

			default:
				h.StatusMethodNotAllowed(w, "Method not allowed")
				return
			}

		case "/todo":
			id := r.URL.Query().Get("_id")
			category := r.URL.Query().Get("category")

			switch r.Method {
			case "DELETE":
				_, err := h.PasetoDecode(w, r, "Authorization")
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				if id == "" {
					h.StatusBadRequest(w, "Missing '_id' parameter in the URL")
					return
				}
				ID, err := primitive.ObjectIDFromHex(id)
				if err != nil {
					h.StatusBadRequest(w, "Invalid '_id' parameter in the URL")
					return
				}
				datatodo.ID = ID
				status, err := modul.DeleteTodo(mconn, "todo", ID)
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				if !status {
					h.StatusConflict(w, "Todo cannot be deleted because it is already deleted or does not exist")
					return
				}
				h.StatusNoContent(w, "Todo has been deleted")
				return

			case "PUT":
				_, err := h.PasetoDecode(w, r, "Authorization")
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				if id == "" {
					h.StatusBadRequest(w, "Missing '_id' parameter in the URL")
					return
				}
				ID, err := primitive.ObjectIDFromHex(id)
				if err != nil {
					h.StatusBadRequest(w, "Invalid '_id' parameter in the URL")
					return
				}
				datatodo.ID = ID
				_, _, err = modul.UpdateTodo(mconn, "todo", ID, r)
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				h.StatusOK(w, "Todo has been updated")
				return

			case "POST":
				payload, err := h.PasetoDecode(w, r, "Authorization")
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				_, err = modul.InsertTodo(mconn, "todo", payload.Id, r)
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				h.StatusCreated(w, "Todo has been created")
				return

			case "GET":
				header := r.Header.Get("AuthorizationA")
				if header != "" {
					payload, err := h.PasetoDecode(w, r, "AuthorizationA")
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					if payload.Role == "admin" {
						todos, err := modul.GetTodoList(mconn, "todo")
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "All Todo has been found", "data", todos)
						return

					} else {
						h.StatusUnauthorized(w, "You are not authorized to access this data")
						return
					}

				} else {
					if category != "" {
						if category == "" {
							h.StatusBadRequest(w, "Missing 'category' parameter in the URL")
							return
						}
						datatodo.Tags.Category = category
						todos, err := modul.GetTodoFromCategory(mconn, "todo", category)
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "Todo has been found", "data", todos)
						return

					} else if id != "" {
						if id == "" {
							h.StatusBadRequest(w, "Missing '_id' parameter in the URL")
							return
						}
						ID, err := primitive.ObjectIDFromHex(id)
						if err != nil {
							h.StatusBadRequest(w, "Invalid '_id' parameter in the URL")
							return
						}
						datatodo.ID = ID
						todos, err := modul.GetTodoFromID(mconn, "todo", ID)
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "Todo has been found", "data", todos)
						return

					} else {
						payload, err := h.PasetoDecode(w, r, "Authorization")
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						todos, err := modul.GetTodoFromIDUser(mconn, "todo", payload.Id)
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "Todo has been found", "data", todos)
						return
					}
				}

			default:
				h.StatusMethodNotAllowed(w, "Method not allowed")
				return
			}

		case "/todo/category":
			if r.Method == "GET" {
				header := r.Header.Get("AuthorizationA")
				if header != "" {
					payload, err := h.PasetoDecode(w, r, "AuthorizationA")
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					if payload.Role == "admin" {
						categories, err := modul.GetCategory(mconn, "category")
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "All Category has been found", "data", categories)
						return
					}

				} else {
					h.StatusUnauthorized(w, "You are not authorized to access this data")
					return
				}
			}

		case "/todo/clear":
			id := r.URL.Query().Get("_id")
			header := r.Header.Get("AuthorizationA")

			switch r.Method {
			case "POST":
				_, err := h.PasetoDecode(w, r, "Authorization")
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				if id == "" {
					h.StatusBadRequest(w, "Missing '_id' parameter in the URL")
					return
				}
				ID, err := primitive.ObjectIDFromHex(id)
				if err != nil {
					h.StatusBadRequest(w, "Invalid '_id' parameter in the URL")
					return
				}
				datatodoclear.Todo.ID = ID
				status, err := modul.TodoClear(mconn, "todoclear", ID)
				if err != nil {
					h.StatusBadRequest(w, err.Error())
					return
				}
				if !status {
					h.StatusConflict(w, "Todo cannot be cleared because it is already cleared or does not exist")
					return
				}
				h.StatusCreated(w, "Todo has been cleared")
				return

			case "GET":
				if header != "" {
					payload, err := h.PasetoDecode(w, r, "AuthorizationA")
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					if payload.Role == "admin" {
						todos, err := modul.GetTodoClear(mconn, "todoclear")
						if err != nil {
							h.StatusBadRequest(w, err.Error())
							return
						}
						h.StatusOK(w, "All Todo has been found", "data", todos)
						return

					} else {
						h.StatusUnauthorized(w, "You are not authorized to access this data")
						return
					}

				} else {
					payload, err := h.PasetoDecode(w, r, "Authorization")
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					todos, err := modul.GetTodoClearFromIDUser(mconn, "todoclear", payload.Id)
					if err != nil {
						h.StatusBadRequest(w, err.Error())
						return
					}
					h.StatusOK(w, "Todo has been found", "data", todos)
					return
				}

			default:
				h.StatusMethodNotAllowed(w, "Method not allowed")
				return
			}

		default:
			h.StatusNotFound(w, "Route not found")
			return
		}
	}))

	handler.ServeHTTP(w, r)
}
