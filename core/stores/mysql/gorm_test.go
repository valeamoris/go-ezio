package mysql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMysql(t *testing.T) {
	_, err := NewMysql("root:aa123456@tcp(localhost:3306)/goezio", &Config{})
	assert.NoError(t, err)
}

func TestNewMysqlWithResolver(t *testing.T) {
	db, err := NewMysql("root:aa123456@tcp(localhost:3306)/goezio", &Config{})
	assert.NoError(t, err)

	resolver := NewMysqlResolver([]string{"root:aa123456@tcp(localhost:3306)/goezio"})
	err = db.Use(resolver)
	assert.NoError(t, err)
}
