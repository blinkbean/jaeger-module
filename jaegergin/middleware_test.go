package jaegergin

import (
	"fmt"
	jaegerModule "github.com/blinkbean/jaeger-module"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var serviceName = "jaeger_gin"

// BenchmarkJaegerMiddleware//hello/tom-8         	  283958	      5519 ns/op
// BenchmarkJaegerMiddleware//sleep/1ms-8         	    1087	   1146574 ns/op
var benchmarkPaths = []string{"/hello/tom", "/sleep/1ms"}

func BenchmarkJaegerMiddleware(b *testing.B) {
	closer := jaegerModule.InitJaeger(serviceName)
	defer closer.Close()
	for _, v := range benchmarkPaths {
		b.Run(v, func(b *testing.B) {
			benchmarkEngine(b, v, JaegerMiddleware)
		})
	}
}

// BenchmarkWithOutJaegerMiddleware//hello/tom-8         	 3403519	       305 ns/op
// BenchmarkWithOutJaegerMiddleware//sleep/1ms-8         	     967	   1236543 ns/op
func BenchmarkWithOutJaegerMiddleware(b *testing.B) {
	for _, v := range benchmarkPaths {
		b.Run(v, func(b *testing.B) {
			benchmarkEngine(b, v, nil)
		})
	}
}

func benchmarkEngine(b *testing.B, path string, middleware func() gin.HandlerFunc) {
	w := httptest.NewRecorder()
	r := testRouter(middleware)
	req, _ := http.NewRequest("GET", path, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}

func testRouter(middleware func() gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	if middleware != nil {
		r.Use(middleware())
	}
	r.GET("/hello/:name", handleHello)
	r.GET("/sleep/:duration", handleSleep)
	return r
}

func handleHello(c *gin.Context) {
	c.String(http.StatusOK, fmt.Sprintf("Hello, %s", c.Param("name")))
}

func handleSleep(c *gin.Context) {
	d, err := time.ParseDuration(c.Param("duration"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	time.Sleep(d)
}
