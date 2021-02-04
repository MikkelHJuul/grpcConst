package main

import (
	"context"
	"github.com/MikkelHJuul/grpcConst"
	pb "github.com/MikkelHJuul/grpcConst/examples/ogc_ish/proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

func main() {
	conn, err := grpc.Dial("localhost:8181", grpc.WithInsecure(), grpc.WithStreamInterceptor(grpcConst.StreamClientInterceptor()))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewOGCishServiceClient(conn)
	ctx := context.Background()
	reqs := [...]*pb.FeatureCollectionRequest{
		bothRequest,
		stationRequest,
		measurementRequest,
		noneRequest,
	}
	allStart := time.Now()
	for _, req := range reqs {
		var start = time.Now()
		performRequest(client, ctx, req)
		log.Printf("%+v took %s", req, time.Since(start))
	}
	log.Printf("done in: %s", time.Since(allStart))
}

var bothRequest = &pb.FeatureCollectionRequest{
	StationName:     "06184",
	MeasurementName: "humidity",
}
var stationRequest = &pb.FeatureCollectionRequest{
	StationName: "06184",
}
var measurementRequest = &pb.FeatureCollectionRequest{
	MeasurementName: "humidity",
}
var noneRequest = &pb.FeatureCollectionRequest{
	StationName:     "",
	MeasurementName: "",
}

func performRequest(client pb.OGCishServiceClient, ctx context.Context, request *pb.FeatureCollectionRequest) {
	manyClient, err := client.Items(ctx, request)
	if err != nil {
		panic(err)
	}
	count := 0
	for {
		_, err := manyClient.Recv()
		if err == io.EOF {
			// read done.
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}
		//log.Printf("%v+", msg) uncomment to see the message (and save to msg)
		count++
	}
	log.Printf("Number returned: %d", count)
}
