package routes

import (
	"eltimn/todo-plus/web/pages"
	"log/slog"
	"net/http"
	"time"
)

func Routes() {
	// serve static files
	fs := http.FileServer(http.Dir("./dist/assets"))
	http.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	todoRoutes()
	// todoRoutes(router.NewGroup("/todo").Use(webErrorHandler))

	http.Handle("GET /hello/{name}", appHandler(helloHandler))
	http.Handle("GET /now", appHandler(nowHandler))
	http.Handle("GET /", appHandler(homeHandler))

	// TODO: Add 404 handler
	// http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
	// 	http.Error(w, "Not Found", http.StatusNotFound)
	// })
}

func helloHandler(w http.ResponseWriter, req *http.Request) *appError {
	name := req.PathValue("name")
	slog.Info("Name", slog.String("name", name))
	// if name == "tim" {
	// 	return &appError{Message: "I'm sorry, I can't do that", Code: http.StatusForbidden}
	// }
	pages.Hello(name).Render(req.Context(), w)
	return nil
}

func homeHandler(w http.ResponseWriter, req *http.Request) *appError {
	pages.HomePage().Render(req.Context(), w)
	return nil
}

func nowHandler(w http.ResponseWriter, req *http.Request) *appError {
	pages.NowPage(time.Now()).Render(req.Context(), w)
	return nil
}
