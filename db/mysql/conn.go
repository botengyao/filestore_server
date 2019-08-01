package mysql

import (
	"database/sql" // 这是一个抽象层包，比如区分mysql、orcal等数据库，只有这个包是连接不上mysql的，还需要搭配下面的mysql包
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" //导入mysql驱动包
)

var db *sql.DB

func init() {
	//mysql数据库,用户名:密码@tcp连接:端口3306/test库?字符集utf8
	db, _ = sql.Open("mysql", "root:Qianxing123@tcp(127.0.0.1:3306)/fileserver?charset=utf8")
	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil {
		fmt.Println("Failed to connect to mysql, err:" + err.Error())
		os.Exit(1)
	}
}

// DBConn : 返回数据库连接对象
func DBConn() *sql.DB {
	return db
}
