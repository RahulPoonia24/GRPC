#greet 
protoc greet/greetpb/greet.proto --go_out=plugins=grpc:.

#calculator
protoc calculator/calculatorpb/calculator.proto --go_out=plugins=grpc:.

#blog
protoc blog/blogpb/blog.proto --go_out=plugins=grpc:.
