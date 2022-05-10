package mysql

import (
	"database/sql"
	"fmt"
	"ginframe/webScaffold/settings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Init(mysqlConfig *settings.MysqlConfig) (err error) {
	err = initDB(mysqlConfig)
	if err != nil {
		return
	}
	// 关闭数据库连接资源
	defer db.Close()

	fmt.Println("数据库链接成功！")
	return
}

func initDB(mysqlConfig *settings.MysqlConfig) (err error) {
	//dsn : Data Source Name 数据资源名，连接库信息
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		mysqlConfig.UserName,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.DbName,
	)
	// 打开数据库，只是验证其参数格式是否正确，实际上并不创建与数据库的连接。
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	// 尝试与数据库建立链接（校验dsn是否正确）
	err = db.Ping()
	if err != nil {
		return err
	}

	// 数据根据业务具体情况设定
	maxLifeTime := time.Second * time.Duration(mysqlConfig.MaxLifetime)
	db.SetConnMaxLifetime(maxLifeTime)            // 最大链接时间
	db.SetMaxOpenConns(mysqlConfig.MaxConns)      // 最大连接数
	db.SetMaxIdleConns(mysqlConfig.MaxIdleConnns) // 最大空闲链接数
	return
}

func Close() {
	_ = db.Close()
}
