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

type Payload interface{}

type AuditTrailCreateResponse struct {
	AuditTrailID string `json:"auditTrailID"`
}

type AuditTrailCreateRequest struct {
	Data string `json:"data"`
}

type AuditTrailReadResponse struct {
	Data string `json:"data"`
}

type IdentityCreateRequest struct {
	Doc interface{} `json:"doc"`
	Type string `json:"type"`
	Metadata MetadataType `json:"metadata"`
}

type MetadataType struct {
	StateControllerAddress string `json:"stateControllerAddress"`
	FundingAddress string `json:"fundingAddress"`
}

type IdentityCreateResponse struct {
	Id string `json:"id"`
}

type IdentityReadResponse map[string]any
