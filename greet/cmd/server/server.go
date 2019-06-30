package main

import (
	"context"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	greetpb "github.com/onkarbanerjee/tpgrpc/greet"
	"google.golang.org/grpc"
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

func main() {

	log.Println("Starting....")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Println("Could not start listener", err)
		return
	}

	s := grpc.NewServer()

	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Println("Could not start server", err)
		return
	}

}
