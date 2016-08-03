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
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	pb "github.com/troylelandshields/helloworld_grpctooling_poc/helloworld"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
	authKey     = "1"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx := metadata.NewContext(context.Background(), metadata.Pairs(authKey, "user123"))

	// Contact the server and print out its response.
	names := []string{defaultName}
	if len(os.Args) > 1 {
		names = append(names, os.Args[1:]...)
	}

	fmt.Println("Say hello to world:")
	sayHelloWorld(c, ctx)

	fmt.Println("Say hello to world without being authed")
	sayHelloWorld(c, context.Background())

	fmt.Println("Say hello to slow world:")
	sayHelloToSlowWorld(c, ctx)

	fmt.Println("Say hello to all my friends:")
	sayHelloToAllMyFriends(c, ctx)
}

func sayHelloWorld(c pb.GreeterClient, ctx context.Context) {
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "world"})
	if err != nil {
		log.Printf("error: could not greet world: %v", err)
		return
	}

	log.Printf("Greeting: %s\n\n", r.Message)
}

func sayHelloToSlowWorld(c pb.GreeterClient, ctx context.Context) {
	r, err := c.SayHelloSlow(ctx, &pb.HelloRequest{Name: "slow world"})
	if err != nil {
		log.Printf("error: could not greet slow world: %v", err)
		return
	}

	log.Printf("Greeting: %s\n\n", r.Message)
}

func sayHelloToAllMyFriends(c pb.GreeterClient, ctx context.Context) {
	names := []string{"Rob", "Sarah", "Michael", "Jon"}

	sm, err := c.SayHelloToMany(ctx)
	if err != nil {
		log.Fatal("Error sending requests for many:", err)
	}

	for _, n := range names {
		err := sm.Send(&pb.HelloRequest{Name: n})
		if err != nil {
			log.Fatal("Error sending requests for name:", n, err)
		}

		r, err := sm.Recv()
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}

		log.Printf("Greeting: %s\n\n", r.Message)
	}
	sm.CloseSend()
}
