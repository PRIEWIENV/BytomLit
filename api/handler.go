package api

import (
	"encoding/hex"
	"encoding/json"
	"context"
	"log"
	"fmt"

	btmBc "github.com/bytom/protocol/bc"
	btmTypes "github.com/bytom/protocol/bc/types"
	"github.com/bytom/protocol/vm/vmutil"
	btmCommon "github.com/bytom/common"
	"github.com/bytom/consensus"
	"github.com/bytom/crypto"
	"github.com/bytom/crypto/randentropy"
	"github.com/bytom/crypto/ed25519"
	"github.com/bytom/crypto/ed25519/chainkd"
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
	Inputs 		[]inputType 	`json:"inputs"`
	InputKeys	[]key					`json:"input_keys"`
	APub 			string 				`json:"pubkey_a"`
	BPub 			string 				`json:"pubkey_b"`
	OutProg		string				`json:"output_program"`
}

type pushReq struct {
	Inputs 		[]inputType 	`json:"inputs"`
	InputKeys	[]key					`json:"input_keys"`
	APub 			string 				`json:"pubkey_a"`
	BPub 			string 				`json:"pubkey_b"`
	OutProgs	[]string			`json:"output_programs"`
	Amount		uint64				`json:"amount"`
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
	Receipt string `json:"receipt"`
}

type inputType struct {
	Program   string     `json:"program"`
	SourceID  btmBc.Hash `json:"source_id"`
	SourcePos uint64     `json:"source_pos"`
	AssetID   string     `json:"asset_id"`
	Amount    uint64     `json:"amount"`
	Arguments string     `json:"arguments,omitempty"`
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
  TxID 			string 	`json:"tx_id"`
}

type pushResp struct {
  Receipt 	string 	`json:"receipt"`
}

type compileResp struct {
	Program 	string 	`json:"program"`
}

type closeChannelResp struct {
	Status 		string 	`json:"status"`
}

type submitTxResp struct {
	TxID 			string 	`json:"tx_id"`
}

type dualFundRawType struct {
	OutputPub string
	Input 		inputType
}

type estTxGasReq struct {
	TxTemp 		builtTx `json:"transaction_template"`
}

type estTxGasResp struct {
	TotalNeu		uint64	`json:"total_neu"`
	StorageNeu 	uint64 	`json:"storage_neu"`
	VMNeu 			uint64	`json:"vm_neu"`
}

type XPubDerivation struct {
	Address					string
	Pubkey					ed25519.PublicKey
	ControlProgram	[]byte
}

func XPub(obj []byte) (chainkd.XPub, error) {
	var rs chainkd.XPub
	if len(obj) != 64 {
		return rs, fmt.Errorf("invalid length")
	}
	copy(rs[:], obj[:64])
	return rs, nil
}

func (s *Server) XPubDerive(XPubs []chainkd.XPub, path [][]byte) (*XPubDerivation, error) {
	derivedXPubs := chainkd.DeriveXPubs(XPubs, path)
	derivedPK := derivedXPubs[0].PublicKey()
	pubHash := crypto.Ripemd160(derivedPK)
	currentParam := consensus.NetParams[s.cfg.Mainchain.NetID]
	address, err := btmCommon.NewAddressWitnessPubKeyHash(pubHash, &currentParam)
	if err != nil {
		return nil, err
	}

	control, err := vmutil.P2WPKHProgram([]byte(pubHash))
	if err != nil {
		return nil, err
	}

	return &XPubDerivation{
		Address: address.EncodeAddress(),
		Pubkey: derivedPK,
		ControlProgram: control,
	}, nil
}

func (s *Server) DualFund(c *gin.Context, req *dualFundReq) (*dualFundResp, error) {
	// DEBUG: only for test
	// http://47.99.208.8/dashboard/transactions/a20cf80fb9e907826eb6f092ed5df3ec7bf94072ade273a76a272c9108af9129
	fmt.Printf("%+v\n%+v\n", c, req)
	return s.DualFundRaw(&req.Inputs, &req.InputKeys, req.APub, req.BPub, req.OutProg)
}

