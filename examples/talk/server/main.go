package main

import (
	"fmt"
	"github.com/MikkelHJuul/grpcConst"
	"github.com/MikkelHJuul/grpcConst/examples/talk/proto"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"time"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8888))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterTalkServiceServer(grpcServer, &talkServer{})
	grpcServer.Serve(lis)
}

type talkServer struct {
	proto.UnimplementedTalkServiceServer
}

func (t *talkServer) GetMany(_ *proto.EmptyRequest, stream proto.TalkService_GetManyServer) error {
	header, _ := grpcConst.HeaderSetConstant(&proto.SingleReply{Constant: "constant name"})
	stream.SendHeader(header)
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100; i++ {
		n := generator.Uint64()
		if err := stream.Send(&proto.SingleReply{Variable: n}); err != nil {
			return err
		}
	}
	return nil
}