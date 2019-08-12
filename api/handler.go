package api

import (
	// "log"

	"github.com/gin-gonic/gin"
)

type buildTxReq struct{}

func (s *Server) BuildTx(c *gin.Context, req *buildTxReq) ([]string, error) {
	// log.Println(req)
	return nil, nil
}
