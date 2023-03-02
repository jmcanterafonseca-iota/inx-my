package my

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/merklehasher"
)

type ProofRequestAndResponse struct {
	Milestone *iotago.Milestone   `json:"milestone"`
	Block     *iotago.Block       `json:"block"`
	Proof     *merklehasher.Proof `json:"proof"`
}

type ValidateProofResponse struct {
	Valid bool `json:"valid"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type AuditTrailCreateRequest struct {
	Data string `json:"data"`
}

type AuditTrailReadResponse struct {
	Data string `json:"data"`
}

type AuditTrailReadRequest struct {
	AuditTrailID string
}
