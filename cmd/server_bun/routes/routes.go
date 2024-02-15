package routes

import (
	"eltimn/todo-plus/web/components"
	"log/slog"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
)

func Routes(router *bunrouter.Router) {

	// serve static files
	fileServer := http.FileServer(http.Dir("web/static"))
	// fileServer = http.StripPrefix("/assets/", fileServer)
	router.GET("/assets/*path", bunrouter.HTTPHandler(fileServer))

	todoRoutes(router.NewGroup("/todo").Use(webErrorHandler))

	router.GET("/hello/:name", helloHandler)

	router.GET("/now", func(w http.ResponseWriter, req bunrouter.Request) error {
		components.DisplayTime(time.Now()).Render(req.Context(), w)
		return nil
	})
}

func helloHandler(w http.ResponseWriter, req bunrouter.Request) error {
	params := req.Params()
	name := params.ByName("name")
	slog.Info("Name", slog.String("name", name))
	// if name == "tim" {
	// 	return fmt.Errorf("i am sorry, I can't do that: %d", http.StatusForbidden)
	// }
	components.Hello(name).Render(req.Context(), w)
	return nil
}
