package main

import (
	"context"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	greetpb "github.com/onkarbanerjee/tpgrpc/greet"

	"google.golang.org/grpc"
)

func main() {

	creds, err := credentials.NewClientTLSFromFile("certs/ca.crt", "")
	if err != nil {
		log.Println("Could not get credentials,got", err)
		return
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
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

	// perform bidirectional streaming
	doBidirectionalStreaming(client)

	// perform unary with deadline
	doGreetWithDeadline(client, 5*time.Second)
	doGreetWithDeadline(client, time.Second)

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

func doBidirectionalStreaming(client greetpb.GreetServiceClient) {

	stream, err := client.GreetEveryone(context.Background())
	if err != nil {
		log.Println("Could not get client stream, got error", err)
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
	done := make(chan struct{})
	go func() {
		for _, req := range requests {
			log.Println("Sending", req)
			if err = stream.Send(&req); err != nil {
				log.Println("Could not send in client stream, got error", err)
				return
			}
			<-time.After(time.Second)
		}
		stream.CloseSend()
	}()

	go func() {
		defer func() {
			done <- struct{}{}
		}()

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Println("Received all responses, now returning")
				return
			}
			if err != nil {
				log.Println("Could not receive in client stream, got", err)
				return
			}
			log.Println("got response ", resp.GetResult())
		}
	}()

	<-done
}

func doGreetWithDeadline(client greetpb.GreetServiceClient, duration time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	req := greetpb.GreetingRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Onkar",
			LastName:  "Banerjee",
		},
	}
	resp, err := client.GreetWithDeadline(ctx, &req)
	if err != nil {
		if statusErr, ok := status.FromError(err); ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				log.Println("Response was cancelled")
				return
			}
			log.Println("Response was cancelled for other some reason", statusErr)
		}
	}
	log.Println("Received response", resp.GetResult())
}
