package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/onkarbanerjee/tpgrpc/pkg/tpgrpc"
)

type RealTpServer struct {
}

func (*RealTpServer) Hello(ctx context.Context, req *tpgrpc.TPRequest) (*tpgrpc.TPResponse, error) {
	return &tpgrpc.TPResponse{Text: "Hi"}, nil
}
func (*RealTpServer) Time(ctx context.Context, req *tpgrpc.TPRequest) (*tpgrpc.TPResponse, error) {
	t := time.Now()
	return &tpgrpc.TPResponse{Text: t.String()}, nil
}

func (*RealTpServer) Greetings(req *tpgrpc.TPRequest, srv tpgrpc.Tp_GreetingsServer) error {
	srv.Send(&tpgrpc.TPResponse{Text: "Good morning"})
	srv.Send(&tpgrpc.TPResponse{Text: "Good afternoon"})
	srv.Send(&tpgrpc.TPResponse{Text: "Good evening"})
	return nil
}

func main() {
	l, err := net.Listen("tcp", ":8088")
	if err != nil {
		log.Println("Could get listener", err)
		return
	}

	srv := grpc.NewServer()

	tpgrpc.RegisterTpServer(srv, &RealTpServer{})

	srv.Serve(l)
}
