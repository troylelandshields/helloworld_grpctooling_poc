package middleware

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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
