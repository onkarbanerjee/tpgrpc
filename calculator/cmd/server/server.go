package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	calculator "github.com/onkarbanerjee/tpgrpc/calculator/pkg"
	"google.golang.org/grpc"
)

type server struct {
}

func (s *server) Sum(ctx context.Context, req *calculator.SumRequest) (*calculator.SumResponse, error) {
	log.Println("Received a Sum request", req)
	return &calculator.SumResponse{
		Result: req.GetFirst() + req.GetSecond(),
	}, status.Error(codes.OK, "Sucess")
}

func (s *server) Decompose(req *calculator.DecomposeRequest, stream calculator.CalculatorService_DecomposeServer) error {
	log.Println("Received a Decompose request", req)
	number := req.GetPrime()

	k := int32(2)
	for number > 1 {
		if number%k == 0 {
			stream.Send(&calculator.DecomposeResponse{Result: k})
			number /= k
			time.Sleep(time.Second)
		} else {
			k++
		}
	}
	log.Println("Sent all prime components")
	return status.Error(codes.OK, "Sucess")
}

func (s *server) Average(stream calculator.CalculatorService_AverageServer) error {
	log.Println("Received an average request")
	total, c := int32(0), 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&calculator.AvgResponse{Result: float64(total) / float64(c)})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "COuld get a request from stream client request, got error %s", err)
		}
		number := req.GetNumber()
		total += number
		c++
	}
}

func (s *server) Maximum(stream calculator.CalculatorService_MaximumServer) error {
	log.Println("Received a maximum request")
	max := int32(-1)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "COuld get a request from stream client request, got error %s", err)
		}
		number := req.GetNumber()

		if number > max {
			max = number
			stream.Send(&calculator.MaxResponse{
				Result: max,
			})
		}
	}
	return status.Error(codes.OK, "Sucess")
}

func main() {

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Println("Could not get listener", err)
		return
	}

	creds, err := credentials.NewServerTLSFromFile("certs/server.crt", "certs/server.pem")
	if err != nil {
		log.Println("COuld not get the credentials", err)
		return
	}
	s := grpc.NewServer(grpc.Creds(creds))

	calculator.RegisterCalculatorServiceServer(s, &server{})
	reflection.Register(s)

	stop, done := make(chan os.Signal), make(chan struct{})
	signal.Notify(stop, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		defer func() {
			done <- struct{}{}
		}()

		<-stop
		log.Println("Shutting down server in 3 seconds")
		<-time.After(3 * time.Second)
		s.GracefulStop()
	}()

	log.Println("Staring Calculator sever")
	if err = s.Serve(lis); err != nil {
		log.Println("Could not start server", err)
		return
	}

	<-done
	log.Println("Done... bye!!!")
}
