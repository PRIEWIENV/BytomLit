package api

import (
	"encoding/hex"
	"encoding/json"
	"context"
	"log"
	"fmt"

	btmBc "github.com/bytom/protocol/bc"
	btmTypes "github.com/bytom/protocol/bc/types"
	"github.com/bytom/crypto"
	"github.com/bytom/crypto/randentropy"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Status      string      `json:"status,omitempty"`
	Code        string      `json:"code,omitempty"`
	Msg         string      `json:"msg,omitempty"`
	ErrorDetail string      `json:"error_detail,detail,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

type buildTxReq struct {
	Inputs  []inputType  `json:"inputs"`
	Outputs []outputType `json:"outputs"`
	Memo    string       `json:"memo"`
}

type signTxReq struct {
	Tx     string   `json:"transaction"`
	Memo   string   `json:"memo"`
}

type sendTxReq struct {
	Tx   string `json:"transaction"`
	Memo string `json:"memo"`
}

type dualFundReq struct {
	FundAssetID   string     `json:"asset_id"`
	FundAmount    uint64     `json:"amount"`
	PeerID        string     `json:"peer_id"`
	PeerAssetID   string     `json:"asset_id"`
	PeerAmount    uint64     `json:"amount"`
}

type pushReq struct {
	AssetID   string     `json:"asset_id"`
	Amount    uint64     `json:"amount"`
	PeerID    string     `json:"peer_id"`
}

type compileArg struct {
	Boolean   bool       `json:"boolean,omitempty"`
	Integer   uint64     `json:"integer,omitempty"`
	String    string     `json:"string,omitempty"`
}

type compileReq struct {
	Contract  string       `json:"contract"`
	Arguments []compileArg `json:"args"`
}

type closeChannelReq struct {
}

type inputType struct {
	Program   string     `json:"program"`
	SourceID  btmBc.Hash `json:"source_id"`
	SourcePos uint64     `json:"source_pos"`
	AssetID   string     `json:"asset_id"`
	Amount    uint64     `json:"amount"`
	Arguments string     `json:"arguments"`
}

type outputType struct {
	Program string `json:"program"`
	AssetID string `json:"asset_id"`
	Amount  uint64 `json:"amount"`
}

type buildTxResp struct {
	// TODO: add sign_insts?
	RawTx *btmTypes.Tx `json:"raw_tx"`
}

type sendTxResp struct {
}

type dualFundResp struct {
  TxID string `json:"tx_id"`
}

type pushResp struct {
  Receipt string `json:"receipt"`
}

type compileResp struct {
	Program string `json:"program"`
}

type closeChannelResp struct {
}

// DualFund makes the funding transaction and put it on Bytom chain
func (s *Server) DualFund(c *gin.Context, req *dualFundReq) (*dualFundResp, error) {
	resp := &dualFundResp{}
	// DEBUG: only for test
	prog, err := s.DualFundScript("001400634e3bc1d423520f21f3dec9dc13ee90b8f6bb", "0014e7e89d57c4eac32d0507beb15f83d0d09320a9f6")
	if err != nil {
		return nil, err
	}
	fmt.Println(prog)
	// buildReq := &buildTxReq{
	// 	Inputs: []inputType{
	// 		inputType{
	// 			SourceID: "01c6ccc6f522228cd4518bba87e9c43fbf55fdf7eb17f5aa300a037db7dca0cb",
	// 			SourcePos: 1,
	// 			Program: "00148c9d063ff74ee6d9ffa88d83aeb038068366c4c4",
	// 			AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	// 			Amount: 110000000,
	// 		},
	// 		inputType{
	// 			SourceID: "d5156f4477fcb694388e6aed7ca390e5bc81bb725ce7461caa241777c1f62236",
	// 			SourcePos: 2,
	// 			Program: "00148c9d063ff74ee6d9ffa88d83aeb038068366c4c4",
	// 			AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	// 			Amount: 110000000,
	// 		},
	// 	},
	// 	Outputs: []outputType{
	// 		outputType{
	// 			Program: prog,
	// 			AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	// 			Amount: 100000000,
	// 		},
	// 		outputType{
	// 			Program: prog,
	// 			AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	// 			Amount: 100000000,
	// 		},
	// 	},
	// }
	// respBuild, errBuild := BuildTx(buildReq)
	// if errBuild != nil {
  //   fmt.Println(errBuild) 
	// }
	return resp, nil
}

func (s *Server) DualFundScript(aPub, bPub string) (string, error) {
	req := &compileReq{
		Contract: dualFundContract,
		Arguments: []compileArg{
			compileArg{
				String: aPub,
			},
			compileArg{
				String: bPub,
			},
			compileArg{
				Integer: dualFundInterval,
			},
		},
	}
	resp, err := s.Compile(req)
	if err != nil {
		return "", err
	}
	valueData, okData := resp.(map[string]interface{})
	if !okData {
		fmt.Errorf("It's not ok for type map[string]interface{}")
		return "", nil
	}
	prog := valueData["program"]
	valueProg, okProg := prog.(string)
	if !okProg {
		fmt.Errorf("It's not ok for type string")
		return "", nil
	}
	return valueProg, nil
}

// Compile compiles contract to program
func (s *Server) Compile(req *compileReq) (interface {}, error) {
	resp := &Response{}
	s.BytomRPCClient.Call(context.Background(), "/compile", &req, &resp)
	if resp.Status != "success" {
		fmt.Errorf(`got=%#v; Err=%#v`, resp.Status, resp.ErrorDetail)
	}

	return resp.Data, nil
}

// Push makes a payment to the peer
func (s *Server) Push(c *gin.Context, req *pushReq) (*pushResp, error) {
	resp := &pushResp{}
	s.BytomRPCClient.Call(context.Background(), "compile", &req, &resp)
	//RevokeCommitmentTx()
	//BuildCommitmentTx()

	return resp, nil
}

func BuildCommitmentTx(req *buildTxReq) (*buildTxResp, error) {
	secret := randentropy.GetEntropyCSPRNG(32)
	secretSha256 := crypto.Sha256(secret)
	fmt.Println("secretSha256: ", secretSha256)
	return BuildTx(req)
}

// func RevokeCommitmentTx() error {
// }

// CloseChannel closes the designated channel

func (s *Server) CloseChannel(c *gin.Context, req *closeChannelReq) (*closeChannelResp, error) {
	resp := &closeChannelResp{}

	return resp, nil
}

// BuildTx builds unsigned raw transactions
func (s *Server) BuildTx(c *gin.Context, req *buildTxReq) (*buildTxResp, error) {
	return BuildTx(req)
}

func BuildTx(req *buildTxReq) (*buildTxResp, error) {
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
	for i, input := range req.Inputs {
		var args [][]byte
		if err := json.Unmarshal([]byte(input.Arguments), &args); err != nil {
			return nil, err
		}

		tx.Inputs[i].SetArguments(args)
	}

	resp := &buildTxResp{
		RawTx: tx,
	}

	return resp, nil
}

// SendTx is
func (s *Server) SendTx(c *gin.Context, req *sendTxReq) (*sendTxResp, error) {
	resp := &sendTxResp{}

	return resp, nil
}

func addInput(txData *btmTypes.TxData, input inputType) error {
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

func addOutput(txData *btmTypes.TxData, output outputType) error {
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
