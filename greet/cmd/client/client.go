package main

import (
	"context"
	"io"
	"log"
	"time"

	greetpb "github.com/onkarbanerjee/tpgrpc/greet"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("COuld not get a client connection", err)
	}
	defer conn.Close()

	client := greetpb.NewGreetServiceClient(conn)

	// perform unary request
	doUnary(client)

	// perform server side streaming request
	doServerSideStreaming(client)

	// perform client side streaming
	doClientSideStreaming(client)

}

func doUnary(client greetpb.GreetServiceClient) {

	req := &greetpb.GreetingRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Onkar",
			LastName:  "Banerjee",
		},
	}

	resp, err := client.Greet(context.Background(), req)
	if err != nil {
		log.Println("COuld not get a client response ", err)
	}

	log.Println("Got response", resp)
	return
}

func doServerSideStreaming(client greetpb.GreetServiceClient) {

	req := &greetpb.GreetingRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Onkar",
			LastName:  "Banerjee",
		},
	}
	respStream, err := client.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Println("COuld not get a respone stream from server", err)
	}

	for {
		resp, err := respStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Could not receive on respStream", err)
			return
		}
		log.Println("Received", resp.Result)
	}

}

func doClientSideStreaming(client greetpb.GreetServiceClient) {
	stream, err := client.GreetManyPeopleOnce(context.Background())

	if err != nil {
		log.Println("COuld not get a stream to send to client")
		return
	}

	requests := []greetpb.GreetingRequest{
		greetpb.GreetingRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Onkar",
				LastName:  "Banerjee",
			},
		},
		greetpb.GreetingRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Rohit",
				LastName:  "Sharma",
			},
		},
		greetpb.GreetingRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Pink",
				LastName:  "Floyd",
			},
		},
	}

	for _, req := range requests {
		if err = stream.Send(&req); err != nil {
			log.Println("COuld not send req", err)
			return
		}
		time.Sleep(time.Second)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Println("COuld not get the response from server", err)
		return
	}
	log.Println("Got response from server", resp.Result)
}
