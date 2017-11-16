package main

import (
	"crypto/rand"
	"flag"
	"log"
	"math"
	"net"
	"time"

	pingpb "github.com/petermattis/pinger/pingpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var streaming = flag.Bool("g", false, "use a streaming grpc RPC")

type pinger struct {
	payload []byte
}

func newPinger() *pinger {
	payload := make([]byte, *serverPayload)
	_, _ = rand.Read(payload)
	return &pinger{payload: payload}
}

func (p *pinger) Ping(_ context.Context, req *pingpb.PingRequest) (*pingpb.PingResponse, error) {
	return &pingpb.PingResponse{Payload: p.payload}, nil
}

func (p *pinger) PingStream(s pingpb.Pinger_PingStreamServer) error {
	for {
		if _, err := s.Recv(); err != nil {
			return err
		}
		if err := s.Send(&pingpb.PingResponse{Payload: p.payload}); err != nil {
			return err
		}
	}
}

func doGrpcServer(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.MaxMsgSize(math.MaxInt32),
		grpc.MaxConcurrentStreams(math.MaxInt32),
		grpc.InitialWindowSize(65535),
		grpc.InitialConnWindowSize(65535),
	)
	pingpb.RegisterPingerServer(s, newPinger())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func grpcWorker(c pingpb.PingerClient) {
	payload := make([]byte, *clientPayload)
	_, _ = rand.Read(payload)

	var s pingpb.Pinger_PingStreamClient
	if *streaming {
		var err error
		s, err = c.PingStream(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}

	for {
		start := time.Now()
		var respLen int
		if s != nil {
			if err := s.Send(&pingpb.PingRequest{Payload: payload}); err != nil {
				log.Fatal(err)
			}
			resp, err := s.Recv()
			if err != nil {
				log.Fatal(err)
			}
			respLen = len(resp.Payload)
		} else {
			if c != nil {
				resp, err := c.Ping(context.TODO(), &pingpb.PingRequest{Payload: payload})
				if err != nil {
					log.Fatal(err)
				}
				respLen = len(resp.Payload)
			}
		}
		elapsed := clampLatency(time.Since(start), minLatency, maxLatency)
		stats.Lock()
		if err := stats.latency.Current.RecordValue(elapsed.Nanoseconds()); err != nil {
			log.Fatal(err)
		}
		stats.ops++
		stats.bytes += uint64(len(payload) + respLen)
		stats.Unlock()
	}
}

func doGrpcClient(addr string, cpp bool) {
	clients := make([]pingpb.PingerClient, *connections)
	for i := 0; i < len(clients); i++ {
		if !cpp {
			conn, err := grpc.Dial(addr,
				grpc.WithInsecure(),
				grpc.WithBlock(),
				grpc.WithInitialWindowSize(65535),
				grpc.WithInitialConnWindowSize(65535),
			)
			if err != nil {
				log.Fatal(err)
			}
			clients[i] = pingpb.NewPingerClient(conn)
		}
	}

	for i := 0; i < *concurrency; i++ {
		go grpcWorker(clients[i%len(clients)])
	}
}
