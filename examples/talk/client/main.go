package main

import (
	"github.com/MikkelHJuul/grpcConst"
	"github.com/MikkelHJuul/grpcConst/examples/talk/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:8888", grpc.WithInsecure(), grpc.WithStreamInterceptor(grpcConst.StreamClientInterceptor()))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := proto.NewTalkServiceClient(conn)
	ctx := context.Background()
	manyClient, err := client.GetMany(ctx, &proto.EmptyRequest{})
	header, err := manyClient.Header()
	if err != nil {
		panic(err)
	}
	log.Print(header)
	for {
		in, err := manyClient.Recv()
		if err == io.EOF {
			// read done.
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}
		log.Printf("data: %s, %d", in.Constant, in.Variable)
	}
	log.Printf("done")
}
