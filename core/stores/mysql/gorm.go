package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type (
	SqlOption func()

	DB        = gorm.DB
	Config    = gorm.Config
	Model     = gorm.Model
	DeletedAt = gorm.DeletedAt
)

func NewMysql(datasource string, readSources []string, c *Config) (*DB, error) {
	conn, err := gorm.Open(mysql.Open(datasource), c)
	if err != nil {
		return nil, err
	}
	// 引入读写分离插件
	if len(readSources) > 0 {
		err = conn.Use(dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(datasource)},
			Replicas: reduceReadSource(readSources),
			Policy:   dbresolver.RandomPolicy{},
		}))
		if err != nil {
			return nil, err
		}
	}
	// 引入breaker插件
	conn.Use(NewBreakerPlugin())
	return conn, nil
}

func reduceReadSource(readSources []string) []gorm.Dialector {
	dialectors := make([]gorm.Dialector, 0)
	for _, readSource := range readSources {
		dialectors = append(dialectors, mysql.Open(readSource))
	}
	return dialectors
}
