package main

import (
	"context"
	"fmt"
	"time"
)

type Sender interface {
	SendBatch() error
	fillBatch(itemChannel chan Item, batchChannel chan Batch, batchSize int)
	addToBatch(processed chan bool, itemChannel chan Item, item Item)
}

type Client struct {
	elemLimit uint64
	timeLimit time.Duration
	serv * Server
}

func CreateClient (serv * Server) * Client {
	elemLim, timeLim := serv.GetLimits()
	c := Client{
		elemLimit: elemLim,
		timeLimit: timeLim,
		serv:      serv,
	}
	return &c
}

func (cli Client) fillBatch(itemChannel chan Item, batchChannel chan Batch, batchSize int) {
	batch := make([]Item, batchSize)
	for i := 0; i < batchSize; i++  {
		batch[i] = <-itemChannel
	}
	batchChannel <- batch
}

func (cli Client) sendBatch(batchChannel chan Batch, errors chan error) {
	batch := <- batchChannel
	con := context.Background()
	err := cli.serv.Process(con, batch)
	if err != nil {
		errors <- err
	}
	errors <- nil
}

func (cli Client) Schedule() error {
	itemChannel := make(chan Item, cli.elemLimit)
	batchChannel := make(chan Batch, 3)
	processed := make(chan bool)
	errors := make(chan error)
	elemsLeft := cli.elemLimit
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cli.timeLimit)
	for i := 0; uint64(i) < cli.elemLimit; i++ {
		go cli.addToBatch(processed, itemChannel, Item{num: i})
	}
	_ = time.AfterFunc(cli.timeLimit, func () {cli.sendBatch(batchChannel, errors)})
	for i := 0; uint64(i) < cli.elemLimit; i++ {
		select {
		case <-processed:
			elemsLeft--
			fmt.Printf("Elems left: %d\n", elemsLeft)
			if elemsLeft == 0 {
				go cli.fillBatch(itemChannel, batchChannel, int(cli.elemLimit))
				<-ctx.Done()
				cancel()
			}
		case <-ctx.Done():
			fmt.Println("Killed by a timeout")
			go cli.fillBatch(itemChannel, batchChannel, int(cli.elemLimit - elemsLeft))
		case err := <- errors:
			return err
		}
	}
	return nil
}

func (cli Client) addToBatch(processed chan bool, itemChannel chan Item, item Item) {
	itemChannel <- item
	processed <- true
}