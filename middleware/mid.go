package middleware

import (
	"fmt"
	"net/http"

	"eltimn/todo-plus/pkg/router"
)

func Mid(i int) router.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("mid", i, "start")
			next.ServeHTTP(w, r)
			fmt.Println("mid", i, "done")
		})
	}
}
