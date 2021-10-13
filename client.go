package main

import (
	"context"
	"time"
)

type Sender interface {
	Produce() error
	addToBatch()
	sendBatch()
	fillBatch()
}

type Client struct {
	elemLimit uint64
	timeLimit time.Duration
	serv * Service
	itemChannel chan Item    // buffered channel for items
	batchChannel chan Batch  // buffered channel for batches
	errors chan error        // channel for errors from server
}

func CreateClient (serv Service) * Client {
	elemLim, timeLim := serv.GetLimits() // get params from server
	c := Client{
		elemLimit:    elemLim,
		timeLimit:    timeLim,
		serv:         &serv,
		itemChannel:  make(chan Item, elemLim),
		batchChannel: make(chan Batch, 5),
		errors:       make(chan error),
	}
	return &c
}

func (cli * Client) fillBatch() {
	var batch []Item = nil
	timer := time.NewTimer(cli.timeLimit) // timer for stopping generation if the batch is not filled
	for {
		select {
		// get the generated item from channel
		// add item to the batch and then check if the batch is filled
		// if the batch is full, send it to the batch channel, then reset the batch and timer
		case item := <-cli.itemChannel:
			batch = append(batch, item)
			if len(batch) == int(cli.elemLimit) {
				cli.batchChannel <- batch
				batch = nil
				timer = time.NewTimer(cli.timeLimit)
			}
		// if the timer is fired, send the batch, even if it not filled, to the batch channel
		// then reset the timer and the batch
		case <-timer.C :
			cli.batchChannel <- batch
			batch = nil
			timer = time.NewTimer(cli.timeLimit)
		}
	}
}

func (cli * Client) sendBatch() {
	for {
		batch := <-cli.batchChannel // receive the batch from channel
		con := context.TODO()
		err := (*cli.serv).Process(con, batch)
		// check if server executed without errors
		if err != nil {
			cli.errors <- err
		}
		cli.errors <- nil
		time.Sleep(cli.timeLimit) // sleep until the server is ready to receive a new batch
	}
}

func (cli * Client) addToBatch() {
	for {
		cli.itemChannel <- Item{}
	}
}

func (cli * Client) Produce() error {
	numOfWorkers := 1000
	// create async producers which will generate items
	for i := 0; i < numOfWorkers; i++ {
		go cli.addToBatch()
	}
	go cli.fillBatch()
	go cli.sendBatch()
	for {
		err := <- cli.errors
		// check if server returned an error, continue execution if not
		if err != nil {
			return err
		}
	}
}