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

func (*ArticlesController) Index(w http.ResponseWriter, r *http.Request) {
	articles, err := article.GetAll()

	if err != nil {
		logger.LogError(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "500, 服务器内部错误")
	} else {
		tmpl, err := template.ParseFiles("resources/views/articles/index.gohtml")
		logger.LogError(err)
		tmpl.Execute(w, articles)
	}

	//
	//// 1. 查出所有数据
	//rows, err := db.Query("SELECT * FROM articles")
	//logger.LogError(err)
	//defer rows.Close()
	//
	//var articles []Article
	//// 2. 循环处理数据
	//for rows.Next() {
	//	var article Article
	//	// 2.1 将每一行数据赋值到 artice 对象中
	//	err := rows.Scan(&article.ID, &article.Title, &article.Body)
	//	logger.LogError(err)
	//
	//	// 2.2 将数据追加到 articles 数组中
	//	articles = append(articles, article)
	//}
	//
	//// 2.3 检测遍历时是否发生错误
	//err = rows.Err()
	//logger.LogError(err)
	//
	//// 3. 加载模板
	//temp, err := template.ParseFiles("resources/views/articles/index.gohtml")
	//logger.LogError(err)
	//
	//// 4. 渲染模板
	//temp.Execute(w, articles)
}
