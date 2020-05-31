package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	pb "github.com/DaviSilvaP/calculator/calculator"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// Usado para implementar o calculator.CalculatorServer
type server struct {
	pb.UnimplementedCalculatorServer
	replies []*pb.NumReply
	mu      sync.Mutex // Mutex para projeter os calculos de diferentes clientes
}

func (s *server) SumC(ctx context.Context, in *pb.NumRequest) (*pb.NumReply, error) {
	a := in.GetA()
	b := in.GetB()
	log.Printf("SumC received: (%.2f, %.2f)", a, b)
	return &pb.NumReply{Result: a + b}, nil
}

func (s *server) SubC(ctx context.Context, in *pb.NumRequest) (*pb.NumReply, error) {
	a := in.GetA()
	b := in.GetB()
	log.Printf("SubC received: (%.2f, %.2f)", a, b)
	return &pb.NumReply{Result: a - b}, nil
}

func (s *server) MultC(ctx context.Context, in *pb.NumRequest) (*pb.NumReply, error) {
	a := in.GetA()
	b := in.GetB()
	log.Printf("MultC received: (%.2f, %.2f)", a, b)
	return &pb.NumReply{Result: a * b}, nil
}

func (s *server) DivC(ctx context.Context, in *pb.NumRequest) (*pb.NumReply, error) {
	a := in.GetA()
	b := in.GetB()
	log.Printf("DivC received: (%.2f, %.2f)", a, b)
	return &pb.NumReply{Result: a / b}, nil
}

func (s *server) SumEC(stream pb.Calculator_SumECServer) error {
	log.Printf("SumEC received ...")
	var result float64
	num, err := stream.Recv()
	if err == io.EOF {
		return stream.SendAndClose(&pb.NumReply{Result: result})
	} else if err != nil {
		return err
	}
	result = num.GetNum()

	for {
		num, err = stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.NumReply{Result: result})
		} else if err != nil {
			return err
		}
		result += num.GetNum()
	}
}

func (s *server) SubEC(stream pb.Calculator_SubECServer) error {
	log.Printf("SubEC received ...")
	var result float64
	num, err := stream.Recv()
	if err == io.EOF {
		return stream.SendAndClose(&pb.NumReply{Result: result})
	} else if err != nil {
		return err
	}
	result = num.GetNum()

	for {
		num, err = stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.NumReply{Result: result})
		} else if err != nil {
			return err
		}
		result -= num.GetNum()
	}
}

func (s *server) MultEC(stream pb.Calculator_MultECServer) error {
	log.Printf("MultEC received ...")
	var result float64
	num, err := stream.Recv()
	if err == io.EOF {
		return stream.SendAndClose(&pb.NumReply{Result: result})
	} else if err != nil {
		return err
	}
	result = num.GetNum()

	for {
		num, err = stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.NumReply{Result: result})
		} else if err != nil {
			return err
		}
		result *= num.GetNum()
	}
}

func (s *server) DivEC(stream pb.Calculator_DivECServer) error {
	log.Printf("DivEC received ...")
	var result float64
	num, err := stream.Recv()
	if err == io.EOF {
		return stream.SendAndClose(&pb.NumReply{Result: result})
	} else if err != nil {
		return err
	}
	result = num.GetNum()

	for {
		num, err = stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.NumReply{Result: result})
		} else if err != nil {
			return err
		}
		result /= num.GetNum()
	}
}

func (s *server) AllCalcs(in *pb.NumRequest, stream pb.Calculator_AllCalcsServer) error {
	a := in.GetA()
	b := in.GetB()
	log.Printf("AllCalcs received: (%.2f, %.2f)", a, b)

	s.mu.Lock()
	s.replies = []*pb.NumReply{
		&pb.NumReply{Result: a + b},
		&pb.NumReply{Result: a - b},
		&pb.NumReply{Result: a * b},
		&pb.NumReply{Result: a / b},
	}
	s.mu.Unlock()

	for _, reply := range s.replies {
		if err := stream.Send(reply); err != nil {
			return err
		}
	}
	return nil
}

func operations(operation string, nums []float64) float64 {
	result := nums[0]

	switch operation {
	case "+":
		for _, num := range nums[1:] {
			result += num
		}
	case "-":
		for _, num := range nums[1:] {
			result -= num
		}
	case "*":
		for _, num := range nums[1:] {
			result *= num
		}
	case "/":
		for _, num := range nums[1:] {
			result /= num
		}
	}
	return result
}

func (s *server) AllCalcsE(stream pb.Calculator_AllCalcsEServer) error {
	log.Println("AllCalcsE received ...")
	nums := make([]float64, 0, 10)
	for {
		num, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		nums = append(nums, num.GetNum())

		s.mu.Lock()
		s.replies = []*pb.NumReply{
			&pb.NumReply{Result: operations("+", nums)},
			&pb.NumReply{Result: operations("-", nums)},
			&pb.NumReply{Result: operations("*", nums)},
			&pb.NumReply{Result: operations("/", nums)},
		}
		s.mu.Unlock()

		for _, reply := range s.replies {
			if err := stream.Send(reply); err != nil {
				return err
			}
		}

	}
}

func init() {
	fmt.Println("Server started ...")
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCalculatorServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
