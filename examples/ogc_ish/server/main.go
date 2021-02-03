package main

import (
	"github.com/MikkelHJuul/grpcConst/examples/ogc_ish/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	initiateDatabase()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8181))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterOGCishServiceServer(grpcServer, &ogcishServer{})
	grpcServer.Serve(lis)
}

func initiateDatabase() {

}

type ogcishServer struct {
	proto.UnimplementedOGCishServiceServer
}

func (receiver *ogcishServer) Items(fcReq *FeatureCollectionRequest, stream OGCishService_ItemsServer) {

}

func (receiver *ogcishServer) ItemsNogRPCConst(fcReq *FeatureCollectionRequest, stream OGCishService_ItemsServer) {

}
