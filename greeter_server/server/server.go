package server

import (
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/weave-lab/helloworld_grpctooling_poc/greeter_server/middleware"
	"google.golang.org/grpc"
)

var defaultUnaryMiddleware = []grpc.UnaryServerInterceptor{middleware.UnaryLogging, middleware.UnaryMetrics}
var defaultStreamingMiddleware = []grpc.StreamServerInterceptor{middleware.StreamLogging}

func New(unaryMiddleWare []grpc.UnaryServerInterceptor, streamMiddleware []grpc.StreamServerInterceptor) *grpc.Server {

	//Add list of passed in middlewares to defaults
	unaryMiddleWare = append(unaryMiddleWare, defaultUnaryMiddleware...)
	streamMiddleware = append(streamMiddleware, defaultStreamingMiddleware...)

	//grpc_middleware has to be used because grpc.Server actually only allows one interceptor
	s := grpc.NewServer(grpc_middleware.WithUnaryServerChain(unaryMiddleWare...), grpc_middleware.WithStreamServerChain(streamMiddleware...))

	return s
}
