package main

import (
	"sync"

	"github.com/PRIEWIENV/PHTLC/api"
	"github.com/PRIEWIENV/PHTLC/config"
)

func main() {
	cfg := config.NewConfig()
	go api.NewServer(nil, cfg).Run()

	// keep the main func running in case of terminating goroutines
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
