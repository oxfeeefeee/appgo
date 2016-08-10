package database

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/oxfeeefeee/appgo"
	"time"
)

func Open(driver string) (*gorm.DB, error) {
	return OpenWithIndex(driver, 0)
}

func MysqlConnStr(index int) (string, int, int) {
	if index < 0 || index >= len(appgo.Conf.Mysql) {
		return "", 0, 0
	}
	c := &appgo.Conf.Mysql[index]
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		c.User, c.Password, c.Host, c.Port, c.DbName, c.Charset), c.MaxLifetime, c.MaxConn
}

func OpenWithIndex(driver string, index int) (*gorm.DB, error) {
	switch driver {
	case "mysql":
		connstr, lt, mc := MysqlConnStr(index)
		gormdb, err := gorm.Open(driver, connstr)
		if err != nil {
			return nil, err
		}
		gormdb.DB().SetConnMaxLifetime(time.Second * time.Duration(lt))
		gormdb.DB().SetMaxOpenConns(mc)
		return gormdb, nil
	default:
		return nil, fmt.Errorf("database: unknown driver %q", driver)
	}
}

func SqlStr(str string) sql.NullString {
	return sql.NullString{str, true}
}
