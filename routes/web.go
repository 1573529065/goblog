package routes

import (
	"github.com/gorilla/mux"
	"goblog/app/http/controllers"
	"net/http"
)

func RegisterWebRoutes(r *mux.Router) {
	// 静态页面
	pc := new(controllers.PagesController)
	// 自定义404 页面
	r.NotFoundHandler = http.HandlerFunc(pc.NotFound)

	r.HandleFunc("/", pc.Home).Methods("GET").Name("home")
	r.HandleFunc("/about", pc.About).Methods("GET").Name("about")

}
