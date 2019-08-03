package lnutil

import (
	"bufio"
	"bytes"

	"github.com/mit-dci/lit/logging"
    "github.com/mit-dci/lit/wire"
)

func PrintTx(tx *wire.MsgTx) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	tx.Serialize(w)
	w.Flush()
	logging.Infof("%x\n", buf.Bytes())
}
