package database

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/oxfeeefeee/appgo"
	"time"
)

func MysqlConnStr() string {
	c := &appgo.Conf.Mysql
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		c.User, c.Password, c.Host, c.Port, c.DbName, c.Charset)
}

func Open(driver string) (*gorm.DB, error) {
	c := &appgo.Conf.Mysql
	switch driver {
	case "mysql":
		gormdb, err := gorm.Open(driver, MysqlConnStr())
		if err != nil {
			return nil, err
		}
		gormdb.DB().SetConnMaxLifetime(time.Second * time.Duration(c.MaxLifetime))
		gormdb.DB().SetMaxOpenConns(c.MaxConn)
		return gormdb, nil
	default:
		return nil, fmt.Errorf("database: unknown driver %q", driver)
	}
}

func SqlStr(str string) sql.NullString {
	return sql.NullString{str, true}
}
