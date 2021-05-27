```go
    import "github.com/grpc-ecosystem/go-grpc-middleware"
    
    myServer := grpc.NewServer(
    grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
    grpc_recovery.StreamServerInterceptor(),
    grpc_ctxtags.StreamServerInterceptor(),
    grpc_opentracing.StreamServerInterceptor(),
    grpc_prometheus.StreamServerInterceptor,
    grpc_zap.StreamServerInterceptor(zapLogger),
    grpc_auth.StreamServerInterceptor(myAuthFunction),
    )),
    grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
    grpc_recovery.UnaryServerInterceptor(),
    grpc_ctxtags.UnaryServerInterceptor(),
    grpc_opentracing.UnaryServerInterceptor(),
    grpc_prometheus.UnaryServerInterceptor,
    grpc_zap.UnaryServerInterceptor(zapLogger),
    grpc_auth.UnaryServerInterceptor(myAuthFunction),
    )),
    )
```