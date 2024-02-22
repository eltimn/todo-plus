package routes

import (
	"eltimn/todo-plus/web/pages"
	"log/slog"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
)

func Routes(router *bunrouter.Router) {

	// serve static files
	fileServer := http.FileServer(http.Dir("dist/assets"))
	fileServer = http.StripPrefix("/assets/", fileServer)
	router.GET("/assets/*path", bunrouter.HTTPHandler(fileServer))

	todoRoutes(router.NewGroup("/todo").Use(webErrorHandler))

	router.GET("/hello/:name", helloHandler)

	router.GET("/now", nowHandler)

	router.GET("/", homeHandler)
}

func homeHandler(w http.ResponseWriter, req bunrouter.Request) error {
	pages.HomePage().Render(req.Context(), w)
	return nil
}

func helloHandler(w http.ResponseWriter, req bunrouter.Request) error {
	params := req.Params()
	name := params.ByName("name")
	slog.Info("Name", slog.String("name", name))
	// if name == "tim" {
	// 	return fmt.Errorf("i am sorry, I can't do that: %d", http.StatusForbidden)
	// }
	pages.Hello(name).Render(req.Context(), w)
	return nil
}

func nowHandler(w http.ResponseWriter, req bunrouter.Request) error {
	pages.NowPage(time.Now()).Render(req.Context(), w)
	return nil
}
