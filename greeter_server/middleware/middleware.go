package middleware

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//UnaryLogging for handling logging for unary gRPC endpoints
func UnaryLogging(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	fmt.Println("Log: ", info.FullMethod, "called by", ctx.Value("user"))

	resp, err = handler(ctx, req)

	fmt.Println("Log:", info.FullMethod, "completed.")

	return resp, err
}

//UnaryMetrics for handling metrics for unary gRPC endpoints
func UnaryMetrics(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	now := time.Now()

	resp, err = handler(ctx, req)

	fmt.Printf("Metrics: [%s] took [%s]\n", info.FullMethod, time.Since(now))
	return resp, err
}

//UnaryUniversalDeadline for add a deadline to unary endpoints
func UnaryUniversalDeadline(d time.Duration) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		//Also can get a deadline from the metadata
		ctx, _ = context.WithDeadline(ctx, time.Now().Add(d))

		done := make(chan bool)

		go func() {
			resp, err = handler(ctx, req)
			done <- true
		}()

		//Will return an error if the deadline passes before the handler is finished... should also figure out how to stop the handler from continuing if possible?
		select {
		case <-ctx.Done():
			return nil, errors.New("Unable to complete request due to deadline")
		case <-done:
			return resp, err
		}
	}
}

//UnaryAuth for handling logging for unary gRPC endpoints. Gets credentials from ctx and adds user info to ctx or rejects call
func UnaryAuth() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		//Get auth data that's passed from client in context.
		md, ok := metadata.FromContext(ctx)
		if !ok {
			fmt.Println("Could not get metadata", err)
		}

		//Check it to make sure it's good
		cred := md[authKey]
		fmt.Println("cred", cred)
		if cred == nil {
			//Reject call if not
			return nil, errors.New("Not authorized to make this call!")
		}

		//Add user data to ctx.
		ctx = context.WithValue(ctx, "user", cred)

		//Pass to next handler
		return handler(ctx, req)
	}
}

const authKey = "1"

//StreamLogging for handling logging for unary gRPC endpoints
func StreamLogging(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	//Do logging before streaming starts
	fmt.Println("Log: Calling", info.FullMethod)

	newStream := wrapServerStream(ss)

	newStream.recvMsgMiddleware = func(m interface{}, recv func(interface{}) error) error {
		//Log when messages are received
		fmt.Println("Log: Receiving msg from stream...")
		return recv(m)
	}

	newStream.sendMsgMiddleware = func(m interface{}, send func(interface{}) error) error {
		//Log when messages are sent
		fmt.Println("Log: Sending msg to stream...")
		return send(m)
	}

	err := handler(srv, newStream)

	//Do logging after streaming finishes
	fmt.Println("Log:", info.FullMethod, "completed.")

	return err
}

//StreamAuth for handling auth on streaming endpoints
func StreamAuth() grpc.StreamServerInterceptor {

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		cred := ss.Context().Value(authKey)
		if cred == nil {
			return errors.New("Not authorized to make this call!")
		}

		return handler(srv, ss)
	}
}

type streamMiddleware func(m interface{}, recv func(interface{}) error) error

//Wraps the server stream that is used for sending/receiving messages
type wrappedServerStream struct {
	grpc.ServerStream
	WrappedContext    context.Context
	recvMsgMiddleware streamMiddleware
	sendMsgMiddleware streamMiddleware
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

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *wrappedServerStream) Context() context.Context {
	return w.WrappedContext
}

func (w *wrappedServerStream) SendMsg(m interface{}) error {
	//TODO make it so that more than one "sendMsgMiddleware" can be added to the wrapped stream
	if w.sendMsgMiddleware != nil {
		return w.sendMsgMiddleware(m, w.ServerStream.SendMsg)
	}

	return w.ServerStream.SendMsg(m)
}
func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	//TODO make it so that more than one "recvMsgMiddleware" can be added to the wrapped stream
	if w.recvMsgMiddleware != nil {
		return w.recvMsgMiddleware(m, w.ServerStream.RecvMsg)
	}

	return w.ServerStream.RecvMsg(m)
}
