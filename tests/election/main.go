package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"test/cluster/election"
	"time"
)

func main() {

	cfg := election.Config{
		ZKURL:             "zookeeper.intranet",
		ZKElectionNodeURI: "/master",
		ZKSlaveNodesURI:   "/slaves",
	}

	election, err := election.New(&cfg)
	if err != nil {
		fmt.Println(err.Error())
	}

	election.Start()

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		<-gracefulStop
		fmt.Println("exiting...")
		election.Close()
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
