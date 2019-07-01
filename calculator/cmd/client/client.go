package main

import (
	"context"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"

	calculator "github.com/onkarbanerjee/tpgrpc/calculator/pkg"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("Could not get a client connection", err)
	}
	defer conn.Close()

	client := calculator.NewCalculatorServiceClient(conn)

	req := &calculator.SumRequest{
		First: "5", Second: "10",
	}
	resp, err := client.Sum(context.Background(), req)
	if err != nil {
		log.Println("Could not get a response", err)
		return
	}
	log.Println("Sum response is", resp.Result)

	decomposeReq := calculator.DecomposeRequest{
		Prime: "120",
	}
	respStream, err := client.Decompose(context.Background(), &decomposeReq)
	if err != nil {
		log.Println("Could not get a response stream", err)
	}

	for {
		resp, err := respStream.Recv()
		if err == io.EOF {
			log.Println("Receing no more")
			break
		}
		if err != nil {
			log.Println("COuld not receive a response from server stream", err)
		}
		log.Println("Received ", resp.Result)
	}

	log.Println("Done with decompose")

	stream, err := client.Average(context.Background())
	if err != nil {
		log.Println("Could not get a client stream to send requests to", err)
		return
	}
	avgRequests := []calculator.AvgRequest{
		calculator.AvgRequest{Number: "1"},
		calculator.AvgRequest{Number: "2"},
		calculator.AvgRequest{Number: "3"},
		calculator.AvgRequest{Number: "4"},
	}

	for _, req := range avgRequests {
		if err = stream.Send(&req); err != nil {
			log.Println("COuld not send to stream", err)
			return
		}
		time.Sleep(time.Second)
	}
	avgResp, err := stream.CloseAndRecv()
	if err != nil {
		log.Println("COuld not get avg response from server", err)
		return
	}
	log.Println("Got avg response from server", avgResp.Result)
}
