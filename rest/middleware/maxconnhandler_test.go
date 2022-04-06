package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/lang"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

const conns = 4

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestMaxConnMiddleware(t *testing.T) {
	e := echo.New()
	var waitGroup sync.WaitGroup
	waitGroup.Add(conns)
	done := make(chan lang.PlaceholderType)
	defer close(done)

	maxConns := MaxConnMiddleware(conns)
	handler := maxConns(func(context echo.Context) error {
		waitGroup.Done()
		<-done
		return nil
	})

	for i := 0; i < conns; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			err := handler(ctx)
			assert.NoError(t, err)
		}()
	}

	waitGroup.Wait()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
}

func TestWithoutMaxConnsMiddleware(t *testing.T) {
	e := echo.New()
	const (
		key   = "block"
		value = "1"
	)
	var waitGroup sync.WaitGroup
	waitGroup.Add(conns)
	done := make(chan lang.PlaceholderType)
	defer close(done)

	maxConns := MaxConnMiddleware(0)
	handler := maxConns(func(ctx echo.Context) error {
		val := ctx.Request().Header.Get(key)
		if val == value {
			waitGroup.Done()
			<-done
		}
		return nil
	})

	for i := 0; i < conns; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			req.Header.Set(key, value)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			err := handler(ctx)
			assert.NoError(t, err)
		}()
	}

	waitGroup.Wait()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
}
