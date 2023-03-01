package my

/*
import (
	"fmt"
	"time"

	"github.com/iotaledger/inx-app/pkg/nodebridge"
	iotago "github.com/iotaledger/iota.go/v3"
	builder "github.com/iotaledger/iota.go/v3/builder"

	"github.com/iotaledger/inx-app/pkg/pow"

	"github.com/iotaledger/hive.go/serializer/v2"
)


func mintAliasOutput(nodeBridge *nodebridge.NodeBridge) (iotago.AliasID, error) {
	messageString := "Hello INX Plugin"

	targetAliasOuput := &iotago.AliasOutput{
		AliasID:    iotago.AliasID{},
		StateIndex: 0,
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: accountSender.Address()},
			&iotago.GovernorAddressUnlockCondition{Address: accountSender.Address()},
		},
		ImmutableFeatures: iotago.Features{
			&iotago.IssuerFeature{Address: accountSender.Address()},
		},
		StateMetadata: []byte(messageString),
	}

	builder.NewTransactionBuilder().AddInput()

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
*/