package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"example.com/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {

	fmt.Println("welcome to greet client")

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("could not connect : %v", err)
	}

	defer conn.Close()

	c := calculatorpb.NewCalculatorServiceClient(conn)
	// sum(c)
	// PrimDecom(c)
	//ComputeAverage(c)
	// FindMaximum(c)
	SquareRoot(c, 10)
	SquareRoot(c, -10)

}

func SquareRoot(c calculatorpb.CalculatorServiceClient, number int32) {
	fmt.Println("inside squareroot unary RPC ...")
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: number})

	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			//actual error from gRPC(user error)
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("we probably sent a negative number ")
			}
		} else {
			log.Fatalf("Big Error calling SqaureRoot %v", err)
		}
	}

	fmt.Println("Result of square root of %v :%v/n", number, res.GetNumberRoot())
}

func sum(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.SumRequest{
		FirstNumber: 10,
		LastNumber:  3,
	}
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling sum RPC: %v", err)
	}
	log.Printf("Response form sum: %v", res.Result)
}

func PrimDecom(c calculatorpb.CalculatorServiceClient) {
	log.Printf("<--- inside primDecom method --->")
	req := &calculatorpb.PrimeDecompositionRequest{
		Number: 120,
	}
	resStream, err := c.PrimDecom(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling PrimDecom rpc:%v", err)
	}
	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// we've reached end of the stream
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Printf("Response from PrimDecom: %v", msg.Result)
	}

}

func ComputeAverage(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to  do a client streaming")

	requests := []*calculatorpb.ComputeAverageRequest{
		{
			Number: 1,
		},
		{
			Number: 2,
		},
		{
			Number: 3,
		},
		{
			Number: 4,
		},
	}

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling longGreet %v", err)
	}
	for _, req := range requests {
		fmt.Println("sending request: ", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiveing response: %v", err)
	}
	fmt.Println("LongGreet Response: ", res)
}

func FindMaximum(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to  do a bidi streaming")

	requests := []*calculatorpb.FindMaximumRequest{
		{
			Number: 1,
		},
		{
			Number: 5,
		},
		{
			Number: 3,
		},
		{
			Number: 6,
		},
		{
			Number: 2,
		},
		{
			Number: 20,
		},
	}

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while calling FindMaximum %v", err)
		return
	}

	waitc := make(chan struct{})
	go func() {
		//func to send a bunch of messages
		for _, req := range requests {
			fmt.Println("sending request: ", req)
			stream.Send(req)
			time.Sleep(100 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		//func to receve a bunch of messages
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiveing response: %v", err)
				break
			}
			fmt.Println("FindMaximum Response: \n", res)
		}
		close(waitc)
	}()

	//bloack until everything is done
	<-waitc
}
