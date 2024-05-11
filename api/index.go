package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mytodolist1/todolist_be/modul"
	"github.com/mytodolist1/todolist_be/paseto"
	. "github.com/tbxark/g4vercel"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	server := New()
	server.Use(Recovery(func(err interface{}, c *Context) {
		if httpError, ok := err.(HttpError); ok {
			c.JSON(httpError.Status, H{
				"message": httpError.Error(),
			})
		} else {
			message := fmt.Sprintf("%s", err)
			c.JSON(500, H{
				"message": message,
			})
		}
	}))

	server.GET("/", func(context *Context) {
		context.JSON(200, H{
			"message": "Hello World!",
		})
	})

	server.POST("/login", func(c *Context) {
		r := c.Req

		if r.ContentLength == 0 {
			c.JSON(400, H{
				"message": "Request body is empty",
			})
			return
		}

		err := json.NewDecoder(r.Body).Decode(&datauser)
		if err != nil {
			c.JSON(400, H{
				"message": "error parsing application/json: " + err.Error(),
			})
			return
		}

		user, err := modul.LogIn(mconn, "user", datauser)
		if err != nil {
			c.JSON(400, H{
				"message": err.Error(),
			})
			return
		}

		tokenstring, err := paseto.Encode(user.ID.Hex(), user.Role, os.Getenv("PRIVATE_KEY"))
		if err != nil {
			c.JSON(400, H{
				"message": "Gagal Encode Token : " + err.Error(),
			})
			return
		}

		c.JSON(200, H{
			"message": "User " + user.Username + " has been logged in",
			"token":   tokenstring,
			"data":    user,
		})
	})

	server.Handle(w, r)
}
