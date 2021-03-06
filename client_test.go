package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

var ProcessMock func (ctx context.Context, batch Batch) error

type ServerMock struct {
	ElemLimit uint64
	TimeLimit time.Duration
}

func (s * ServerMock) GetLimits() (n uint64, p time.Duration) {
	return s.ElemLimit, s.TimeLimit
}

func (s * ServerMock) Process (ctx context.Context, batch Batch) error {
	return ProcessMock(ctx, batch)
}

func TestClient_ElementLimit (t *testing.T) {
	var elemLimit uint64 = 100
	timeLimit := 200 * time.Millisecond
	serv := ServerMock{
		ElemLimit: elemLimit,
		TimeLimit: timeLimit,
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cli := CreateClient(&serv)
	ProcessMock = func(ctx context.Context, batch Batch) error {
		if len(batch) > int(elemLimit) {
			t.Error("Client sent more items than required")
		}
		return nil
	}
	go cli.Produce()
	time.Sleep(2 * time.Second)
	cancel()
}

func TestClient_TimeLimit (t *testing.T) {
	var elemLimit uint64 = 100
	timeLimit := 1000 * time.Millisecond
	serv := ServerMock{
		ElemLimit: elemLimit,
		TimeLimit: timeLimit,
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cli := CreateClient(&serv)
	lastSent := time.Now()
	first := true
	ProcessMock = func(ctx context.Context, batch Batch) error {
		if !first {
			timePassed := time.Since(lastSent)
			if timePassed >= timeLimit {
				t.Errorf("Client failed the time limit requirement. Time between batches %s", timePassed)
			}
			lastSent = time.Now()
		}
		first = false
		return nil
	}
	go cli.Produce()
	time.Sleep(5 * time.Second)
	cancel()
}

func TestClient_ServerBlock (t *testing.T) {
	var elemLimit uint64 = 100
	timeLimit := 1000 * time.Millisecond
	serv := ServerMock{
		ElemLimit: elemLimit,
		TimeLimit: timeLimit,
	}
	cli := CreateClient(&serv)
	ProcessMock = func(ctx context.Context, batch Batch) error {
		return errors.New("blocked")
	}
	err := cli.Produce()
	if err == nil {
		t.Error("Client didn't catch the error from server")
	}
}