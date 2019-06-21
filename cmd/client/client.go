package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/onkarbanerjee/tpgrpc/pkg/tpgrpc"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8088", grpc.WithInsecure())
	if err != nil {
		log.Println("Could not get a connection", err)
		return
	}
	defer conn.Close()

	client := tpgrpc.NewTpClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Hello(ctx, &tpgrpc.TPRequest{})
	if err != nil {
		log.Println("Could not get a response", err)
		return
	}
	fmt.Println("Server response is", resp.Text)

	resp, err = client.Time(ctx, &tpgrpc.TPRequest{})
	if err != nil {
		log.Println("Could not get a response", err)
		return
	}
	fmt.Println("Server response is", resp.Text)

	gc, err := client.Greetings(ctx, &tpgrpc.TPRequest{})
	if err != nil {
		log.Println("Could not get a greetings client", err)
		return
	}

	for i := 0; i < 2; i++ {
		resp, err = gc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Could not get a response", err)
			return
		}
		fmt.Println("Got", resp.Text)
	}
}
