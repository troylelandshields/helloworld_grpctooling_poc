/*
 *
 * Copyright 2015, Google Inc.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     * Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above
 * copyright notice, this list of conditions and the following disclaimer
 * in the documentation and/or other materials provided with the
 * distribution.
 *     * Neither the name of Google Inc. nor the names of its
 * contributors may be used to endorse or promote products derived from
 * this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/weave-lab/wlib/wlog"
	"github.com/weave-lab/wlib/wlog/tag"

	"github.com/weave-lab/helloworld_grpctooling_poc/greeter_server/middleware"
	"github.com/weave-lab/helloworld_grpctooling_poc/greeter_server/server"
	pb "github.com/weave-lab/helloworld_grpctooling_poc/helloworld"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type greeterserver struct{}

// SayHello implements helloworld.GreeterServer
func (s *greeterserver) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Println("Responding to", in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

//SayHelloSlow takes 5 seconds to say hello
func (s *greeterserver) SayHelloSlow(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	<-time.After(5 * time.Second)
	return &pb.HelloReply{Message: "Helllllllooooooo " + in.Name}, nil
}

//SayHelloToMany receives requests from a stream and sends responses to a stream
func (s *greeterserver) SayHelloToMany(ss pb.Greeter_SayHelloToManyServer) error {
	for {
		in, err := ss.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Err while receiving name from many:", err)
			return err
		}

		err = ss.Send(&pb.HelloReply{Message: "Hello to " + in.Name})
		if err != nil {
			fmt.Println("Err while saying hello to many:", err)
			return err
		}
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//Create a gRPC server with default middleware and add unary deadline and unary auth middleware too
	s := server.New([]grpc.UnaryServerInterceptor{
		middleware.UnaryUniversalDeadline(1 * time.Second),
		middleware.UnaryAuth(),
	}, nil)

	pb.RegisterGreeterServer(s, &greeterserver{})

	fmt.Println("Starting server...")
	go s.Serve(lis)

	//Listen for signal to execute graceful shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)
	for sig := range signalChannel {
		switch sig {
		case syscall.SIGTERM, syscall.SIGINT:
			fmt.Println("stopping server...")
			s.GracefulStop()
			fmt.Println("stopped server...")
			os.Exit(0)
		default:
			wlog.Error("Received unknown signal", tag.String("signal", sig.String()))
		}
	}
}
