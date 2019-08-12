package main

import (
	"sync"

	"github.com/PRIEWIENV/PHTLC/api"
)

func main() {
	go api.NewServer(nil, nil).Run()

	// keep the main func running in case of terminating goroutines
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
