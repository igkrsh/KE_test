package main

import (
	"context"
	"fmt"
	"time"
)

type Sender interface {
	Schedule() error
	sendBatch(batchChannel chan Batch, errors chan error)
	produce(itemChannel chan Item, processed chan bool, batchChannel chan Batch)
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
	for i := 0; i < 10; i++ {
		batch := <-batchChannel
		con := context.Background()
		err := cli.serv.Process(con, batch)
		if err != nil {
			errors <- err
		}
		errors <- nil
		time.Sleep(cli.timeLimit)
	}
}

func (cli Client) addToBatch(processed chan bool, itemChannel chan Item, item Item) {
	for {
		itemChannel <- item
		processed <- true
	}
}

func (cli Client) Schedule() error {
	itemChannel := make(chan Item, cli.elemLimit)
	batchChannel := make(chan Batch, 5)
	processed := make(chan bool)
	errors := make(chan error)
	go cli.produce(itemChannel, processed, batchChannel)
	go cli.sendBatch(batchChannel, errors)
	for {
		error := <- errors
		if error != nil {
			return error
		}
	}
}

func (cli Client) produce(itemChannel chan Item, processed chan bool, batchChannel chan Batch) {
	for i := 0; i < 10; i++ {
		elemsLeft := cli.elemLimit
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, cli.timeLimit)
		for i := 0; i < 1000; i++ {
			go cli.addToBatch(processed, itemChannel, Item{num: i})
		}
		for i := 0; uint64(i) < cli.elemLimit; i++ {
			select {
			case <-processed:
				elemsLeft--
				if elemsLeft == 0 {
					go cli.fillBatch(itemChannel, batchChannel, int(cli.elemLimit))
					cancel()
					fmt.Printf("Items produced. Batches in channel: %d\n", len(batchChannel))
				}
			case <-ctx.Done():
				go cli.fillBatch(itemChannel, batchChannel, int(cli.elemLimit-elemsLeft))
			}
		}
	}
}