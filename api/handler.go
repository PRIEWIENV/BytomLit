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

type bytomResponse struct {
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
	Password	string    `json:"password"`
	Tx     		builtTx   `json:"transaction"`
}

type builtTx struct {
	AllowAdditionalActions    bool              `json:"allow_additional_actions"`
	Local                     bool              `json:"local"`
	RawTx                     string            `json:"raw_transaction"`
	SigningIns                []signingInType   `json:"signing_instructions"`
}

type signingInType struct {
	Position       uint64          `json:"position"`
	WitnessComps   []witnessComp   `json:"witness_components"`
}

type witnessComp struct {
	Keys          []key     `json:"keys,omitempty"`
	Quorom        uint64    `json:"quorum,omitempty"`
	Type          string	  `json:"type"`
	Sigs					[]string	`json:"signatures",omitempty`
	Value         string    `json:"value,omitempty"`
}

type key struct {
	DerivPath    []string		`json:"derivation_path"`
	XPub         string			`json:"xpub"`
}

type sendTxReq struct {
	Tx		string		`json:"transaction"`
	Memo	string 		`json:"memo"`
}

type dualFundReq struct {
	FundAssetID   string     `json:"fund_asset_id"`
	FundAmount    uint64     `json:"fund_amount"`
	PeerID        string     `json:"peer_id"`
	PeerAssetID   string     `json:"peer_asset_id"`
	PeerAmount    uint64     `json:"peer_amount"`
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

type signTxResp struct {
	SignComplete	bool			`json:"sign_complete"`
	Tx						builtTx		`json:"transaction"`
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
	Status string `json:"status"`
}

// BuildTx builds unsigned raw transactions
func (s *Server) BuildTx(c *gin.Context, req *buildTxReq) (*buildTxResp, error) {
	return BuildTx(req)
}

// DualFund makes the funding transaction and put it on Bytom chain
func (s *Server) DualFund(c *gin.Context, req *dualFundReq) (*dualFundResp, error) {
	resp := &dualFundResp{
		// DEBUG: Fake
		TxID: "2c0624a7d251c29d4d1ad14297c69919214e78d995affd57e73fbf84ece316cb",
	}
	// DEBUG: only for test
	prog, err := s.DualFundScript("001400634e3bc1d423520f21f3dec9dc13ee90b8f6bb", "0014e7e89d57c4eac32d0507beb15f83d0d09320a9f6")
	if err != nil {
		return nil, err
	}

	sID1, sID2 := btmBc.Hash{}, btmBc.Hash{}
	sID1.UnmarshalText([]byte("01c6ccc6f522228cd4518bba87e9c43fbf55fdf7eb17f5aa300a037db7dca0cb"))
	sID2.UnmarshalText([]byte("d5156f4477fcb694388e6aed7ca390e5bc81bb725ce7461caa241777c1f62236"))

	buildReq := &buildTxReq{
		Inputs: []inputType{
			inputType{
				SourceID: sID1,
				SourcePos: 1,
				Program: "00148c9d063ff74ee6d9ffa88d83aeb038068366c4c4",
				AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				Amount: 110000000,
			},
			inputType{
				SourceID: sID2,
				SourcePos: 2,
				Program: "00148c9d063ff74ee6d9ffa88d83aeb038068366c4c4",
				AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				Amount: 110000000,
			},
		},
		Outputs: []outputType{
			outputType{
				Program: prog,
				AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				Amount: 100000000,
			},
			outputType{
				Program: prog,
				AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				Amount: 100000000,
			},
		},
	}
	respBuild, errBuild := BuildTx(buildReq)
	if errBuild != nil {
    fmt.Println(errBuild) 
	}
	fmt.Println("raw_tx: ", respBuild.RawTx)

	// s.SignTx()

	return resp, nil
}

// Push makes a payment to the peer
func (s *Server) Push(c *gin.Context, req *pushReq) (*pushResp, error) {
	resp := &pushResp{
		Receipt: "success",
	}
	secret := randentropy.GetEntropyCSPRNG(32)
	secretSha256 := crypto.Sha256(secret)
	fmt.Println("secretSha256: ", secretSha256)
	prog, err := s.CommitScript("001400634e3bc1d423520f21f3dec9dc13ee90b8f6bb", "0014e7e89d57c4eac32d0507beb15f83d0d09320a9f6", string(secretSha256))
	if err != nil {
		return resp, err
	}
	log.Println(prog)

	resp.Receipt = prog
	return resp, nil
}

// CloseChannel closes the designated channel
func (s *Server) CloseChannel(c *gin.Context, req *closeChannelReq) (*closeChannelResp, error) {
	resp := &closeChannelResp{
		Status: "success",
	}

	return resp, nil
}

// SendTx is
func (s *Server) SendTx(req *sendTxReq) (*sendTxResp, error) {
	resp := &sendTxResp{}

	return resp, nil
}

func (s *Server) CommitScript(aPub, bPub, h string) (string, error) {
	reqComp := &compileReq{
		Contract: commitContract,
		Arguments: []compileArg{
			compileArg{
				String: aPub,
			},
			compileArg{
				String: bPub,
			},
			compileArg{
				String: h,
			},
			compileArg{
				Integer: 6,
			},
		},
	}
	respComp, errComp := s.Compile(reqComp)
	if errComp != nil {
		return "", errComp
	}
	valueData, okData := respComp.(map[string]interface{})
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

func (s *Server) SignTx(req *signTxReq) (*signTxResp, error) {
	resp := &signTxResp{}
	s.BytomRPCClient.Call(context.Background(), "/sign-transaction", &req, &resp)

	return resp, nil
}

// Compile compiles contract to program
func (s *Server) Compile(req *compileReq) (interface {}, error) {
	resp := &bytomResponse{}
	s.BytomRPCClient.Call(context.Background(), "/compile", &req, &resp)
	if resp.Status != "success" {
		fmt.Errorf(`got=%#v; Err=%#v`, resp.Status, resp.ErrorDetail)
	}

	return resp.Data, nil
}

func (s *Server) BuildCommitmentTx(req *buildTxReq) (*buildTxResp, error) {
	respBuild := &buildTxResp{}
	return respBuild, nil
}

// func RevokeCommitmentTx() error {
// }

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
		// log.Println("mark", input)
		if input.Arguments != "" {
			if err := json.Unmarshal([]byte(input.Arguments), &args); err != nil {
				return nil, err
			}
		}

		tx.Inputs[i].SetArguments(args)
	}
	
	resp := &buildTxResp{
		RawTx: tx,
	}

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
