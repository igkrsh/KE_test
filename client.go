package main

import (
	"context"
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

func (cli Client) fillBatch(itemChannel chan Item, batchChannel chan Batch, timer * time.Timer, batchSize int) {
	var batch []Item = nil
	for {
		select {
		case item := <-itemChannel:
			batch = append(batch, item)
			if len(batch) == int(cli.elemLimit) {
				batchChannel <- batch
				batch = nil
			}
		case <-timer.C :
			batchChannel <- batch
			batch = nil
			timer = time.NewTimer(cli.timeLimit)
		}
	}
}

func (cli Client) sendBatch(batchChannel chan Batch, errors chan error) {
	for {
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

func (cli Client) addToBatch(itemChannel chan<- Item) {
	for {
		itemChannel <- Item{}
	}
}

func (cli Client) Schedule() error {
	itemChannel := make(chan Item, cli.elemLimit)
	batchChannel := make(chan Batch, 5)
	errors := make(chan error)
	go cli.produce(itemChannel, batchChannel)
	go cli.sendBatch(batchChannel, errors)
	defer func() {
		close(itemChannel)
		close(batchChannel)
		close(errors)
	}()
	for {
		error := <- errors
		if error != nil {
			return error
		}
	}
}

func (cli Client) produce(itemChannel chan Item, batchChannel chan Batch) {
	numOfWorkers := 100000
	timer := time.NewTimer(cli.timeLimit)
	for i := 0; i < numOfWorkers; i++ {
		go cli.addToBatch(itemChannel)
	}
	go cli.fillBatch(itemChannel, batchChannel, timer, int(cli.elemLimit))
}