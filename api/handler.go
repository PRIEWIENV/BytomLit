package api

import (
	"encoding/json"
	"log"

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

func (s *Server) BuildTx(c *gin.Context, req *buildTxReq) ([]string, error) {
	if b, err := json.Marshal(req); err == nil {
		log.Println("received req:", string(b))
	}

	return nil, nil
}
