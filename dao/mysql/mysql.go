package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/viper"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Init() (err error) {
	err = initDB()
	if err != nil {
		return
	}
	// 关闭数据库连接资源
	defer db.Close()

	fmt.Println("数据库链接成功！")
	return
}

func initDB() (err error) {
	//dsn : Data Source Name 数据资源名，连接库信息
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		viper.GetString("mysql.dbname"),
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
	maxLifeTime := time.Second * time.Duration(viper.GetInt("mysql.max_lifetime"))
	db.SetConnMaxLifetime(maxLifeTime)                      // 最大链接时间
	db.SetMaxOpenConns(viper.GetInt("mysql.max_conns"))     // 最大连接数
	db.SetMaxIdleConns(viper.GetInt("mysql.max_idleconns")) // 最大空闲链接数
	return
}

func Close() {
	_ = db.Close()
}
