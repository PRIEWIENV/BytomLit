package api

import (
	"encoding/hex"
	"encoding/json"
	"log"

	btmBc "github.com/bytom/protocol/bc"
	btmTypes "github.com/bytom/protocol/bc/types"
	"github.com/gin-gonic/gin"
)

type buildTxReq struct {
	Inputs  []io   `json:"inputs"`
	Outputs []io   `json:"outputs"`
	Memo    string `json:"memo"`
}

type io struct {
	Program   string     `json:"program"`
	SourceID  btmBc.Hash `json:"source_id"`  // for input only
	SourcePos uint64     `json:"source_pos"` // for input only
	AssetID   string     `json:"asset_id"`
	Amount    uint64     `json:"amount"`
}

type buildTxResp struct {
	// TODO: add sign_insts?
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

func addInput(txData *btmTypes.TxData, input io) error {
	assetID := &btmBc.AssetID{}
	if err := assetID.UnmarshalText([]byte(input.AssetID)); err != nil {
		return err
	}

	program, err := hex.DecodeString(input.Program)
	if err != nil {
		return err
	}

	txInput := btmTypes.NewSpendInput(nil, input.SourceID, *assetID, input.Amount, input.SourcePos, program)
	txData.Inputs = append(txData.Inputs, txInput)
	return nil
}

func addOutput(txData *btmTypes.TxData, output io) error {
	assetID := &btmBc.AssetID{}
	if err := assetID.UnmarshalText([]byte(output.AssetID)); err != nil {
		return err
	}

	program, err := hex.DecodeString(output.Program)
	if err != nil {
		return err
	}

	out := btmTypes.NewTxOutput(*assetID, output.Amount, program)
	txData.Outputs = append(txData.Outputs, out)
	return nil
}
