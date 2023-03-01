package ledger

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/inx-app/pkg/pow"
	iotago "github.com/iotaledger/iota.go/v3"
	builder "github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/jmcanterafonseca-iota/inx-my/pkg/hdwallet"
)

type  LedgerService struct {
	wallet *hdwallet.HDWallet

	nodeBridge *nodebridge.NodeBridge

	indexerClient *nodeclient.IndexerClient

	// Next Output to be consumed
	nextOutput string

	log *logger.Logger
}

func  New(wallet *hdwallet.HDWallet, 
	nodeBridge *nodebridge.NodeBridge, 
	indexer *nodeclient.IndexerClient,
	log *logger.Logger,
) (*LedgerService) {
	return &LedgerService{
		wallet: wallet,
		nodeBridge: nodeBridge,
		indexerClient: indexer,
		log: log,
	}
}

func (l *LedgerService) MintAlias(context context.Context) (*iotago.AliasID) {
	return nil;
}

func (l *LedgerService) AddTaggedData(context context.Context) (iotago.BlockID, error) {
	tag := "T1"

	tagBytes := []byte(tag)
	if len(tagBytes) > iotago.MaxTagLength {
		tagBytes = tagBytes[:iotago.MaxTagLength]
	}

	messageString := "Hello INX Plugin"
	payload := &iotago.TaggedData{Tag: tagBytes, Data: []byte(messageString)}

	tips, err := l.nodeBridge.RequestTips(context, 4, true)
	if (err != nil) {
		l.log.Errorf("Error while getting tips: %w", err);
		return iotago.EmptyBlockID(), err;
	} 

	block, err := builder.
		NewBlockBuilder().
		ProtocolVersion(l.nodeBridge.ProtocolParameters().Version).
		Payload(payload).
		Parents(tips).
		Build()

	pow.DoPoW(context, block, serializer.DeSeriModePerformLexicalOrdering, 
		l.nodeBridge.ProtocolParameters(), 1, 5 * time.Second,
		 func() (tips iotago.BlockIDs, err error) {
			return;
		},
	);
	
	blockId, err := l.nodeBridge.SubmitBlock(context, block)

	if err != nil {
		return iotago.EmptyBlockID(), fmt.Errorf("build block failed, error: %w", err)
	}

	return blockId, nil
}
