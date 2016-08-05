package middleware

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//Wraps the server stream that is used for sending/receiving messages
type wrappedServerStream struct {
	grpc.ServerStream
	WrappedContext    context.Context
	recvMsgDispatch   StreamHandler
	sendMsgDispatch   StreamHandler
	recvMsgMiddleware []func(StreamHandler) StreamHandler
	sendMsgMiddleware []func(StreamHandler) StreamHandler
}

// WrapServerStream returns a ServerStream that has the ability to overwrite context.
func wrapServerStream(stream grpc.ServerStream) *wrappedServerStream {
	if existing, ok := stream.(*wrappedServerStream); ok {
		return existing
	}
	return &wrappedServerStream{
		ServerStream:   stream,
		WrappedContext: stream.Context(),
	}
}

func (w *wrappedServerStream) RegisterRecvMiddleware(middleware func(StreamHandler) StreamHandler) {
	w.recvMsgMiddleware = append(w.recvMsgMiddleware, func(handler StreamHandler) StreamHandler {
		return outerBridge{middleware, handler}
	})

	w.recvMsgDispatch = buildChain(w.ServerStream.RecvMsg, w.recvMsgMiddleware)
}

func (w *wrappedServerStream) RegisterSendMiddleware(middleware func(StreamHandler) StreamHandler) {
	w.sendMsgMiddleware = append(w.sendMsgMiddleware, func(handler StreamHandler) StreamHandler {
		return outerBridge{middleware, handler}
	})

	w.sendMsgDispatch = buildChain(w.ServerStream.SendMsg, w.sendMsgMiddleware)
}

//builds the stream middleware chain
func buildChain(root func(interface{}) error, middleware []func(StreamHandler) StreamHandler) StreamHandler {
	var handler StreamHandler
	handler = StreamFunc(root)

	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}

	return handler
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *wrappedServerStream) Context() context.Context {
	return w.WrappedContext
}

//SendMsg calls SendMsg on the underlying grpc.ServerStream but allows for middleware
func (w *wrappedServerStream) SendMsg(m interface{}) error {
	return w.sendMsgDispatch.Stream(m)
}

//RecvMsg calls RecvMsg on the underlying grpc.ServerStream but allows for middleware
func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	return w.recvMsgDispatch.Stream(m)
}

//Middleware BS

//StreamFunc implements stream
type StreamFunc func(m interface{}) error

func (s StreamFunc) Stream(m interface{}) error {
	return s(m)
}

//StreamHandler can stream data
type StreamHandler interface {
	Stream(m interface{}) error
}

//outerBridge that calls an inner StreamHandler
type outerBridge struct {
	mware func(StreamHandler) StreamHandler
	inner StreamHandler
}

func (b outerBridge) Stream(m interface{}) error {
	return b.mware(innerBridge{b.inner}).Stream(m)
}

//innerBridge that executes a stream
type innerBridge struct {
	inner StreamHandler
}

func (b innerBridge) Stream(m interface{}) error {
	return b.inner.Stream(m)
}