// DualFundRaw makes the funding transaction and put it on Bytom chain
func (s *Server) DualFundRaw(
	inputs		*[]inputType,
	inputKeys	*[]key,
	aPub 			string,
	bPub			string,
	outProg 	string,
) (*dualFundResp, error) {
	resp := &dualFundResp{}
	// Compile contract
	prog, err := s.DualFundScript(aPub, bPub)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	estimatedGasFee := uint64(10000000)
	aInput := (*inputs)[0]
	bInput := (*inputs)[1]
	gasInput := (*inputs)[2]
	// Build unsigned Tx
	if aInput.AssetID != bInput.AssetID {
		return nil, fmt.Errorf("different input AssetIDs")
	}
	fundOutput := outputType{
		Program: prog,
		AssetID: aInput.AssetID,
		Amount: aInput.Amount + bInput.Amount,
	}
	gasOutput := outputType{
		Program: outProg,
		AssetID: gasInput.AssetID,
		Amount: gasInput.Amount - estimatedGasFee,
	}

	bTxReq := &buildTxReq{
		Inputs: []inputType{
			aInput,
			bInput,
			gasInput,
		},
		Outputs: []outputType{
			fundOutput,
			gasOutput,
		},
	}
	bTxResp, bTxErr := BuildTx(bTxReq)
	if bTxErr != nil {
		return nil, bTxErr
	}
	rawTxBytes, mErr := bTxResp.RawTx.MarshalText()
	if mErr != nil {
		return nil, mErr
	}

	if len(*inputKeys) != 3 {
		return nil, fmt.Errorf("invalid inputKeys length, suuposed to be 3")
	}

	// Construct builtTx
	var signIns []signingInType
	for i, k := range *inputKeys {
		var pathBytes [][]byte
		for j, str := range k.DerivPath {
			pathBytes[j], err = hex.DecodeString(str)
			if err != nil {
				return nil, err
			}
		}
		
		xpubByte, e := hex.DecodeString(k.XPub)
		if e != nil{
			return nil, e
		}
		xpub, er := XPub(xpubByte)
		if er != nil{
			return nil, er
		}
		xpubs := []chainkd.XPub{xpub}
		deriv, err := s.XPubDerive(xpubs, pathBytes)
		if err != nil {
			return nil, err
		}

		inputPub := hex.EncodeToString(deriv.Pubkey)
		inputWit := []witnessComp{
			witnessComp{
				Keys: []key{k},
				Sigs: nil,
				Quorom: uint64(1),
				Type: "raw_tx_signature",
			},
			witnessComp{
				Value: inputPub,
				Type: "data",
			},
		}
		signIns = append(signIns, signingInType{
			Position: uint64(i),
			WitnessComps: inputWit,
		})
	}
	tx := builtTx{
		AllowAdditionalActions: false,
		Local: true,
		RawTx: string(rawTxBytes),
		SigningIns: signIns,
	}

	// Estimate tx gas
	// estGasResp, gErr := s.EstTxGas(
	// 	&estTxGasReq{
	// 		TxTemp: tx,
	// 	},
	// )
	// if gErr != nil {
	// 	return nil, gErr
	// }

	// Sign tx
	signReq := &signTxReq{
		Password: s.cfg.Wallet.Password,
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
	fmt.Printf("%+v", subReq)
	subResp, subErr := s.SubmitTx(subReq)
	if subErr != nil {
		return nil, subErr
	}
	resp.TxID = subResp.TxID

	return resp, nil
}

// Push makes a payment to the peer
func (s *Server) Push(c *gin.Context, req *pushReq) (*pushResp, error) {
	return s.PushRaw(&req.Inputs, &req.InputKeys, req.APub, req.BPub, req.OutProgs, req.Amount)
}

func (s *Server) PushRaw(
	inputs			*[]inputType,
	inputKeys		*[]key,
	aPub				string,
	bPub 				string,
	outProgs 		[]string,
	amount 			uint64,
) (*pushResp, error) {
	resp := &pushResp{
		Receipt: "success",
	}
	secret := randentropy.GetEntropyCSPRNG(32)
	secretSha256 := crypto.Sha256(secret)
	prog, err := s.CommitScript(aPub, bPub, hex.EncodeToString(secretSha256))
	if err != nil {
		return resp, err
	}
	log.Println(prog)

	estimatedGasFee := uint64(10000000)
	fundInput := (*inputs)[0]
	gasInput := (*inputs)[1]

	// Build unsigned Tx
	aOutput := outputType{
		Program: outProgs[0],
		AssetID: fundInput.AssetID,
		Amount: fundInput.Amount - amount,
	}
	bOutput := outputType{
		Program: outProgs[1],
		AssetID: fundInput.AssetID,
		Amount: fundInput.Amount + amount,
	}
	gasOutput := outputType{
		Program: outProgs[2],
		AssetID: gasInput.AssetID,
		Amount: gasInput.Amount - estimatedGasFee,
	}

	bTxReq := &buildTxReq{
		Inputs: []inputType{
			fundInput,
			gasInput,
		},
		Outputs: []outputType{
			aOutput,
			bOutput,
			gasOutput,
		},
	}
	bTxResp, bTxErr := BuildTx(bTxReq)
	if bTxErr != nil {
		return nil, bTxErr
	}
	rawTxBytes, mErr := bTxResp.RawTx.MarshalText()
	if mErr != nil {
		return nil, mErr
	}

	var inputPubs []string
	for _, key := range *inputKeys {
		var pathBytes [][]byte
		for j, str := range key.DerivPath {
			pathBytes[j], err = hex.DecodeString(str)
			if err != nil {
				return nil, err
			}
		}
		xpubByte, e := hex.DecodeString(key.XPub)
		if e != nil{
			return nil, e
		}
		xpub, er := XPub(xpubByte)
		if er != nil{
			return nil, er
		}
		xpubs := []chainkd.XPub{xpub}
		deriv, err := s.XPubDerive(xpubs, pathBytes)
		if err != nil {
			return nil, err
		}

		inputPubs = append(inputPubs, hex.EncodeToString(deriv.Pubkey))
	}
	fundInputKeys := (*inputKeys)[0:1]
	gasInputKeys := []key{(*inputKeys)[2]}

	fundInputWit := []witnessComp{
		witnessComp{
			Keys: fundInputKeys,
			Sigs: nil,
			Quorom: uint64(2),
			Type: "raw_tx_signature",
		},
		witnessComp{
			Value: inputPubs[0],
			Type: "data",
		},
		witnessComp{
			Value: inputPubs[1],
			Type: "data",
		},
	}
	gasInputWit := []witnessComp{
		witnessComp{
			Keys: gasInputKeys,
			Sigs: nil,
			Quorom: uint64(1),
			Type: "raw_tx_signature",
		},
		witnessComp{
			Value: inputPubs[2],
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
				WitnessComps: fundInputWit,
			},
			signingInType{
				Position: 1,
				WitnessComps: gasInputWit,
			},
		},
	}

	// Estimate tx gas
	// estGasResp, gErr := s.EstTxGas(
	// 	&estTxGasReq{
	// 		TxTemp: tx,
	// 	},
	// )
	// if gErr != nil {
	// 	return nil, gErr
	// }

	// Sign tx
	signReq := &signTxReq{
		Password: s.cfg.Wallet.Password,
		Tx: tx,
	}
	signResp, signErr := s.SignTx(signReq)
	if signErr != nil {
		return nil, signErr
	}
	signReq2 := &signTxReq{
		Password: s.cfg.Wallet.Password,
		Tx: signResp.Tx,
	}
	signResp2, signErr2 := s.SignTx(signReq2)
	if !signResp2.SignComplete {
		return nil, fmt.Errorf("signing not complete")
	} else if signErr2 != nil {
		return nil, signErr
	}

	resp.Receipt = signResp2.Tx.RawTx
	return resp, nil
}

