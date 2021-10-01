package main

import (
	"context"
	"fmt"
	"time"
)

type Client struct {
	elemLimit uint64
	timeLimit time.Duration
	serv Server
	processed chan bool
}

func CreateClient (serv Server) Client {
	c := Client{
		elemLimit: serv.elemLimit,
		timeLimit: serv.timeLimit,
		serv:      serv,
		processed: make(chan bool),
	}
	return c
}

func (cli Client) FillBatch(itemChannel chan Item, batchChannel chan Batch, batchSize int) {
	batch := make([]Item, batchSize)
	for i := 0; i < batchSize; i++  {
		batch[i] = <-itemChannel
	}
	batchChannel <- batch
}

func (cli Client) Send () error {
	itemChannel := make(chan Item, cli.elemLimit)
	defer close(itemChannel)
	elemsLeft := cli.elemLimit
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cli.timeLimit)
	for i := 0; uint64(i) < cli.elemLimit; i++ {
		go cli.AddToBatch(itemChannel, Item{num: i})
	}
	for i := 0; uint64(i) < cli.elemLimit; i++ {
		select {
		case <-cli.processed:
			elemsLeft--
			fmt.Printf("Elems left: %d\n", elemsLeft)
			if elemsLeft == 0 {
				batchChannel := make(chan Batch)
				defer close(batchChannel)
				go cli.FillBatch(itemChannel, batchChannel, int(cli.elemLimit))
				batch := <- batchChannel
				cancel()
				con := context.Background()
				err := cli.serv.Process(con, batch)
				if err != nil {
					return err
				}
				return nil			}
		case <-ctx.Done():
			batchChannel := make(chan Batch)
			defer close(batchChannel)
			fmt.Println("Killed by a timeout")
			go cli.FillBatch(itemChannel, batchChannel, int(cli.elemLimit - elemsLeft))
			batch := <- batchChannel
			con := context.Background()
			err := cli.serv.Process(con, batch)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (cli Client) AddToBatch(itemChannel chan Item, item Item) {
	itemChannel <- item
	cli.processed <- true
}