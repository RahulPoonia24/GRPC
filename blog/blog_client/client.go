package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"example.com/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {

	fmt.Println("welcome to blog client")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("could not connect : %v", err)
	}

	defer conn.Close()

	c := blogpb.NewBlogServiceClient(conn)

	// createBlog(c)
	// readBlog(c)
	// UpdateBlog(c)
	// deleteBlog(c)
	listBlog(c)

}

func createBlog(c blogpb.BlogServiceClient) {
	fmt.Println("inside createBlogfunction ")
	req := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "Rahul",
			Title:    "My third Blog",
			Content:  "Content of my third Blog",
		},
	}
	res, err := c.CreateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling createBlog RPC: %v", err)
	}
	log.Printf("Response form CreateBlog: %v", res.Blog)

}

func readBlog(c blogpb.BlogServiceClient) {
	log.Println("inside readBlog Funciton ")
	req := &blogpb.ReadBlogRequest{
		BlogId: "618798d8d334a39f69d0ef24",
	}
	res, err := c.ReadBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling readBlog RPC: %v", err)
	}
	log.Printf("Response form readBlog: %v", res.Blog)
}

func UpdateBlog(c blogpb.BlogServiceClient) {
	log.Println("inside UpdateBlog Funciton ")
	req := &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:      "618798d8d334a39f69d0cf24",
			Title:   "My First Blogupdated",
			Content: "Content of my first Blog updated",
		},
	}
	res, err := c.UpdateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling updateBlog RPC: %v", err)
	}
	log.Printf("Response form updateBlog: %v", res.Blog)
}

func deleteBlog(c blogpb.BlogServiceClient) {
	log.Println("inside DeleteBlog Funciton ")
	req := &blogpb.DeleteBlogRequest{
		BlogId: "618798d8d334a39f69d0cf24",
	}
	res, err := c.DeleteBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling readBlog RPC: %v", err)
	}
	log.Printf("Response form readBlog: %v", res.BlogId)
}

func listBlog(c blogpb.BlogServiceClient) {
	log.Println("inside ListBlog Funciton ")
	req := &blogpb.ListBlogRequest{}
	stream, err := c.ListBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling ListBlog RPC: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happend: %v", err)
		}

		log.Println(res)

	}
}
