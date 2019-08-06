package cpuminer

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bytom/account"
	"github.com/bytom/consensus/difficulty"
	"github.com/bytom/event"
	"github.com/bytom/mining"
	"github.com/bytom/protocol"
	"github.com/bytom/protocol/bc/types"
)

const (
	maxNonce          = ^uint64(0) // 2^64 - 1
	defaultNumWorkers = 1
	hashUpdateSecs    = 1
	logModule         = "cpuminer"
)

// CPUMiner provides facilities for solving blocks (mining) using the CPU in
// a concurrency-safe manner.
type CPUMiner struct {
	sync.Mutex
	chain            *protocol.Chain
	accountManager   *account.Manager
	txPool           *protocol.TxPool
	numWorkers       uint64
	started          bool
	discreteMining   bool
	workerWg         sync.WaitGroup
	updateNumWorkers chan struct{}
	quit             chan struct{}
	eventDispatcher  *event.Dispatcher
}

// solveBlock attempts to find some combination of a nonce, extra nonce, and
// current timestamp which makes the passed block hash to a value less than the
// target difficulty.
func (m *CPUMiner) solveBlock(block *types.Block, ticker *time.Ticker, quit chan struct{}) bool {
	header := &block.BlockHeader
	seed, err := m.chain.CalcNextSeed(&header.PreviousBlockHash)
	if err != nil {
		return false
	}

	for i := uint64(0); i <= maxNonce; i++ {
		select {
		case <-quit:
			return false
		case <-ticker.C:
			if m.chain.BestBlockHeight() >= header.Height {
				return false
			}
		default:
		}

		header.Nonce = i
		headerHash := header.Hash()
		if difficulty.CheckProofOfWork(&headerHash, seed, header.Bits) {
			return true
		}
	}
	return false
}

// generateBlocks is a worker that is controlled by the miningWorkerController.
// It is self contained in that it creates block templates and attempts to solve
// them while detecting when it is performing stale work and reacting
// accordingly by generating a new block template.  When a block is solved, it
// is submitted.
//
// It must be run as a goroutine.
func (m *CPUMiner) generateBlocks(quit chan struct{}) {
	ticker := time.NewTicker(time.Second * hashUpdateSecs)
	defer ticker.Stop()

out:
	for {
		select {
		case <-quit:
			break out
		default:
		}

		block, err := mining.NewBlockTemplate(m.chain, m.txPool, m.accountManager)
		if err != nil {
			log.Errorf("Mining: failed on create NewBlockTemplate: %v", err)
			continue
		}

		if m.solveBlock(block, ticker, quit) {
			if isOrphan, err := m.chain.ProcessBlock(block); err == nil {
				log.WithFields(log.Fields{
					"module":   logModule,
					"height":   block.BlockHeader.Height,
					"isOrphan": isOrphan,
					"tx":       len(block.Transactions),
				}).Info("Miner processed block")

				// Broadcast the block and announce chain insertion event
				if err = m.eventDispatcher.Post(event.NewMinedBlockEvent{Block: *block}); err != nil {
					log.WithFields(log.Fields{"module": logModule, "height": block.BlockHeader.Height, "error": err}).Errorf("Miner fail on post block")
				}
			} else {
				log.WithFields(log.Fields{"module": logModule, "height": block.BlockHeader.Height, "error": err}).Errorf("Miner fail on ProcessBlock")
			}
		}
	}

	m.workerWg.Done()
}

