package my

import (
	"github.com/jmcanterafonseca-iota/inx-my/pkg/ledger"
	"github.com/labstack/echo/v4"

	// import implementation.
	_ "golang.org/x/crypto/blake2b"
)

func createProof(c echo.Context, ledgerService *ledger.LedgerService) (*MessageResponse, error) {
	CoreComponent.LogDebug("Create Proof Function")

	blockId, err := ledgerService.MintAlias(CoreComponent.Daemon().ContextStopped())

	if (err != nil) {
		CoreComponent.LogErrorf("Error creating block: %w", err)
		return nil, err
	}
	
	return &MessageResponse{Message: blockId.String()}, nil
}
