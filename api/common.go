package api

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

const (
	serverLabel  = "server_label"
	reqBodyLabel = "request_body_label"
	dualFundInterval = 10
	dualFundContract = "contract DualFund(publicKey1: PublicKey,  publicKey2: PublicKey,  blockHeight: Integer) locks valueAmount of valueAsset { clause spend(sig1: Signature, sig2: Signature) { verify checkTxMultiSig([publicKey1, publicKey2], [sig1, sig2]) unlock valueAmount of valueAsset } clause cancel(sig1: Signature) { verify above(blockHeight) verify checkTxSig(publicKey1, sig1) unlock valueAmount of valueAsset }}"
	commitContract = "contract Commitment(publicKey1: PublicKey, publicKey2: PublicKey, revocation1: Hash, blockHeight: Integer) locks valueAmount of valueAsset { clause spend(preimage1: String, sig2: Signature) { verify sha256(preimage1) == revocation1 verify checkTxSig(publicKey2, sig2) unlock valueAmount of valueAsset } clause cancel(sig1: Signature) {   verify above(blockHeight) verify checkTxSig(publicKey1, sig1) unlock valueAmount of valueAsset }}"
	htlcContract = "contract HTLC(sender: PublicKey, recipient: PublicKey, blockHeight: Integer, hash: Hash) locks valueAmount of valueAsset { clause complete(preimage: String, sig: Signature) { verify sha256(preimage) == hash verify checkTxSig(recipient, sig) unlock valueAmount of valueAsset } clause cancel(sig: Signature) { verify above(blockHeight) verify checkTxSig(sender, sig) unlock valueAmount of valueAsset  }}"
)

var (
	errorType           = reflect.TypeOf((*error)(nil)).Elem()
	contextType         = reflect.TypeOf((*gin.Context)(nil))
	paginationQueryType = reflect.TypeOf((*PaginationQuery)(nil))
)