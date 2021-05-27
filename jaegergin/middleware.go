package jaegergin

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/http"
)

func JaegerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !opentracing.IsGlobalTracerRegistered() {
			c.Next()
			return
		}
		tr := opentracing.GlobalTracer()
		ctx, _ := tr.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		op := "HTTP " + c.Request.URL.Path
		sp := tr.StartSpan(op, ext.RPCServerOption(ctx))
		defer sp.Finish()
		ext.HTTPMethod.Set(sp, c.Request.Method)
		ext.HTTPUrl.Set(sp, c.Request.URL.String())

		c.Request = c.Request.WithContext(
			opentracing.ContextWithSpan(c.Request.Context(), sp))

		c.Next()

		ext.HTTPStatusCode.Set(sp, uint16(c.Writer.Status()))
		if c.Writer.Status() > http.StatusInternalServerError {
			ext.Error.Set(sp, true)
		}
		if err := tr.Inject(sp.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header)); err != nil {
			ext.LogError(sp, err)
		}
	}
}
