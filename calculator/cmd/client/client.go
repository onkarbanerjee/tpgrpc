package main

import (
	"context"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	if err = doSum(client); err != nil {
		if _, ok := status.FromError(err); ok {
			log.Println("Got a grpc error", err)
		} else {
			log.Println("Got a big error", err)
		}
		return
	}

	if err = doDecompose(client); err != nil {
		if _, ok := status.FromError(err); ok {
			log.Println("Got a grpc error", err)
		} else {
			log.Println("Got a big error", err)
		}
		return
	}

	if err = doAverage(client); err != nil {
		if _, ok := status.FromError(err); ok {
			log.Println("Got a grpc error", err)
		} else {
			log.Println("Got a big error", err)
		}
		return
	}

	if err = doMax(client); err != nil {
		if _, ok := status.FromError(err); ok {
			log.Println("Got a grpc error", err)
		} else {
			log.Println("Got a big error", err)
		}
		return
	}

	log.Println("Bye")
}

func doSum(client calculator.CalculatorServiceClient) error {
	req := &calculator.SumRequest{
		First: 5, Second: 10,
	}
	resp, err := client.Sum(context.Background(), req)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not get a response , got error %s", err)
	}
	log.Println("Sum response is", resp.Result)
	return status.Error(codes.OK, "Success!")
}

func doDecompose(client calculator.CalculatorServiceClient) error {
	decomposeReq := calculator.DecomposeRequest{
		Prime: 120,
	}
	respStream, err := client.Decompose(context.Background(), &decomposeReq)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not get a response , got error %s", err)
	}

	for {
		resp, err := respStream.Recv()
		if err == io.EOF {
			log.Println("Receiving no more")
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "Could not get a response , got error %s", err)
		}
		log.Println("Received ", resp.Result)
	}

	log.Println("Done with decompose")
	return status.Error(codes.OK, "SUccess!")
}

func doAverage(client calculator.CalculatorServiceClient) error {
	stream, err := client.Average(context.Background())
	if err != nil {
		return status.Errorf(codes.Internal, "Could not get a client stream to send requests to, got error %s", err)
	}
	avgRequests := []calculator.AvgRequest{
		calculator.AvgRequest{Number: 1},
		calculator.AvgRequest{Number: 2},
		calculator.AvgRequest{Number: 3},
		calculator.AvgRequest{Number: 4},
	}

	for _, req := range avgRequests {
		if err = stream.Send(&req); err != nil {
			return status.Errorf(codes.Internal, "Could not send to stream , got error %s", err)
		}
		<-time.After(time.Second)
	}
	avgResp, err := stream.CloseAndRecv()
	if err != nil {
		return status.Errorf(codes.Internal, "Could not get Avg response , got error %s", err)
	}
	log.Println("Got avg response from server", avgResp.Result)
	return status.Error(codes.OK, "Success!")
}

func doMax(client calculator.CalculatorServiceClient) error {
	maxStream, err := client.Maximum(context.Background())
	if err != nil {
		return status.Errorf(codes.Internal, "Could not get a client stream to send requests to, %s", err)
	}
	maxRequests := []calculator.MaxRequest{
		calculator.MaxRequest{Number: 1},
		calculator.MaxRequest{Number: 5},
		calculator.MaxRequest{Number: 3},
		calculator.MaxRequest{Number: 6},
		calculator.MaxRequest{Number: 2},
		calculator.MaxRequest{Number: 20},
	}

	done := make(chan struct{})
	go func() {
		for _, req := range maxRequests {
			if err = maxStream.Send(&req); err != nil {
				log.Println("Could not send request to max stream", err)
				return
			}
			<-time.After(time.Second)
		}
		maxStream.CloseSend()
		log.Println("Finished sending all max requests")
	}()

	go func() {
		defer func() {
			done <- struct{}{}
		}()

		for {
			resp, err := maxStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println("Could not get response from max stream", err)
				return
			}
			log.Println("Got response ", resp.GetResult())
		}
	}()

	<-done
	return status.Error(codes.OK, "Success!")
}
