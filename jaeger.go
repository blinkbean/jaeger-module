package jaeger_module

import (
	"github.com/opentracing/opentracing-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	"io"
)

const HOSTPORT = "127.0.0.1:6831"
const DATASOURCE = "root:root@tcp(127.0.0.1:3306)/jaeger?parseTime=true&loc=Local&charset=utf8mb4"


func InitJaeger(serviceName string) io.Closer {
	cfg := &jaegerConfig.Configuration {
		Sampler: &jaegerConfig.SamplerConfig{
			Type  : "const", //固定采样
			Param : 1,       //1=全采样、0=不采样
		},

		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans           : true,
			LocalAgentHostPort : HOSTPORT, // 先写死
		},

		ServiceName: serviceName,
	}
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil
	}
	if tracer != nil {
		opentracing.SetGlobalTracer(tracer)
	}
	return closer
}

