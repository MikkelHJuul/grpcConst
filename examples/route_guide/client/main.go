package main

import (
	"github.com/MikkelHJuul/grpcConst"
	"github.com/MikkelHJuul/grpcConst/examples/route_guide/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure(), grpc.WithStreamInterceptor(grpcConst.StreamClientInterceptor()))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := proto.NewRouteGuideClient(conn)
	ctx := context.Background()
	manyClient, err := client.ListFeatures(ctx, &proto.Rectangle{})
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
		log.Printf("data: %s, %+v", in.Name, in.Location)
	}
	log.Printf("done")
}
