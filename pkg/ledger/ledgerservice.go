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

type UTXO struct {
	OutputID iotago.OutputID
	Output   iotago.Output
}

type LedgerService struct {
	wallet *hdwallet.HDWallet

	nodeBridge *nodebridge.NodeBridge

	indexerClient nodeclient.IndexerClient

	log *logger.Logger
}

func New(wallet *hdwallet.HDWallet,
	nodeBridge *nodebridge.NodeBridge,
	indexer nodeclient.IndexerClient,
	log *logger.Logger,
) *LedgerService {
	return &LedgerService{
		wallet:        wallet,
		nodeBridge:    nodeBridge,
		indexerClient: indexer,
		log:           log,
	}
}

func (l *LedgerService) MintAlias(context context.Context, data string) (iotago.AliasID, error) {
	var outputToConsume *UTXO

	// Go to the indexer and obtain the first Basic Output that has funds
	ed25519Addr, err := l.wallet.Ed25519AddressFromIndex(0)
	if err != nil {
		l.log.Errorf("Error while obtaining address %w", err)
		return iotago.AliasID{}, err
	}

	bech32Addr := ed25519Addr.Bech32(l.nodeBridge.ProtocolParameters().Bech32HRP)

	query := &nodeclient.BasicOutputsQuery{AddressBech32: bech32Addr}
	basicOutputs, err := l.queryIndexer(context, query)

	if err != nil {
		l.log.Errorf("Error while calling indexer %w", err)
		return iotago.AliasID{}, err
	}

	// Only first output is taken
	outputToConsume = basicOutputs[0]

	// With the output to Consume we have the input that will fund the Alias Output
	input := &builder.TxInput{
		UnlockTarget: ed25519Addr,
		InputID:      outputToConsume.OutputID,
		Input:        outputToConsume.Output,
	}

	// Two outputs have to be defined. The new Alias Output and the remaining funds Output
	targetAliasOutput := &iotago.AliasOutput{
		AliasID:    iotago.AliasID{},
		StateIndex: 0,
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addr},
			&iotago.GovernorAddressUnlockCondition{Address: ed25519Addr},
		},
		ImmutableFeatures: iotago.Features{
			&iotago.IssuerFeature{Address: ed25519Addr},
		},
		StateMetadata: []byte(data),
	}

	outputCost := l.nodeBridge.ProtocolParameters().RentStructure.MinRent(targetAliasOutput)
	targetAliasOutput.Amount = outputCost
	remainder := outputToConsume.Output.Deposit() - outputCost

	_, signer, err := l.wallet.Ed25519AddressAndSigner(0)
	if err != nil {
		l.log.Errorf("Error while obtaining signer %w", err)
		return iotago.AliasID{}, err
	}

	basicOutputRemainder := &iotago.BasicOutput{
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: ed25519Addr},
		},
		Amount: remainder,
	}

	transaction, err := builder.NewTransactionBuilder(l.nodeBridge.ProtocolParameters().NetworkID()).
		AddInput(input).AddOutput(targetAliasOutput).AddOutput(basicOutputRemainder).
		Build(l.nodeBridge.ProtocolParameters(), signer)

	if err != nil {
		l.log.Errorf("Error while preparing transaction: %w", err)
		return iotago.AliasID{}, err
	}

	tips, err := l.nodeBridge.RequestTips(context, 4, true)
	if err != nil {
		l.log.Errorf("Error while getting tips: %w", err)
		return iotago.AliasID{}, err
	}

	block, err := builder.
		NewBlockBuilder().
		ProtocolVersion(l.nodeBridge.ProtocolParameters().Version).
		Payload(transaction).
		Parents(tips).
		Build()

	pow.DoPoW(context, block, serializer.DeSeriModePerformLexicalOrdering,
		l.nodeBridge.ProtocolParameters(), 1, 5*time.Second,
		func() (tips iotago.BlockIDs, err error) {
			return
		},
	)

	blockId, err := l.nodeBridge.SubmitBlock(context, block)

	if err != nil {
		return iotago.AliasID{}, fmt.Errorf("build block failed, error: %w", err)
	}

	l.log.Debugf("Submitted Block ID %s", blockId.String())

	txId, err := transaction.ID()
	if err != nil {
		l.log.Errorf("Error while generating tx ID: %w", err)
		return iotago.AliasID{}, err
	}

	outputID := iotago.OutputIDFromTransactionIDAndIndex(txId, 0)
	return iotago.AliasIDFromOutputID(outputID), nil
}

func (l *LedgerService) ReadAlias(ctx context.Context, aliasID *iotago.AliasID) (*iotago.AliasOutput, error) {
	_, aliasOutput, err := l.indexerClient.Alias(ctx, *aliasID)

	if err != nil {
		l.log.Errorf("Error while calling indexer %w", err)
		return &iotago.AliasOutput{}, err
	}

	return aliasOutput, nil
}

func (l *LedgerService) queryIndexer(ctx context.Context, query nodeclient.IndexerQuery, maxResults ...int) ([]*UTXO, error) {
	result, err := l.indexerClient.Outputs(ctx, query)
	if err != nil {
		return nil, err
	}

	utxos := []*UTXO{}
	var utxosCount int
	for result.Next() {

		outputs, err := result.Outputs()
		if err != nil {
			return nil, err
		}
		outputIDs := result.Response.Items.MustOutputIDs()

		for i := range outputs {
			if (len(maxResults) > 0) && (utxosCount >= maxResults[0]) {
				return utxos, nil
			}

			utxos = append(utxos, &UTXO{
				OutputID: outputIDs[i],
				Output:   outputs[i],
			})

			utxosCount++
		}
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return utxos, nil
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
	if err != nil {
		l.log.Errorf("Error while getting tips: %w", err)
		return iotago.EmptyBlockID(), err
	}

	block, err := builder.
		NewBlockBuilder().
		ProtocolVersion(l.nodeBridge.ProtocolParameters().Version).
		Payload(payload).
		Parents(tips).
		Build()

	pow.DoPoW(context, block, serializer.DeSeriModePerformLexicalOrdering,
		l.nodeBridge.ProtocolParameters(), 1, 5*time.Second,
		func() (tips iotago.BlockIDs, err error) {
			return
		},
	)

	blockId, err := l.nodeBridge.SubmitBlock(context, block)

	if err != nil {
		return iotago.EmptyBlockID(), fmt.Errorf("build block failed, error: %w", err)
	}

	return blockId, nil
}