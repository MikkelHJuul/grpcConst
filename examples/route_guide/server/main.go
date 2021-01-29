package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/MikkelHJuul/grpcConst"
	"google.golang.org/grpc"

	"github.com/MikkelHJuul/grpcConst/examples/route_guide/proto"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterRouteGuideServer(grpcServer, &routeServer{})
	grpcServer.Serve(lis)
}

type routeServer struct {
	proto.UnimplementedRouteGuideServer
}

func (t *routeServer) ListFeatures(_ *proto.Rectangle, stream proto.RouteGuide_ListFeaturesServer) error {
	header, _ := grpcConst.HeaderSetConstant(&proto.Feature{Name: "constant name", Location: &proto.Point{Latitude: 10}})
	stream.SendHeader(header)
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 10; i++ {
		n := generator.Int31()
		if err := stream.Send(&proto.Feature{Location: &proto.Point{Longitude: n}}); err != nil {
			return err
		}
	}
	for i := 0; i < 10; i++ {
		n := generator.Int31()
		if err := stream.Send(&proto.Feature{Name: "Other Name", Location: &proto.Point{Longitude: n, Latitude: 44}}); err != nil {
			return err
		}
	}
	for i := 0; i < 10; i++ {
		n := generator.Int31()
		if err := stream.Send(&proto.Feature{Name: "Third Name", Location: &proto.Point{Longitude: n}}); err != nil {
			return err
		}
	}
	return nil
}