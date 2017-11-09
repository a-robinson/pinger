package main

import (
	"crypto/rand"
	"flag"
	"log"
	"math"
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// #cgo CXXFLAGS: -std=c++11
// #cgo LDFLAGS: -L/usr/local/lib -lprotobuf -D_THREAD_SAFE -lgrpc++ -lgrpc -lgrpc++_reflection -ldl
// #include "grpc.h"
import "C"

var streaming = flag.Bool("g", false, "use a streaming grpc RPC")

type pinger struct {
	payload []byte
}

func newPinger() *pinger {
	payload := make([]byte, *serverPayload)
	_, _ = rand.Read(payload)
	return &pinger{payload: payload}
}

func (p *pinger) Ping(_ context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{Payload: p.payload}, nil
}

func (p *pinger) PingStream(s Pinger_PingStreamServer) error {
	for {
		if _, err := s.Recv(); err != nil {
			return err
		}
		if err := s.Send(&PingResponse{Payload: p.payload}); err != nil {
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
	RegisterPingerServer(s, newPinger())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func grpcWorker(c PingerClient, cppClient C.Client) {
	payload := make([]byte, *clientPayload)
	_, _ = rand.Read(payload)

	var s Pinger_PingStreamClient
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
			if err := s.Send(&PingRequest{Payload: payload}); err != nil {
				log.Fatal(err)
			}
			resp, err := s.Recv()
			if err != nil {
				log.Fatal(err)
			}
			respLen = len(resp.Payload)
		} else {
			if c != nil {
				resp, err := c.Ping(context.TODO(), &PingRequest{Payload: payload})
				if err != nil {
					log.Fatal(err)
				}
				respLen = len(resp.Payload)
			} else {
				respLen = int(C.ClientSend(cppClient))
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
	clients := make([]PingerClient, *connections)
	cppClients := make([]C.Client, *connections)
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
			clients[i] = NewPingerClient(conn)
		} else {
			cppClients[i] = C.NewClient()
		}
	}

	for i := 0; i < *concurrency; i++ {
		go grpcWorker(clients[i%len(clients)], cppClients[i%len(cppClients)])
	}
}
