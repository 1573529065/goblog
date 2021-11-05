package controllers

import (
	"fmt"
	"goblog/app/models/article"
	"goblog/pkg/logger"
	"goblog/pkg/route"
	"goblog/pkg/types"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"text/template"
	"unicode/utf8"
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

type ArticlesFormData struct {
	Title  string
	Body   string
	URL    string
	Errors map[string]string
}

func (*ArticlesController) Create(w http.ResponseWriter, r *http.Request) {
	storeURL := route.Name2URL("articles.store")
	fmt.Println(storeURL)
	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
		Errors: nil,
	}

	tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}

	tmpl.Execute(w, data)
}

func validateArticleFormData(title string, body string) map[string]string {
	errors := make(map[string]string)

	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度在3-40字节之间"
	}

	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度需大于或等于10个字节"
	}
	return errors
}

func (*ArticlesController) Store(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	body := r.PostFormValue("body")

	errors := validateArticleFormData(title, body)
	if len(errors) == 0 {
		_article := article.Article{
			Title: title,
			Body:  body,
		}
		_article.Create()
		if _article.ID > 0 {
			fmt.Fprintf(w, "插入成功, ID为"+strconv.FormatUint(_article.ID, 10))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "创建文章失败, 请联系管理员")
		}
	} else {
		storeURL := route.Name2URL("articles.store")

		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
		}
		//tmpl, err := template.New("create-form").Parse(html)
		tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
		if err != nil {
			panic(err)
		}

		tmpl.Execute(w, data)
	}
}
