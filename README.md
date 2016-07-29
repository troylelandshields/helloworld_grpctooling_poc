# GRPC middleware POC

Just want to see if we can add things like logging, metrics, etc. to gRPC endpoints.

The gRPC server has allows setting a single "interceptor" for unary endpoints and a single "interceptor" for streaming endpoints. 
Although only 1 is allowed, the `grpc_middleware` package allows chaining multiple interceptors that get wrapped into one.

The bottom-line is that it looks like middleware and tooling is totally possible and quite flexible. The downside is that each middleware 
has to be written to support both unary and streaming endpoints since the interceptors are different interfaces.

_Sidenote: there are quite a few gRPC options that can be specified when starting a gRPC server or when executing gRPC requests from the client that also
give a lot of flexibility to each request. I did not look into these that much, so it's something to explore further._

To see this in magic:

* Step 1, start the server: `go run greeter_server/main.go`

* Step 2, in a separate terminal run the client:
`go run greeter_client/main.go`

* Step 3, profit.