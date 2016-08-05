package middleware

import (
	"fmt"
	"time"

	"github.com/weave-lab/wlib/wlog"
	"github.com/weave-lab/wlib/wlog/tag"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var Logger = wlog.NewWLogger(wlog.WlogdLogger)

//UnaryLogging for handling logging for unary gRPC endpoints
func UnaryLogging(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	//TODO get request id

	//What info can we and should we log here
	Logger.InfoC(
		ctx,
		"",
		tag.String("FullMethod", info.FullMethod),
		tag.String("t", time.Now().String()))

	resp, err = handler(ctx, req)

	Logger.InfoC(
		ctx,
		"",
		tag.String("t", time.Now().String()),
		tag.String("duration", time.Since(start).String()))

	return resp, err
}

//StreamLogging for handling logging for unary gRPC endpoints
func StreamLogging(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	//Do logging before streaming starts
	fmt.Println("Log: Calling", info.FullMethod)

	newStream := wrapServerStream(ss)

	newStream.RegisterRecvMiddleware(func(inner StreamHandler) StreamHandler {

		//Log when messages are received
		mw := func(m interface{}) error {
			fmt.Printf("Log: Receiving msg from stream: %v\n", m)

			err := inner.Stream(m)

			fmt.Printf("Log: Received msg from stream: %v\n", m)
			return err
		}

		return StreamFunc(mw)
	})

	newStream.RegisterSendMiddleware(func(inner StreamHandler) StreamHandler {
		//Log when messages are sent
		mw := func(m interface{}) error {
			fmt.Printf("Log: Sending msg to stream: %v\n", m)

			err := inner.Stream(m)

			fmt.Printf("Log: Sent msg to stream: %v\n", m)
			return err
		}

		return StreamFunc(mw)
	})

	err := handler(srv, newStream)

	//Do logging after streaming finishes
	fmt.Println("Log:", info.FullMethod, "completed.")

	if err != nil {
		return err
	}

	return nil
}
