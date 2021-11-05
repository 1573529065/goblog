package article

import (
	"goblog/pkg/logger"
	"goblog/pkg/model"
	"goblog/pkg/types"
)

// 通过id 获取文章信息
func Get(idstr string) (Article, error) {
	var article Article
	id := types.StringToUint64(idstr)

	if err := model.DB.First(&article, id).Error; err != nil {
		return article, err
	}
	return article, nil
}

func GetAll() ([]Article, error) {
	var articles []Article
	if err := model.DB.Find(&articles).Error; err != nil {
		return articles, err
	}
	return articles, nil
}

func (article *Article) Create() (err error) {
	if err = model.DB.Create(&article).Error; err != nil {
		logger.LogError(err)
		return err
	}
	return nil

	//// 变量初始化
	//var (
	//	id   int64
	//	err  error
	//	rs   sql.Result
	//	stmt *sql.Stmt
	//)
	//
	//stmt, err = db.Prepare("INSERT INTO articles (title, body) values (?, ?)")
	//if err != nil {
	//	return 0, err
	//}
	//
	//defer stmt.Close()
	//
	//rs, err = stmt.Exec(title, body)
	//if err != nil {
	//	return 0, err
	//}
	//
	//if id, err = rs.LastInsertId(); id > 0 {
	//	return id, err
	//}
	//
	//return 0, err
}
