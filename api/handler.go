package api

import (
	"encoding/json"
	"log"

	btmTypes "github.com/bytom/protocol/bc/types"
	"github.com/gin-gonic/gin"
)

type buildTxReq struct {
	Inputs  []io   `json:"inputs"`
	Outputs []io   `json:"outputs"`
	Memo    string `json:"memo"`
}

type io struct {
	Program string `json:"program"`
	AssetID string `json:"asset_id"`
	Amount  uint64 `json:"amount"`
}

type buildTxResp struct {
	RawTx *btmTypes.Tx `json:"raw_tx"`
}

func (s *Server) BuildTx(c *gin.Context, req *buildTxReq) (*buildTxResp, error) {
	if b, err := json.Marshal(req); err == nil {
		log.Println("received req:", string(b))
	}

	txData := &btmTypes.TxData{Version: 1}
	for _, input := range req.Inputs {
		if err := addInput(txData, input); err != nil {
			return nil, err
		}
	}

	for _, output := range req.Outputs {
		if err := addOutput(txData, output); err != nil {
			return nil, err
		}
	}

	tx := btmTypes.NewTx(*txData)
	// TODO: add witness?
	resp := &buildTxResp{
		RawTx: tx,
	}
	return resp, nil
}

// TODO:
func addInput(txData *btmTypes.TxData, input io) error {
	return nil
}

// TODO:
func addOutput(txData *btmTypes.TxData, output io) error {
	return nil
}
