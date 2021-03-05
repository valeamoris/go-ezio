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

	DBOption func()
)

func NewMysqlResolver(readSources []string) *dbresolver.DBResolver {
	resolver := dbresolver.Register(dbresolver.Config{
		Replicas: reduceReadSource(readSources),
		Policy:   dbresolver.RandomPolicy{},
	})
	return resolver
}

func NewMysql(datasource string, c *Config) (*DB, error) {
	conn, err := gorm.Open(mysql.Open(datasource), c)
	if err != nil {
		return nil, err
	}
	// 引入breaker插件
	err = conn.Use(NewBreakerPlugin())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func reduceReadSource(readSources []string) []gorm.Dialector {
	dialectors := make([]gorm.Dialector, 0)
	for _, readSource := range readSources {
		dialectors = append(dialectors, mysql.Open(readSource))
	}
	return dialectors
}
