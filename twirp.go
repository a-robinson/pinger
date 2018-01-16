package main

import (
	"crypto/rand"
	"log"
	http "net/http"
	"time"

	"golang.org/x/net/context"
)

type twirpPinger struct {
	payload []byte
}

func newTwirpPinger() *twirpPinger {
	payload := make([]byte, *serverPayload)
	_, _ = rand.Read(payload)
	return &twirpPinger{payload: payload}
}

func (p *twirpPinger) Ping(_ context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{Payload: p.payload}, nil
}

func (p *twirpPinger) PingStream(_ context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{Payload: p.payload}, nil
}

func doTwirpServer(port string) {
	s := NewPingerServer(newTwirpPinger(), nil)
	mux := http.NewServeMux()
	mux.Handle(PingerPathPrefix, s)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func twirpWorker(c Pinger) {
	payload := make([]byte, *clientPayload)
	_, _ = rand.Read(payload)

	for {
		start := time.Now()
		var resp *PingResponse
		var err error
		resp, err = c.Ping(context.TODO(), &PingRequest{Payload: payload})
		if err != nil {
			log.Fatal(err)
		}
		elapsed := clampLatency(time.Since(start), minLatency, maxLatency)
		stats.Lock()
		if err := stats.latency.Current.RecordValue(elapsed.Nanoseconds()); err != nil {
			log.Fatal(err)
		}
		stats.ops++
		stats.bytes += uint64(len(payload) + len(resp.Payload))
		stats.Unlock()
	}
}

func doTwirpClient(addr string) {
	clients := make([]Pinger, *connections)
	for i := 0; i < len(clients); i++ {
		clients[i] = NewPingerProtobufClient("http://"+addr, &http.Client{})
	}

	for i := 0; i < *concurrency; i++ {
		go twirpWorker(clients[i%len(clients)])
	}
}
