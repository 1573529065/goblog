package controllers

import (
	"fmt"
	"goblog/app/models/article"
	"goblog/pkg/logger"
	"goblog/pkg/route"
	"goblog/pkg/types"
	"gorm.io/gorm"
	"net/http"
	"text/template"
)

// PagesController 处理静态页面
type ArticlesController struct {
}

// 文章详情页
func (*ArticlesController) Show(w http.ResponseWriter, r *http.Request) {
	// 1. 获取参数
	id := route.GetRouterVariable("id", r)

	// 2. 读取数据库数据
	article, err := article.Get(id)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 文章未找到")
		} else {
			// 3.2 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		// 4. 读取成功
		//fmt.Fprintf(w, "读取成功, 文章标题 - %v", article.Title)
		//tmpl, err := template.ParseFiles("resources/views/articles/show.gohtml")
		tmpl, err := template.New("show.gohtml").
			Funcs(template.FuncMap{
				"RouteName2URL":  route.Name2URL,
				"Uint64ToString": types.Uint64ToString,
			}).ParseFiles("resources/views/articles/show.gohtml")
		logger.LogError(err)

		tmpl.Execute(w, article)
	}
}
