package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"goblog/pkg/logger"
	"goblog/pkg/route"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"
)

var router *mux.Router
var db *sql.DB

type Article struct {
	ID    int64
	Title string
	Body  string
}

type articlesFormData struct {
	Title  string
	Body   string
	URL    *url.URL
	Errors map[string]string
}

func initDB() {
	var err error
	config := mysql.Config{
		User:                 "root",
		Passwd:               "123456",
		Addr:                 "192.168.1.133",
		Net:                  "tcp",
		DBName:               "goblog",
		AllowNativePasswords: true,
	}
	fmt.Println(config.FormatDSN())

	//准备连接数据库数据池
	db, err = sql.Open("mysql", config.FormatDSN())
	logger.LogError(err)

	// 设置最大连接数
	db.SetMaxOpenConns(25)

	// 设置最大空闲连接数
	db.SetMaxIdleConns(25)
	// 设置每个连接的过期时间
	db.SetConnMaxLifetime(5 * time.Minute)

	// 尝试连接, 失败会报错
	err = db.Ping()
	logger.LogError(err)

}

func createTables() {
	createArticlesSQL := `CREATE TABLE IF NOT EXISTS articles(
	id bigint(20) PRIMARY KEY AUTO_INCREMENT NOT NULL,
	title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
	body longtext COLLATE utf8mb4_unicode_ci
	);
	`
	_, err := db.Exec(createArticlesSQL)
	logger.LogError(err)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello, 这里是 goblog</h1>")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "此博客是用以记录1编程笔记,如您有反馈或建议, 请联系"+
		"<a href=\"mailto:summer@example.com\">summer@example.com</a>")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页面未找到: (</h1><p>如有疑惑, 请联系我们.</p>")
}

func articlesShowHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取参数
	id := route.GetRouterVariable("id", r)

	// 2. 读取数据库数据
	article, err := getArticelByID(id)

	if err != nil {
		if err == sql.ErrNoRows {
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
		tmpl, err := template.New("show.gohtml").Funcs(template.FuncMap{
			"RouteName2URL": route.Name2URL,
			"Int64ToString": Int64ToString,
		}).ParseFiles("resources/views/articles/show.gohtml")
		logger.LogError(err)

		tmpl.Execute(w, article)
	}
}

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 查出所有数据
	rows, err := db.Query("SELECT * FROM articles")
	logger.LogError(err)
	defer rows.Close()

	var articles []Article
	// 2. 循环处理数据
	for rows.Next() {
		var article Article
		// 2.1 将每一行数据赋值到 artice 对象中
		err := rows.Scan(&article.ID, &article.Title, &article.Body)
		logger.LogError(err)

		// 2.2 将数据追加到 articles 数组中
		articles = append(articles, article)
	}

	// 2.3 检测遍历时是否发生错误
	err = rows.Err()
	logger.LogError(err)

	// 3. 加载模板
	temp, err := template.ParseFiles("resources/views/articles/index.gohtml")
	logger.LogError(err)

	// 4. 渲染模板
	temp.Execute(w, articles)
}

func (a Article) Link() string {
	showURL, err := router.Get("articles.show").URL("id", strconv.FormatInt(a.ID, 10))
	if err != nil {
		logger.LogError(err)
		return ""
	}
	return showURL.String()
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

func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	body := r.PostFormValue("body")

	errors := validateArticleFormData(title, body)
	if len(errors) == 0 {
		lastInsertID, err := saveArticleToDB(title, body)
		if lastInsertID > 0 {
			fmt.Fprintf(w, "插入成功, ID为"+strconv.FormatInt(lastInsertID, 10))
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500, 服务器内部错误")
		}
	} else {
		storeURL, _ := router.Get("articles.store").URL()

		data := articlesFormData{
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

func saveArticleToDB(title string, body string) (int64, error) {
	// 变量初始化
	var (
		id   int64
		err  error
		rs   sql.Result
		stmt *sql.Stmt
	)

	stmt, err = db.Prepare("INSERT INTO articles (title, body) values (?, ?)")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	rs, err = stmt.Exec(title, body)
	if err != nil {
		return 0, err
	}

	if id, err = rs.LastInsertId(); id > 0 {
		return id, err
	}

	return 0, err
}

func articlesCreateHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}

	storeURL, _ := router.Get("articles.store").URL()
	data := articlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
		Errors: nil,
	}
	tmpl.Execute(w, data)
}

