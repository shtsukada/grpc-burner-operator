package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/shtsukada/grpc-burner-operator/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedBurnerServer
	messageSize int
}

func (s *server) UnaryBurn(ctx context.Context, req *pb.BurnRequest) (*pb.BurnResponse, error) {
	log.Printf("Received UnaryBurn request")
	payload := make([]byte, s.messageSize)
	return &pb.BurnResponse{Payload: payload}, nil
}

func main() {
	messageSize := 512

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterBurnerServer(s, &server{messageSize: messageSize})

	fmt.Println("gRPC Burner server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
