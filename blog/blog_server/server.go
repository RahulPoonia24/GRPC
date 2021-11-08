package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"example.com/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct {
}

type blogItem struct {
	Id       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorId string             `bson:"author_id,omitempty"`
	Content  string             `bson:"content,omitempty"`
	Title    string             `bson:"title,omitempty"`
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Inside CreateBlog method")
	blog := req.GetBlog()

	data := blogItem{
		AuthorId: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}
	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		fmt.Println("Failed to insert blog")
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to insert blog due to error:%v", err))
	}
	Id := res.InsertedID.(primitive.ObjectID)

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       Id.Hex(),
			AuthorId: blog.GetAuthorId(),
			Content:  blog.GetContent(),
			Title:    blog.GetTitle(),
		},
	}, nil
}

func dataToBlog(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.Id.Hex(),
		AuthorId: data.AuthorId,
		Title:    data.Title,
		Content:  data.Content,
	}
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Inside ReadBlog method")
	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"),
		)
	}
	data := &blogItem{}
	err = collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(data)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Cannot find the blog with given id: %v\n", err))
	}
	return &blogpb.ReadBlogResponse{
		Blog: dataToBlog(data),
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	log.Println("Inside UpdateBlog method")
	data := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(data.Id)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, "Cannot parse ID",
		)
	}
	blog := &blogItem{}
	blog.AuthorId = data.GetAuthorId()
	blog.Content = data.GetContent()
	blog.Title = data.GetTitle()
	updateErr := collection.FindOneAndUpdate(context.Background(), bson.M{"_id": oid}, bson.M{"$set": blog}).Decode(blog)
	if updateErr != nil {
		if updateErr == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("Cannot find the blog with given id: %v\n", err))
		} else {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Cannot update object in mongodb: %v", updateErr))
		}
	}
	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlog(blog),
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	log.Println("Inside DeleteBlog method")
	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, "Cannot parse ID",
		)
	}
	fmt.Println(oid)
	res, delErr := collection.DeleteOne(context.Background(), bson.M{"_id": oid})
	if delErr != nil {
		return nil, status.Error(
			codes.Internal, fmt.Sprintf("Cannnot Delete blog due to Error :%v", delErr),
		)
	}
	if res.DeletedCount == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Cannot find the blog with given id: %v\n", err))
	}
	return &blogpb.DeleteBlogResponse{
		BlogId: oid.Hex(),
	}, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("Inside ReadBlog method")
	cur, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return status.Error(
			codes.Internal, fmt.Sprintf("Cannnot Read blog due to Error :%v", err),
		)
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Error(
				codes.Internal, fmt.Sprintf("Cannnot Read blog due to Error :%v", err),
			)
		}
		stream.Send(&blogpb.ListBlogResponse{
			Blog: dataToBlog(data),
		})
	}
	return nil
}
func main() {
	// if we crash code, we will get exact file and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Blog Service started ")

	//connect to database
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Erorr creating a mongo client: %v", err)
	}

	mongoErr := client.Connect(context.TODO())
	if mongoErr != nil {
		log.Fatalf("Error while connecting to mongo client: %v", mongoErr)
	}
	//close the mongo connection at end
	defer client.Disconnect(context.TODO())

	//create or connect a collection
	collection = client.Database("BLOG").Collection("blog")
	fmt.Println(collection)

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}
	s := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(s, &server{})
	reflection.Register(s)
	go func() {
		fmt.Println("Starting server")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve : %v", err)
		}
	}()
	//wait for control c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// block until a signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("End of blog service")
}
