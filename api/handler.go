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
	// "github.com/bytom/consensus/segwit"
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
	Local                     bool              `json:"local,omitempty"`
	RawTx                     string            `json:"raw_transaction"`
	SigningIns                []signingInType   `json:"signing_instructions"`
	Fee												uint64						`json:"fee,omitempty"`
}

type signingInType struct {
	Position       uint64          `json:"position"`
	WitnessComps   []witnessComp   `json:"witness_components"`
}

type witnessComp struct {
	Keys          []key     `json:"keys,omitempty"`
	Quorom        uint64    `json:"quorum,omitempty"`
	Type          string	  `json:"type"`
	Sigs					[]string	`json:"signatures,omitempty"`
	Value         string    `json:"value,omitempty"`
}

type key struct {
	DerivPath    []string		`json:"derivation_path"`
	XPub         string			`json:"xpub"`
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

type submitTxReq struct {
	RawTx string `json:"raw_transaction"`
}

type buildTxResp struct {
	// TODO: add sign_insts?
	RawTx *btmTypes.Tx `json:"raw_tx"`
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

type submitTxResp struct {
	TxID string `json:"tx_id"`
}

// BuildTx builds unsigned raw transactions
func (s *Server) BuildTx(c *gin.Context, req *buildTxReq) (*buildTxResp, error) {
	return BuildTx(req)
}

// DualFund makes the funding transaction and put it on Bytom chain
func (s *Server) DualFund(c *gin.Context, req *dualFundReq) (*dualFundResp, error) {
	resp := &dualFundResp{}
	// DEBUG: only for test
	// http://47.99.208.8/dashboard/transactions/a20cf80fb9e907826eb6f092ed5df3ec7bf94072ade273a76a272c9108af9129
	prog, err := s.DualFundScript("b7e5e40c0de6d4cd0048968f047f1ed05215e04e03b7ce22f92ade9ff0791c5d", "343132656a747d98a40488fcd68670f6723abb1f29dfaba36a3b6af18c6360d4")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	sID1, sID2, sID3 := btmBc.Hash{}, btmBc.Hash{}, btmBc.Hash{}
	sID1.UnmarshalText([]byte("dcdd21f5775d8205519204a9e8380632ba6d255f5d9f83640f7f00fe0414c942"))
	sID2.UnmarshalText([]byte("dcdd21f5775d8205519204a9e8380632ba6d255f5d9f83640f7f00fe0414c942"))
	sID3.UnmarshalText([]byte("d9b60b8b3d1e3d249b0efaefddd50a2bcc846f5462a2903c228d3c39b6dfcdf3"))

	buildReq := &buildTxReq{
		Inputs: []inputType{
			inputType{
				SourceID: sID1,
				SourcePos: 2,
				Program: "0014f077b8a83998adfa8df7c529e8643cfebce2dff8",
				AssetID: "f08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf",
				Amount: 10000000000000000,
			},
			inputType{
				SourceID: sID2,
				SourcePos: 1,
				Program: "001472e49786aea9ae75a5ec4543259b6d10c2c4f57d",
				AssetID: "f08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf",
				Amount: 10000000000000000,
			},
			inputType{
				SourceID: sID3,
				SourcePos: 0,
				Program: "001472e49786aea9ae75a5ec4543259b6d10c2c4f57d",
				AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				Amount: 41250000000,
			},
		},
		Outputs: []outputType{
			outputType{
				Program: prog,
				AssetID: "f08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf",
				Amount: 10000000000000000,
			},
			outputType{
				Program: prog,
				AssetID: "f08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf",
				Amount: 10000000000000000,
			},
			outputType{
				Program: "0014a796b852f5db234d4450f80260e5640faf3808ce",
				AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				Amount: 41240000000,
			},
		},
	}
	buildResp, buildErr := BuildTx(buildReq)
	if buildErr != nil {
		return nil, buildErr
	}
	rawTxBytes, mErr := buildResp.RawTx.MarshalText()
	if mErr != nil {
		return nil, mErr
	}

	// Sign the built transaction
	key0 := []key{
		key{
			XPub: "e0446ee8c0f0d559d6eeaeddf5b676ff89de4cdd0477f69f2662269f5e4a6e43d7e6aefa9547408839bc6796f9d22e6865c796d652027a966f1da46a34b94e78",
			DerivPath: []string{
				"2c000000",
				"99000000",
				"01000000",
				"00000000",
				"02000000",
			},
		},
	}
	key1 := []key{
		key{
			XPub: "e0446ee8c0f0d559d6eeaeddf5b676ff89de4cdd0477f69f2662269f5e4a6e43d7e6aefa9547408839bc6796f9d22e6865c796d652027a966f1da46a34b94e78",
			DerivPath: []string{
				"2c000000",
				"99000000",
				"01000000",
				"00000000",
				"01000000",
			},
		},
	}
	key2 := []key{
		key{
			XPub: "e0446ee8c0f0d559d6eeaeddf5b676ff89de4cdd0477f69f2662269f5e4a6e43d7e6aefa9547408839bc6796f9d22e6865c796d652027a966f1da46a34b94e78",
			DerivPath: []string{
				"2c000000",
				"99000000",
				"01000000",
				"00000000",
				"01000000",
			},
		},
	}
	witComps0 := []witnessComp{
		witnessComp{
			Keys: key0,
			Sigs: nil,
			Quorom: 1,
			Type: "raw_tx_signature",
		},
		witnessComp{
			Value: "18cd420713da2b5075f5282dc0ab8abd32e0ad0ec611ddc936d866e042973101",
			Type: "data",
		},
	}
	witComps1 := []witnessComp{
		witnessComp{
			Keys: key1,
			Sigs: nil,
			Quorom: 1,
			Type: "raw_tx_signature",
		},
		witnessComp{
			Value: "d1a80162ad4c529000196b1c44d8bcb07b045190779648a1441e31d086d2e71d",
			Type: "data",
		},
	}
	witComps2 := []witnessComp{
		witnessComp{
			Keys: key2,
			Sigs: nil,
			Quorom: 1,
			Type: "raw_tx_signature",
		},
		witnessComp{
			Value: "d1a80162ad4c529000196b1c44d8bcb07b045190779648a1441e31d086d2e71d",
			Type: "data",
		},
	}
	tx := builtTx{
		AllowAdditionalActions: false,
		Local: true,
		RawTx: string(rawTxBytes),
		SigningIns: []signingInType{
			signingInType{
				Position: 0,
				WitnessComps: witComps0,
			},
			signingInType{
				Position: 1,
				WitnessComps: witComps1,
			},
			signingInType{
				Position: 2,
				WitnessComps: witComps2,
			},
		},
	}
	signReq := &signTxReq{
		Password: "12345",
		Tx: tx,
	}

	signResp, signErr := s.SignTx(signReq)
	if !signResp.SignComplete {
		return nil, fmt.Errorf("signing not complete")
	} else if signErr != nil {
		return nil, signErr
	}

	// Submit the signed raw tx
	subReq := &submitTxReq{
		RawTx: signResp.Tx.RawTx,
	}
	subResp, subErr := s.SubmitTx(subReq)
	if subErr != nil {
		return nil, subErr
	}
	resp.TxID = subResp.TxID

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
	prog, err := s.CommitScript("d1a80162ad4c529000196b1c44d8bcb07b045190779648a1441e31d086d2e71d", "18cd420713da2b5075f5282dc0ab8abd32e0ad0ec611ddc936d866e042973101", string(secretSha256))
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
		return "", fmt.Errorf("It's not ok for type map[string]interface{}")
	}
	prog := valueData["program"]
	valueProg, okProg := prog.(string)
	if !okProg {
		return "", fmt.Errorf("It's not ok for type string")
	}
	return valueProg, nil
}

func (s *Server) SignTx(req *signTxReq) (*signTxResp, error) {
	resp := &bytomResponse{
		Data: &signTxResp{},
	}
	s.BytomRPCClient.Call(context.Background(), "/sign-transaction", &req, &resp)
	if resp.Status != "success" {
		return nil, fmt.Errorf(`got=%#v; Err=%#v`, resp.Status, resp.ErrorDetail)
	}

	vResp, ok := resp.Data.(*signTxResp)
	if !ok {
		return nil, fmt.Errorf("It's not ok for type *signTxResp")
	}

	return vResp, nil
}

// SubmitTx submits a raw transaction
func (s *Server) SubmitTx(req *submitTxReq) (*submitTxResp, error) {
	resp := &bytomResponse{
		Data: &submitTxResp{},
	}
	s.BytomRPCClient.Call(context.Background(), "/submit-transaction", &req, &resp)
	if resp.Status != "success" {
		return nil, fmt.Errorf(`got=%#v; Err=%#v`, resp.Status, resp.ErrorDetail)
	}

	vResp, ok := resp.Data.(*submitTxResp)
	if !ok {
		return nil, fmt.Errorf("It's not ok for type *submitTxResp")
	}

	return vResp, nil
}

// Compile compiles contract to program
func (s *Server) Compile(req *compileReq) (interface {}, error) {
	resp := &bytomResponse{}
	s.BytomRPCClient.Call(context.Background(), "/compile", &req, &resp)
	if resp.Status != "success" {
		return nil, fmt.Errorf(`got=%#v; Err=%#v`, resp.Status, resp.ErrorDetail)
	}

	return resp.Data, nil
}

func (s *Server) BuildCommitmentTx(req *buildTxReq) (*buildTxResp, error) {
	buildResp := &buildTxResp{}
	return buildResp, nil
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