// miningWorkerController launches the worker goroutines that are used to
// generate block templates and solve them.  It also provides the ability to
// dynamically adjust the number of running worker goroutines.
//
// It must be run as a goroutine.
func (m *CPUMiner) miningWorkerController() {
	// launchWorkers groups common code to launch a specified number of
	// workers for generating blocks.
	var runningWorkers []chan struct{}
	launchWorkers := func(numWorkers uint64) {
		for i := uint64(0); i < numWorkers; i++ {
			quit := make(chan struct{})
			runningWorkers = append(runningWorkers, quit)

			m.workerWg.Add(1)
			go m.generateBlocks(quit)
		}
	}

	// Launch the current number of workers by default.
	runningWorkers = make([]chan struct{}, 0, m.numWorkers)
	launchWorkers(m.numWorkers)

out:
	for {
		select {
		// Update the number of running workers.
		case <-m.updateNumWorkers:
			// No change.
			numRunning := uint64(len(runningWorkers))
			if m.numWorkers == numRunning {
				continue
			}

			// Add new workers.
			if m.numWorkers > numRunning {
				launchWorkers(m.numWorkers - numRunning)
				continue
			}

			// Signal the most recently created goroutines to exit.
			for i := numRunning - 1; i >= m.numWorkers; i-- {
				close(runningWorkers[i])
				runningWorkers[i] = nil
				runningWorkers = runningWorkers[:i]
			}

		case <-m.quit:
			for _, quit := range runningWorkers {
				close(quit)
			}
			break out
		}
	}

	m.workerWg.Wait()
}

// Start begins the CPU mining process as well as the speed monitor used to
// track hashing metrics.  Calling this function when the CPU miner has
// already been started will have no effect.
//
// This function is safe for concurrent access.
func (m *CPUMiner) Start() {
	m.Lock()
	defer m.Unlock()

	// Nothing to do if the miner is already running
	if m.started {
		return
	}

	m.quit = make(chan struct{})
	go m.miningWorkerController()

	m.started = true
	log.Infof("CPU miner started")
}

// Stop gracefully stops the mining process by signalling all workers, and the
// speed monitor to quit.  Calling this function when the CPU miner has not
// already been started will have no effect.
//
// This function is safe for concurrent access.
func (m *CPUMiner) Stop() {
	m.Lock()
	defer m.Unlock()

	// Nothing to do if the miner is not currently running
	if !m.started {
		return
	}

	close(m.quit)
	m.started = false
	log.Info("CPU miner stopped")
}

// IsMining returns whether or not the CPU miner has been started and is
// therefore currenting mining.
//
// This function is safe for concurrent access.
func (m *CPUMiner) IsMining() bool {
	m.Lock()
	defer m.Unlock()

	return m.started
}

// SetNumWorkers sets the number of workers to create which solve blocks.  Any
// negative values will cause a default number of workers to be used which is
// based on the number of processor cores in the system.  A value of 0 will
// cause all CPU mining to be stopped.
//
// This function is safe for concurrent access.
func (m *CPUMiner) SetNumWorkers(numWorkers int32) {
	if numWorkers == 0 {
		m.Stop()
	}

	// Don't lock until after the first check since Stop does its own
	// locking.
	m.Lock()
	defer m.Unlock()

	// Use default if provided value is negative.
	if numWorkers < 0 {
		m.numWorkers = defaultNumWorkers
	} else {
		m.numWorkers = uint64(numWorkers)
	}

	// When the miner is already running, notify the controller about the
	// the change.
	if m.started {
		m.updateNumWorkers <- struct{}{}
	}
}

// NumWorkers returns the number of workers which are running to solve blocks.
//
// This function is safe for concurrent access.
func (m *CPUMiner) NumWorkers() int32 {
	m.Lock()
	defer m.Unlock()

	return int32(m.numWorkers)
}

// NewCPUMiner returns a new instance of a CPU miner for the provided configuration.
// Use Start to begin the mining process.  See the documentation for CPUMiner
// type for more details.
func NewCPUMiner(c *protocol.Chain, accountManager *account.Manager, txPool *protocol.TxPool, dispatcher *event.Dispatcher) *CPUMiner {
	return &CPUMiner{
		chain:            c,
		accountManager:   accountManager,
		txPool:           txPool,
		numWorkers:       defaultNumWorkers,
		updateNumWorkers: make(chan struct{}),
		eventDispatcher:  dispatcher,
	}
}
