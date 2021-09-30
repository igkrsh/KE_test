package main

import (
	"context"
	"errors"
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
	processed uint64
}

func SetLimits(serv Server, n uint64, p time.Duration) {
	serv.elemLimit = n
	serv.timeLimit = p
	serv.processed = 0
}

func GetLimits(serv Server) (n uint64, p time.Duration) {
	return serv.elemLimit, serv.timeLimit
}

func Process(serv Server, ctx context.Context, batch Batch) error {
	if serv.processed < serv.elemLimit {
		serv.processed ++
	} else {
		return ErrBlocked
	}
	return nil
}

// Batch is a batch of items.
type Batch []Item

// Item is some abstract item.
type Item struct{}
