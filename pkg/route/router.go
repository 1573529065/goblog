package route

import (
	"github.com/gorilla/mux"
	"net/http"
)

func Name2URL(routeName string, pairs ...string) string {
	var router *mux.Router
	url, err := router.Get(routeName).URL(pairs...)
	if err != nil {
		//checkError(err)
		return ""
	}
	return url.String()
}

// 获取路由参数
func GetRouterVariable(parameterName string, r *http.Request) string {
	vars := mux.Vars(r)
	return vars[parameterName]
}
