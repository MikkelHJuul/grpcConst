package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/MikkelHJuul/grpcConst"
	pb "github.com/MikkelHJuul/grpcConst/examples/ogc_ish/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8181))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	db, err := initiateDatabase()
	if err != nil {
		panic(err)
	}
	pb.RegisterOGCishServiceServer(grpcServer, &ogcishServer{db: db})
	grpcServer.Serve(lis)
}

func initiateDatabase() ([]DBFeature, error) {
	jsonFile, err := os.Open("1995-09.txt")
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	features := make([]DBFeature, 0)
	s := bufio.NewScanner(jsonFile)
	for s.Scan() {
		var v DBFeature
		if err := json.Unmarshal(s.Bytes(), &v); err != nil {
			return nil, err
		}
		features = append(features, v)
	}
	if s.Err() != nil {
		return nil, err
	}
	return features, nil
}

type DBFeature struct {
	Id            string  `json:"_id"`
	ParameterName string  `json:"parameterId"`
	StationName   string  `json:"stationId"`
	TimeObserved  int64   `json:"timeObserved"`
	Value         float32 `json:"value"`
}

type ogcishServer struct {
	pb.UnimplementedOGCishServiceServer
	db []DBFeature
}

var geom = pb.Geometry{
	Type: "Point",
	Coordinates: &pb.Point{
		Latitude:  123,
		Longitude: 321,
	}}

func (o *ogcishServer) Items(fcReq *pb.FeatureCollectionRequest, stream pb.OGCishService_ItemsServer) error {
	if fcReq.MeasurementName != "" || fcReq.StationName != "" {
		constantFeature := &pb.Feature{
			Type: "Feature",
			Id:   "",
			Properties: &pb.Properties{
				Measurement: &pb.Measurement{
					Name:  fcReq.MeasurementName,
					Value: 0,
				},
			},
		}
		if fcReq.StationName != "" {
			time.Sleep(time.Duration(100)) // find the station data somehow...
			constantFeature.Properties.Station = &pb.Station{
				Name:     fcReq.StationName,
				Metadata: "Some Station Metadata",
			}
			constantFeature.Geometry = &geom
		}
		md, err := grpcConst.HeaderSetConstant(constantFeature)
		if err != nil {
			return err
		}
		stream.SetHeader(md)
	}
	for _, feature := range o.db {
		if fcReq.StationName != "" && fcReq.MeasurementName != "" {
			if feature.StationName == fcReq.StationName && feature.ParameterName == fcReq.MeasurementName {
				_ = stream.Send(&pb.Feature{
					Id: feature.Id,
					Properties: &pb.Properties{
						Measurement: &pb.Measurement{
							Value: feature.Value,
						},
					},
				})
			}
			continue
		}
		if fcReq.StationName != "" {
			if feature.StationName == fcReq.StationName {
				_ = stream.Send(&pb.Feature{
					Id: feature.Id,
					Properties: &pb.Properties{
						Measurement: &pb.Measurement{
							Name:  feature.ParameterName,
							Value: feature.Value,
						},
					},
				})
			}
			continue
		}
		if fcReq.MeasurementName != "" {
			if feature.ParameterName == fcReq.MeasurementName {
				_ = stream.Send(&pb.Feature{
					Id: feature.Id,
					Properties: &pb.Properties{
						Measurement: &pb.Measurement{
							Value: feature.Value,
						},
						Station: &pb.Station{
							Name:     feature.StationName,
							Metadata: "Some metadata",
						},
					},
					Geometry: &geom,
				})
			}
			continue
		}
		//both are empty string... stream everything
		_ = stream.Send(mapEverything(feature))
	}
	return nil
}

func (o *ogcishServer) ItemsNogRPCConst(fcReq *pb.FeatureCollectionRequest, stream pb.OGCishService_ItemsNogRPCConstServer) error {
	for _, feature := range o.db {
		if fcReq.StationName != "" && fcReq.MeasurementName != "" {
			if feature.StationName == fcReq.StationName && feature.ParameterName == fcReq.MeasurementName {
				_ = stream.Send(mapEverything(feature))
			}
			continue
		}
		if fcReq.StationName != "" {
			if feature.StationName == fcReq.StationName {
				_ = stream.Send(mapEverything(feature))
			}
			continue
		}
		if fcReq.MeasurementName != "" {
			if feature.ParameterName == fcReq.MeasurementName {
				_ = stream.Send(mapEverything(feature))
			}
			continue
		}
		//both are empty string... stream everything
		_ = stream.Send(mapEverything(feature))
	}
	return nil
}

func mapEverything(feature DBFeature) *pb.Feature {
	return &pb.Feature{
		Type: "Feature",
		Id:   feature.Id,
		Properties: &pb.Properties{
			Measurement: &pb.Measurement{
				Name:  feature.ParameterName,
				Value: feature.Value,
			},
			Station: &pb.Station{
				Name:     feature.StationName,
				Metadata: "Some metadata",
			},
		},
		Geometry: &geom,
	}
}
