package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	calculator "github.com/onkarbanerjee/tpgrpc/calculator/pkg"
	"google.golang.org/grpc"
)

type server struct {
}

func (s *server) Sum(ctx context.Context, req *calculator.SumRequest) (*calculator.SumResponse, error) {
	log.Println("Received request", req)
	first, err := strconv.Atoi(req.First)
	if err != nil {
		return nil, err
	}
	second, err := strconv.Atoi(req.Second)
	if err != nil {
		return nil, err
	}

	return &calculator.SumResponse{
		Result: strconv.Itoa(first + second),
	}, nil
}

func (s *server) Decompose(req *calculator.DecomposeRequest, stream calculator.CalculatorService_DecomposeServer) error {
	log.Println("Received a request", req)
	n := req.GetPrime()

	number, err := strconv.Atoi(n)
	if err != nil {
		return fmt.Errorf("Not a valid number")
	}

	k := 2
	for number > 1 {
		if number%k == 0 {
			stream.Send(&calculator.DecomposeResponse{Result: strconv.Itoa(k)})
			number /= k
			time.Sleep(time.Second)
		} else {
			k++
		}
	}
	log.Println("Sent all prime components")
	return nil
}

func (s *server) Average(stream calculator.CalculatorService_AverageServer) error {
	log.Println("Got a streaming request")
	total, c := 0, 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&calculator.AvgResponse{Result: fmt.Sprintf("%f", float64(total)/float64(c))})
		}
		if err != nil {
			log.Println("COuld get a request from stream client request", err)
			return nil
		}
		number, err := strconv.Atoi(req.GetNumber())
		if err != nil {
			return fmt.Errorf("Invalid number passed *%s", err)
		}
		total += number
		c++
	}
}

func main() {

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Println("Could not get listener", err)
		return
	}

	s := grpc.NewServer()

	calculator.RegisterCalculatorServiceServer(s, &server{})

	log.Println("Staring Sum sever")
	if err = s.Serve(lis); err != nil {
		log.Println("Could not start server", err)
		return
	}
}