// 通过id 获取文章信息
func getArticelByID(id string) (Article, error) {
	article := Article{}
	query := "SELECT * FROM articles WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
	return article, err
}

func articlesEditHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取参数
	id := route.GetRouterVariable("id", r)

	// 2. 读取数据
	article, err := getArticelByID(id)

	if err != nil {
		if err == sql.ErrNoRows {
			// 3.1 找不到数据
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 数据未找到")
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器错误")
		}
	} else {
		// 4. 读取成功
		updateURL, _ := router.Get("articles.update").URL("id", id)
		data := articlesFormData{
			Title:  article.Title,
			Body:   article.Body,
			URL:    updateURL,
			Errors: nil,
		}

		temp, err := template.ParseFiles("resources/views/articles/edit.gohtml")
		logger.LogError(err)

		temp.Execute(w, data)
	}

}

func articlesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取参数
	id := route.GetRouterVariable("id", r)

	// 2. 读取文章
	_, err := getArticelByID(id)
	if err != nil {
		// 3. 错误处理
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "数据不存在")
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "服务器错误")
		}
	} else {
		// 4. 未出现错误, 表单验证

		title := r.PostFormValue("title")
		body := r.PostFormValue("body")

		errors := validateArticleFormData(title, body)
		if len(errors) == 0 {
			// 验证通过
			query := "UPDATE articles SET title = ?, body = ? WHERE id = ?"
			rs, err := db.Exec(query, title, body, id)
			if err != nil {
				logger.LogError(err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "服务器内容错误")
			}

			if n, _ := rs.RowsAffected(); n > 0 {
				showURL, _ := router.Get("articles.show").URL("id", id)
				http.Redirect(w, r, showURL.String(), http.StatusFound)
			} else {
				fmt.Fprint(w, "没有任何更改")
			}
		} else {
			// 验证未通过
			updateUrl, _ := router.Get("articles.update").URL("id", id)
			data := articlesFormData{
				Title:  title,
				Body:   body,
				URL:    updateUrl,
				Errors: errors,
			}
			tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
			logger.LogError(err)
			tmpl.Execute(w, data)
		}
	}
}

func Int64ToString(num int64) string {
	return strconv.FormatInt(num, 10)
}

// Delete 方法用以从数据库中删除单条记录
func (a Article) Delete() (rowsAffected int64, err error) {
	rs, err := db.Exec("DELETE FROM articles WHERE id = " + strconv.FormatInt(a.ID, 10))
	if err != nil {
		return 0, err
	}

	// √ 删除成功，跳转到文章详情页
	if n, _ := rs.RowsAffected(); n > 0 {
		return n, nil
	}

	return 0, nil
}

func articlesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := route.GetRouterVariable("id", r)

	article, err := getArticelByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
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
		// 4. 没有报错. 删除
		rowsAffected, err := article.Delete()

		if err != nil {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "服务器错误")
		} else {
			if rowsAffected > 0 {
				indexURL, _ := router.Get("articles.index").URL()
				http.Redirect(w, r, indexURL.String(), http.StatusFound)
			} else {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, " 文章未找到")
			}
		}
	}
}

func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 设置表头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 2. 继续处理请求
		next.ServeHTTP(w, r)
	})
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	initDB()
	createTables()

	route.Initialize()
	router = route.Router
	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHandler).Methods("GET").Name("about")

	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")
	router.HandleFunc("/articles/{id:[0-9]+}/edit", articlesEditHandler).Methods("GET").Name("articles.edit")
	router.HandleFunc("/articles/{id:[0-9]+}", articlesUpdateHandler).Methods("POST").Name("articles.update")
	router.HandleFunc("/articles/{id:[0-9]+}/delete", articlesDeleteHandler).Methods("POST").Name("articles.delete")

	// 自定义404 页面
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	// 中间件 强制内容类型为html
	router.Use(forceHTMLMiddleware)

	http.ListenAndServe(":3000", removeTrailingSlash(router))
}
