package my

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/jmcanterafonseca-iota/inx-my/pkg/ledger"
	"github.com/labstack/echo/v4"

	// import implementation.
	_ "golang.org/x/crypto/blake2b"
)

func createAuditTrail(c echo.Context, ledgerService *ledger.LedgerService, data string) (*AuditTrailCreateResponse, error) {
	CoreComponent.LogDebug("Create Audit Trail Function")

	aliasID, err := ledgerService.MintAlias(CoreComponent.Daemon().ContextStopped(), data)

	if (err != nil) {
		CoreComponent.LogErrorf("Error creating Audit Trail: %w", err)
		return nil, err
	}
	
	return &AuditTrailCreateResponse{AuditTrailID: aliasID.String()}, nil
}

func readAuditTrail(c echo.Context, ledgerService *ledger.LedgerService, aliasID *iotago.AliasID) (*AuditTrailReadResponse, error) {
	CoreComponent.LogDebug("Read Audit Trail Function")

	aliasOutput, err := ledgerService.ReadAlias(CoreComponent.Daemon().ContextStopped(), aliasID)

	if (err != nil) {
		CoreComponent.LogErrorf("Error reading Audit Trail: %w", err)
		return nil, err
	}
	
	return &AuditTrailReadResponse{Data: string(aliasOutput.StateMetadata)}, nil
}
