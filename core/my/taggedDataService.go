package my

import (
	"fmt"
	"time"

	"github.com/iotaledger/inx-app/pkg/nodebridge"
	iotago "github.com/iotaledger/iota.go/v3"
	builder "github.com/iotaledger/iota.go/v3/builder"

	"github.com/iotaledger/inx-app/pkg/pow"

	"github.com/iotaledger/hive.go/serializer/v2"
)

type (
ProtocolParametersFunc = func() *iotago.ProtocolParameters
)


func createTaggedDataBlock(nodeBridge *nodebridge.NodeBridge) (iotago.BlockID, error) {
	tag := "T1"

	tagBytes := []byte(tag)
	if len(tagBytes) > iotago.MaxTagLength {
		tagBytes = tagBytes[:iotago.MaxTagLength]
	}

	messageString := "Hello INX Plugin"
	payload := &iotago.TaggedData{Tag: tagBytes, Data: []byte(messageString)}

	context := CoreComponent.Daemon().ContextStopped()

	tips, err := nodeBridge.RequestTips(context, 4, true)
	if (err != nil) {
		CoreComponent.LogError("Error while getting tips: %w", err);
		return iotago.EmptyBlockID(), err;
	} 

	block, err := builder.
		NewBlockBuilder().
		ProtocolVersion(nodeBridge.ProtocolParameters().Version).
		Payload(payload).
		Parents(tips).
		Build()

	pow.DoPoW(context, block, serializer.DeSeriModePerformLexicalOrdering, 
		nodeBridge.ProtocolParameters(), 1, 5 * time.Second,
		 func() (tips iotago.BlockIDs, err error) {
			return;
		},
	);
	
	blockId, err := nodeBridge.SubmitBlock(context, block)

	if err != nil {
		return iotago.EmptyBlockID(), fmt.Errorf("build block failed, error: %w", err)
	}

	return blockId, nil
}
