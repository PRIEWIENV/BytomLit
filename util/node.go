package util

import (
	"encoding/json"

	"github.com/bytom/errors"
	"github.com/bytom/protocol/bc/types"
)

// Node can invoke the api which provide by the full node server
type Node struct {
	hostPort string
}

// NewNode create a api client with target server
func NewNode(hostPort string) *Node {
	return &Node{hostPort: hostPort}
}

type response struct {
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data"`
	ErrDetail string          `json:"error_detail"`
}

func (n *Node) request(path string, payload []byte, respData interface{}) error {
	resp := &response{}
	if err := Post(n.hostPort+path, payload, resp); err != nil {
		return err
	}

	if resp.Status != "success" {
		return errors.New(resp.ErrDetail)
	}

	return json.Unmarshal(resp.Data, respData)
}

// func (n *Node) getRawBlock(req *getRawBlockReq) (*types.Block, error) {
// 	url := "/get-raw-block"
// 	payload, err := json.Marshal(req)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "json marshal")
// 	}
// 	resp := &api.GetRawBlockResp{}
// 	return resp.RawBlock, n.request(url, payload, resp)
// }

type GetBlockCntResp struct {
	BlockCount int32 `json:"block_count"`
}

func (n *Node) GetBlockCnt() (int32, error) {
	url := "/get-block-count"

	resp := &GetBlockCntResp{}
	return resp.BlockCount, n.request(url, nil, resp)
}

type BlockReq struct {
	BlockHeight uint64 `json:"block_height"`
	BlockHash   string `json:"block_hash"`
}

type GetBlockHeaderResp struct {
	BlockHash *types.BlockHeader `json:"block_header"`
}

func (n *Node) GetBlockHeader(req *BlockReq) (*types.BlockHeader, error) {
	url := "/get-block-header"
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal")
	}

	resp := &GetBlockHeaderResp{}
	return resp.BlockHash, n.request(url, payload, resp)
}
