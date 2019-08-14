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

type sendTxReq struct {
	Tx  string   `json:"transaction"`
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

type sendTxResp struct {

}

type dualFundReq struct {

}

type dualFundResp struct {

}

type pushReq struct {

}

type pushResp struct {

}

// DualFund makes the funding transaction and put it on Bytom chain
func (s *Server) DualFund(c *gin.Context, req *dualFundReq) (*dualFundResp, error) {
	resp := &dualFundResp{
	}
	
	err := s.mainchainRPCClient.Call("dual-fund", req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Push makes a payment to the peer
func (s *Server) Push(c *gin.Context, req *pushReq) (*pushResp, error) {
	resp := &pushResp{
	}

	return resp, nil
}

// CloseChannel closes the designated channel
func (s *Server) CloseChannel(c *gin.Context, req *closeChannelReq) (*closeChannelResp, error) {
	resp := &closeChannelResp{
	}

	return resp, nil
}

// BuildTx builds unsigned raw transactions
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

// SendTx is 
func (s *Server) SendTx(c *gin.Context, req *sendTxReq) (*sendTxResp, error) {
	resp := &sendTxResp{
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
