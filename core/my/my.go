package my

import (
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/inx-app/pkg/nodebridge"

	// import implementation.
	_ "golang.org/x/crypto/blake2b"

)

func createProof(c echo.Context, nodeBridge *nodebridge.NodeBridge) (*MessageResponse, error) {
	CoreComponent.LogDebug("Create Proof Function")

	blockId, err := createTaggedDataBlock(nodeBridge)

	if (err != nil) {
		CoreComponent.LogErrorf("Error creating block: %w", err)
		return nil, err
	}
	
	return &MessageResponse{Message: blockId.String()}, nil
}
