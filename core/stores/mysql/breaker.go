package mysql

import (
	"github.com/go-sql-driver/mysql"
	"github.com/tal-tech/go-zero/core/breaker"
	"gorm.io/gorm"
)

const (
	GormContextKey            = "gorm:context_key"
	duplicateEntryCode uint16 = 1062
)

type breakerPlugin struct {
	*gorm.DB
	brk breaker.Breaker
}

func NewBreakerPlugin() *breakerPlugin {
	return &breakerPlugin{}
}

func (b *breakerPlugin) Name() string {
	return "gorm:db_breaker"
}

func (b *breakerPlugin) Initialize(db *gorm.DB) error {
	b.DB = db
	b.brk = breaker.NewBreaker()
	b.registerCallbacks()
	return nil
}

func (b *breakerPlugin) registerCallbacks() {
	b.Callback().Create().Before("*").Register("gorm:db_breaker:create:before", b.before)
	b.Callback().Create().After("*").Register("grom:db_breaker:create:after", b.after)
	b.Callback().Query().Before("*").Register("gorm:db_breaker:query:before", b.before)
	b.Callback().Query().After("*").Register("gorm:db_breaker:query:after", b.after)
	b.Callback().Update().Before("*").Register("gorm:db_breaker:update:before", b.before)
	b.Callback().Delete().Before("*").Register("gorm:db_breaker:update:before", b.after)
	b.Callback().Row().Before("*").Register("gorm:db_breaker:row:before", b.before)
	b.Callback().Row().After("*").Register("gorm:db_breaker:row:after", b.after)
	b.Callback().Raw().Before("*").Register("gorm:db_breaker:raw:before", b.before)
	b.Callback().Raw().After("*").Register("gorm:db_breaker:raw:after", b.after)
}

func (b *breakerPlugin) before(db *gorm.DB) {
	promise, err := b.brk.Allow()
	if err != nil {
		// 如果已经有error后续就不执行了
		db.AddError(err)
		return
	}
	db.Set(GormContextKey, promise)
}

func (b *breakerPlugin) after(db *gorm.DB) {
	i, ok := db.Get(GormContextKey)
	if !ok {
		return
	}
	promise := i.(breaker.Promise)
	if db.Error != nil {
		ok := db.Error == gorm.ErrRecordNotFound
		if ok || mysqlAcceptable(db.Error) {
			promise.Accept()
		} else {
			promise.Reject(db.Error.Error())
		}
	} else {
		promise.Accept()
	}
}

func mysqlAcceptable(err error) bool {
	if err == nil {
		return true
	}

	myerr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}

	switch myerr.Number {
	case duplicateEntryCode:
		return true
	default:
		return false
	}
}
