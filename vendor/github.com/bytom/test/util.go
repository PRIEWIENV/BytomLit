package test

import (
	"context"
	"time"

	"github.com/bytom/account"
	"github.com/bytom/blockchain/pseudohsm"
	"github.com/bytom/blockchain/txbuilder"
	"github.com/bytom/consensus"
	"github.com/bytom/crypto/ed25519/chainkd"
	"github.com/bytom/database"
	"github.com/bytom/event"
	"github.com/bytom/protocol"
	"github.com/bytom/protocol/bc"
	"github.com/bytom/protocol/bc/types"
	"github.com/bytom/protocol/vm"
	dbm "github.com/bytom/database/leveldb"
)

const (
	vmVersion    = 1
	assetVersion = 1
)

// MockChain mock chain with genesis block
func MockChain(testDB dbm.DB) (*protocol.Chain, *database.Store, *protocol.TxPool, error) {
	store := database.NewStore(testDB)
	dispatcher := event.NewDispatcher()
	txPool := protocol.NewTxPool(store, dispatcher)
	chain, err := protocol.NewChain(store, txPool)
	return chain, store, txPool, err
}

// MockUTXO mock a utxo
func MockUTXO(controlProg *account.CtrlProgram) *account.UTXO {
	utxo := &account.UTXO{}
	utxo.OutputID = bc.Hash{V0: 1}
	utxo.SourceID = bc.Hash{V0: 2}
	utxo.AssetID = *consensus.BTMAssetID
	utxo.Amount = 1000000000
	utxo.SourcePos = 0
	utxo.ControlProgram = controlProg.ControlProgram
	utxo.AccountID = controlProg.AccountID
	utxo.Address = controlProg.Address
	utxo.ControlProgramIndex = controlProg.KeyIndex
	utxo.Change = controlProg.Change
	return utxo
}

// MockTx mock a tx
func MockTx(utxo *account.UTXO, testAccount *account.Account) (*txbuilder.Template, *types.TxData, error) {
	txInput, sigInst, err := account.UtxoToInputs(testAccount.Signer, utxo)
	if err != nil {
		return nil, nil, err
	}

	b := txbuilder.NewBuilder(time.Now())
	if err := b.AddInput(txInput, sigInst); err != nil {
		return nil, nil, err
	}
	out := types.NewTxOutput(*consensus.BTMAssetID, 100, []byte{byte(vm.OP_FAIL)})
	if err := b.AddOutput(out); err != nil {
		return nil, nil, err
	}
	return b.Build()
}

// MockSign sign a tx
func MockSign(tpl *txbuilder.Template, hsm *pseudohsm.HSM, password string) (bool, error) {
	err := txbuilder.Sign(nil, tpl, password, func(_ context.Context, xpub chainkd.XPub, path [][]byte, data [32]byte, password string) ([]byte, error) {
		return hsm.XSign(xpub, path, data[:], password)
	})
	if err != nil {
		return false, err
	}
	return txbuilder.SignProgress(tpl), nil
}

// MockBlock mock a block
func MockBlock() *bc.Block {
	return &bc.Block{
		BlockHeader: &bc.BlockHeader{Height: 1},
	}
}
