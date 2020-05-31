package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	pb "github.com/DaviSilvaP/calculator/calculator"
	"google.golang.org/grpc"
)

const (
	address          = "localhost:50051"
	defaultA float64 = 3.0
	defaultB float64 = 4.0
)

func calc(operation string, a, b float64, c pb.CalculatorClient) (string, float64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	switch operation {
	case "Sum":
		r, err := c.SumC(ctx, &pb.NumRequest{A: a, B: b})
		if err != nil {
			log.Fatalf("Could not %s: %v", operation, err)
		}
		return operation, r.GetResult()
	case "Subtraction":
		r, err := c.SubC(ctx, &pb.NumRequest{A: a, B: b})
		if err != nil {
			log.Fatalf("Could not %s: %v", operation, err)
		}
		return operation, r.GetResult()
	case "Multiplication":
		r, err := c.MultC(ctx, &pb.NumRequest{A: a, B: b})
		if err != nil {
			log.Fatalf("Could not %s: %v", operation, err)
		}
		return operation, r.GetResult()
	case "Division":
		r, err := c.DivC(ctx, &pb.NumRequest{A: a, B: b})
		if err != nil {
			log.Fatalf("Could not %s: %v", operation, err)
		}
		return operation, r.GetResult()
	default:
		return operation, 0.0
	}
}

func calcClientStreams(operation string, nums []*pb.NumSRequest, c pb.CalculatorClient) (string, float64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	switch operation {
	case "Sum":
		stream, err := c.SumEC(ctx)
		if err != nil {
			log.Fatalf("Could not %s: %v", operation, err)
		}
		for _, num := range nums {
			if err := stream.Send(num); err != nil {
				log.Fatalf("%v.Send(%v) = %v", stream, num, err)
			}
		}
		reply, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
		}
		return operation, reply.GetResult()
	case "Subtraction":
		stream, err := c.SubEC(ctx)
		if err != nil {
			log.Fatalf("Could not %s: %v", operation, err)
		}
		for _, num := range nums {
			if err := stream.Send(num); err != nil {
				log.Fatalf("%v.Send(%v) = %v", stream, num, err)
			}
		}
		reply, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
		}
		return operation, reply.GetResult()
	case "Multiplication":
		stream, err := c.MultEC(ctx)
		if err != nil {
			log.Fatalf("Could not %s: %v", operation, err)
		}
		for _, num := range nums {
			if err := stream.Send(num); err != nil {
				log.Fatalf("%v.Send(%v) = %v", stream, num, err)
			}
		}
		reply, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
		}
		return operation, reply.GetResult()
	case "Division":
		stream, err := c.DivEC(ctx)
		if err != nil {
			log.Fatalf("Could not %s: %v", operation, err)
		}
		for _, num := range nums {
			if err := stream.Send(num); err != nil {
				log.Fatalf("%v.Send(%v) = %v", stream, num, err)
			}
		}
		reply, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
		}
		return operation, reply.GetResult()
	default:
		return operation, 0.0
	}
}

func calcStreams(a, b float64, c pb.CalculatorClient) []float64 {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stream, err := c.AllCalcs(ctx, &pb.NumRequest{A: a, B: b})
	if err != nil {
		log.Fatalf("Could not use calcStreams: %v.calcStreams(_) = _, %v", c, err)
	}

	calcs := make([]float64, 0, 10)

	for {
		result, err := stream.Recv()
		if err == io.EOF {
			return calcs
		}
		if err != nil {
			log.Fatalf("%v.calcStreams(_) = _, %v", c, err)
		}
		calcs = append(calcs, result.GetResult())
	}
}

func calcEachStreams(nums []*pb.NumSRequest, c pb.CalculatorClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stream, err := c.AllCalcsE(ctx)
	if err != nil {
		log.Fatalf("Could not use calcEachStreams: %v.calcEachStreams(_) = _, %v", c, err)
	}

	channel := make(chan float64)

	go func() {
		calcs := make([]float64, 0, 10)
		for i := 1; ; i++ {
			result, err := stream.Recv()
			if err == io.EOF {
				close(channel)
				return
			}
			if err != nil {
				log.Fatalf("Failed in receive values in %v because: %v", c, err)
			}
			calcs = append(calcs, result.GetResult())
			fmt.Printf("%dª result: %v\n", i, calcs)
		}
	}()

	for _, num := range nums {
		if err := stream.Send(num); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream, num, err)
		}
	}

	stream.CloseSend()
	// Mantém o processo pai vivo para continuar recebendo os resultados
	<-channel
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCalculatorClient(conn)

	// Contact the server and print out its response.
	a := defaultA
	b := defaultB
	if len(os.Args) > 2 {
		a, _ = strconv.ParseFloat(os.Args[1], 64)
		b, _ = strconv.ParseFloat(os.Args[2], 64)
	}

	operation, result := calc("Sum", a, b, c)
	log.Printf("%s = %.2f", operation, result)

	operation, result = calc("Subtraction", a, b, c)
	log.Printf("%s = %.2f", operation, result)

	operation, result = calc("Multiplication", a, b, c)
	log.Printf("%s = %.2f", operation, result)

	operation, result = calc("Division", a, b, c)
	log.Printf("%s = %.2f", operation, result)

	nums := []*pb.NumSRequest{
		&pb.NumSRequest{Num: a},
		&pb.NumSRequest{Num: b},
	}

	if len(os.Args) > 1 {
		nums = make([]*pb.NumSRequest, 0, 10)
		for _, value := range os.Args[1:] {
			num, _ := strconv.ParseFloat(value, 64)
			nums = append(nums, &pb.NumSRequest{Num: num})
		}
	}

	operation, result = calcClientStreams("Sum", nums, c)
	log.Printf("%s = %v", operation, result)

	operation, result = calcClientStreams("Subtraction", nums, c)
	log.Printf("%s = %v", operation, result)

	operation, result = calcClientStreams("Multiplication", nums, c)
	log.Printf("%s = %v", operation, result)

	operation, result = calcClientStreams("Division", nums, c)
	log.Printf("%s = %v", operation, result)

	//

	results := calcStreams(a, b, c)
	log.Printf("Result of calcStreams: %v", results)

	calcEachStreams(nums, c)

}
