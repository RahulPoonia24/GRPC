package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"example.com/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {

	fmt.Println("welcome to greet client")
	certFile := "ssl/ca.crt"
	creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
	if sslErr != nil {
		log.Fatalf("Failed to load ca trust certificate: %v", sslErr)
	}
	opts := grpc.WithTransportCredentials(creds)
	conn, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("could not connect : %v", err)
	}

	defer conn.Close()

	c := greetpb.NewGreetServiceClient(conn)

	doUnary(c)
	// doServerStream(c)
	// doClientStream(c)
	// doBiditStreaming(c)
	// doUnaryWithDeadling(c)

}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("inside do unary function ")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "rahul",
			LastName:  "poonia",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling greet RPC: %v", err)
	}
	log.Printf("Response form Greet: %v", res.Result)

}

func doServerStream(c greetpb.GreetServiceClient) {
	fmt.Println("inside do doServerStream function ")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "rahul",
			LastName:  "poonia",
		},
	}
	resStream, err := c.GreetMAnyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling greetManyTimes rpc:%v", err)
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
		log.Printf("Response from GreetMAnyTimes: %v", msg.Result)
	}

}

func doClientStream(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to  do a client streaming")

	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{FirstName: "rahul"},
		},
		{
			Greeting: &greetpb.Greeting{FirstName: "poonia"},
		},
		{
			Greeting: &greetpb.Greeting{FirstName: "test"},
		},
		{
			Greeting: &greetpb.Greeting{FirstName: "check"},
		},
		{
			Greeting: &greetpb.Greeting{FirstName: "mike"},
		},
	}

	stream, err := c.LongGreet(context.Background())
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

func doBiditStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to  do a bidi streaming")

	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{FirstName: "rahul"},
		},
		{
			Greeting: &greetpb.Greeting{FirstName: "poonia"},
		},
		{
			Greeting: &greetpb.Greeting{FirstName: "test"},
		},
		{
			Greeting: &greetpb.Greeting{FirstName: "check"},
		},
		{
			Greeting: &greetpb.Greeting{FirstName: "mike"},
		},
	}

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while calling longGreet %v", err)
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
			fmt.Println("GreetEveryone Response: \n", res)
		}
		close(waitc)

	}()

	//bloack until everything is done
	<-waitc
}

func doUnaryWithDeadling(c greetpb.GreetServiceClient) {
	fmt.Println("inside do unary with Deadline  function ")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "rahul",
			LastName:  "poonia",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {

		statusErr, ok := status.FromError(err)
		if ok {

			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("timeout was  hit! Deadline exceeded")
			} else {
				fmt.Println("unExpected Error occurred: ", err)
			}
		} else {
			log.Fatalf("error while calling greet RPC: %v", err)
		}
		return
	}
	log.Printf("Response form Greet: %v", res.Result)

}
