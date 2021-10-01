package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrBlocked reports if service is blocked.
var ErrBlocked = errors.New("blocked")

// Service defines external service that can process batches of items.
type Service interface {
	GetLimits() (n uint64, p time.Duration)
	Process(ctx context.Context, batch Batch) error
}

type Server struct {
	elemLimit uint64
	timeLimit time.Duration
}

func CreateServer(n uint64, p time.Duration) Server {
	s := Server{
		elemLimit: n,
		timeLimit: p,
	}
	return s
}

func (serv Server) GetLimits() (n uint64, p time.Duration) {
	return serv.elemLimit, serv.timeLimit
}

func (serv Server) Process(ctx context.Context, batch Batch) error {
	fmt.Printf("Items sent: %d\n", len(batch))
	if uint64(len(batch)) <= serv.elemLimit {
		return nil
	} else {
		return ErrBlocked
	}
}

// Batch is a batch of items.
type Batch []Item

// Item is some abstract item.
type Item struct{
	num int
}
