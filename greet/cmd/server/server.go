package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	greetpb "github.com/onkarbanerjee/tpgrpc/greet"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
}

func (s *server) Greet(ctx context.Context, req *greetpb.GreetingRequest) (*greetpb.GreetingResponse, error) {
	log.Println("Received request", req)
	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()
	resp := &greetpb.GreetingResponse{
		Result: "Hello" + firstName + lastName,
	}
	return resp, nil
}

func (s *server) GreetManyTimes(req *greetpb.GreetingRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	log.Println("Received request", req)
	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		stream.Send(&greetpb.GreetingResponse{Result: firstName + strconv.Itoa(i)})
		time.Sleep(time.Second)
	}
	return nil
}

func (s *server) GreetManyPeopleOnce(stream greetpb.GreetService_GreetManyPeopleOnceServer) error {
	log.Println("Got a stream request")
	var result string
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.GreetingResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Println("Could not receive from client stream", err)
			return nil
		}
		firstName, lastName := req.GetGreeting().GetFirstName(), req.GetGreeting().GetLastName()
		result += "Hello " + firstName + " " + lastName
	}
}

func (s *server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	log.Println("Got a bidirectional streaming request")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "Could not receive from stream, got error")
		}
		firstName, lastName := req.GetGreeting().GetFirstName(), req.GetGreeting().GetLastName()
		stream.Send(&greetpb.GreetingResponse{
			Result: "Hello " + firstName + " " + lastName,
		})
	}
	return nil
}

func (s *server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetingRequest) (*greetpb.GreetingResponse, error) {
	log.Println("Recieved a request with Deadline")
	for i := 0; i < 3; i++ {
		<-time.After(time.Second)
		if err := ctx.Err(); err != nil {
			if err == context.DeadlineExceeded {
				log.Println("Received cancelled, hence returning")
				return nil, status.Error(codes.Canceled, "Cancelled")
			}
			log.Println("Received error, hence returning", err)
			return nil, status.Errorf(codes.Internal, "Internal error %s", err)
		}
	}
	return &greetpb.GreetingResponse{
		Result: "Hello " + req.GetGreeting().GetFirstName() + req.GetGreeting().GetLastName(),
	}, nil
}

func main() {

	log.Println("Starting....")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Println("Could not start listener", err)
		return
	}

	creds, err := credentials.NewServerTLSFromFile("certs/server.crt", "certs/server.pem")
	if err != nil {
		log.Println("COuld not get a credentials, got", err)
		return
	}

	s := grpc.NewServer(grpc.Creds(creds))
	greetpb.RegisterGreetServiceServer(s, &server{})

	// Register for reflection
	reflection.Register(s)

	stop, done := make(chan os.Signal), make(chan struct{})
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT)

	go func() {
		defer func() {
			done <- struct{}{}
		}()

		<-stop
		log.Println("Shutting down the server in 3 sec seconds")
		<-time.After(3 * time.Second)
		s.GracefulStop()
	}()

	if err := s.Serve(lis); err != nil {
		log.Println("Could not start server", err)
		return
	}

	<-done
	log.Println("Done... bye")

}