// CloseChannel closes the designated channel
// Receipt: 0701000201b90101b6017423542dade2528182812b199eafedc8cb013f04dcf62ddae0c4ef207bfd4e8af08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf808084fea6dee11101016b5a20343132656a747d98a40488fcd68670f6723abb1f29dfaba36a3b6af18c6360d420b7e5e40c0de6d4cd0048968f047f1ed05215e04e03b7ce22f92ade9ff0791c5d7424537a641b000000537a547a526bae547a547a526c7cad63240000007bcd9f697b7cae7cac00c0c5010440fd083f7923f88d5a3d427e6519d149573d34fcdbf18d583ecd26d5ac2dba198b2cd5a455140dea12746a80df80daf6312173941fe4d4d28aadeb72549f04140240f3fae7a1734cb144c75e5d370ffcf42c746ce9008d0121551751da06d0ebbb30d0886fbf0966b713572350afb10b537a4585252cc2065f9b8dcbd939e2f1c10c2018cd420713da2b5075f5282dc0ab8abd32e0ad0ec611ddc936d866e04297310120d1a80162ad4c529000196b1c44d8bcb07b045190779648a1441e31d086d2e71d0161015f409fa556dad4ab99f1cedf78656b9221231aa70f06cb531e4df068127598582effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8099c4d59901000116001472e49786aea9ae75a5ec4543259b6d10c2c4f57d6302400903027dc48f4352d08169be7cf7d44e6cf5e2f373d9f666bbd4729ee52c6751c69d52dc868992dae1900a814adf3fcbd41a87a64b4a9c677f93119cf7f59c0020d1a80162ad4c529000196b1c44d8bcb07b045190779648a1441e31d086d2e71d020140f08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf808084fea6dee11101160014a796b852f5db234d4450f80260e5640faf3808ce00013effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80ece1d0990101160014a796b852f5db234d4450f80260e5640faf3808ce00
// http://47.99.208.8/dashboard/transactions/a1c6173d238e15f1dd1f589ac371a891612cbe2b9e2ccc41ae898d3f468eda4e
func (s *Server) CloseChannel(c *gin.Context, req *closeChannelReq) (*closeChannelResp, error) {
	resp := &closeChannelResp{
		Status: "fail",
	}
	subTxReq := &submitTxReq{
		RawTx: req.Receipt,
	}
	subTxResp, sErr := s.SubmitTx(subTxReq)
	if sErr != nil {
		return resp, sErr
	}
	resp.Status = subTxResp.TxID

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

func (s *Server) EstTxGas(req *estTxGasReq) (*estTxGasResp, error) {
	resp := &bytomResponse{
		Data: &estTxGasResp{},
	}
	s.BytomRPCClient.Call(context.Background(), "/estimate-transaction-gas", &req, &resp)
	if resp.Status != "success" {
		return nil, fmt.Errorf(`got=%#v; Err=%#v`, resp.Status, resp.ErrorDetail)
	}

	vResp, ok := resp.Data.(*estTxGasResp)
	if !ok {
		return nil, fmt.Errorf("It's not ok for type *signTxResp")
	}

	return vResp, nil
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
	// if b, err := json.Marshal(req); err == nil {
	// 	log.Println("received req:", string(b))
	// }

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
