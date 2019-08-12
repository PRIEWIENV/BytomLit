package api

import (
	"github.com/gin-gonic/gin"
)

type listCrosschainTxsReq struct{ Display }

func (s *Server) ListCrosschainTxs(c *gin.Context, listTxsReq *listCrosschainTxsReq, query *PaginationQuery) ([]string, error) {
	return nil, nil
}
