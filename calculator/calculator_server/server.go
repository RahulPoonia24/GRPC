package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"example.com/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Println("Sum function is invoked: \n", req)
	firstName := req.GetFirstNumber()
	lastNumber := req.GetLastNumber()
	result := firstName + lastNumber
	res := &calculatorpb.SumResponse{
		Result: result,
	}
	return res, nil
}

func (*server) PrimDecom(req *calculatorpb.PrimeDecompositionRequest, stream calculatorpb.CalculatorService_PrimDecomServer) error {

	number := req.GetNumber()
	var k int32 = 2
	for number > 1 {
		if number%k == 0 {
			res := &calculatorpb.PrimeDecompositionResponse{
				Result: k,
			}
			stream.Send(res)
			number = number / k
		} else {
			k = k + 1
		}
	}
	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Println("ComputeAverage function is invoked with a streaming request")
	var computeAverage []float32
	for {
		req, err := stream.Recv()
		number := float32(0)
		if err == io.EOF {
			//porcessing finish
			for _, num := range computeAverage {
				number += num
			}
			result := float32(number / float32(len(computeAverage)))
			return stream.SendAndClose(
				&calculatorpb.ComputeAverageResponse{
					Result: result,
				},
			)
		}
		if err != nil {
			log.Fatalf("Error while reading client stream %v", err)
		}
		reqnumber := req.GetNumber()
		computeAverage = append(computeAverage, reqnumber)
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Println("greet function is invoked with a streaming request")
	var numberArray []int32
	maximum := int32(0)
	currentMaxi := int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream %v", err)
			return err
		}
		reqnumber := req.GetNumber()
		numberArray = append(numberArray, reqnumber)
		for index, number := range numberArray {
			if index == 0 || number > maximum {
				maximum = number
			}
		}
		if currentMaxi != maximum {
			fmt.Println("new maximun found :", currentMaxi)
			currentMaxi = maximum
			err = stream.Send(&calculatorpb.FindMaximumResponse{
				Result: maximum,
			})
			if err != nil {
				log.Fatalf("Error while sending data to  client: %v", err)
				return err
			}
		} else {
			fmt.Println("No new maximun found :", currentMaxi)
		}

	}

}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Println("Received SquareRoot rpc")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Error(
			codes.InvalidArgument, fmt.Sprintf("received a negative number %v", number),
		)
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Println("welcome to greet server")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}
	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve : %v", err)
	}
}
