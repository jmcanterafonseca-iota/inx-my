package my

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/inx-app/pkg/httpserver"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/jmcanterafonseca-iota/inx-my/pkg/daemon"
	"github.com/jmcanterafonseca-iota/inx-my/pkg/hdwallet"
	"github.com/jmcanterafonseca-iota/inx-my/pkg/ledger"
)

const indexerPluginAvailableTimeout = 30 * time.Second

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:     "MY",
			Params:   params,
			DepsFunc: func(cDeps dependencies) { deps = cDeps },
			Provide:  provide,
			Run:      run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
)

type dependencies struct {
	dig.In
	NodeBridge    *nodebridge.NodeBridge
	LedgerService *ledger.LedgerService
}

func provide(c *dig.Container) error {
	type ledgerServiceDeps struct {
		dig.In
		NodeBridge      *nodebridge.NodeBridge
	}

	if err := c.Provide(func(deps ledgerServiceDeps) (*ledger.LedgerService, error) {
		CoreComponent.LogInfo("Setting up Ledger Service ...")

		var wallet *hdwallet.HDWallet
		var indexer nodeclient.IndexerClient
		mnemonic, err := loadMnemonicFromEnvironment("INX_MNEMONIC")
		if err != nil {
			CoreComponent.LogErrorfAndExit("mnemonic seed discovery failed, err: %s", err)
		}

		if len(mnemonic) == 0 {
			CoreComponent.LogErrorfAndExit("mnemonic is empty")
		}

		// new HDWallet instance for address derivation
		wallet, err = hdwallet.NewHDWallet(deps.NodeBridge.ProtocolParameters(), mnemonic, "", 0, false)
		if err != nil {
			return nil, err
		}

		ctxIndexer, cancelIndexer := context.WithTimeout(CoreComponent.Daemon().ContextStopped(), indexerPluginAvailableTimeout)
		defer cancelIndexer()

		indexer, err = deps.NodeBridge.Indexer(ctxIndexer)
		if err != nil {
			return nil, err
		}

		return ledger.New(wallet, deps.NodeBridge, indexer, CoreComponent.Logger()), nil

	}); err != nil {
		return err
	}

	return nil
}

func run() error {
	// create a background worker that handles the API
	if err := CoreComponent.Daemon().BackgroundWorker("API", func(ctx context.Context) {
		CoreComponent.LogInfo("Starting API ... done")

		e := httpserver.NewEcho(CoreComponent.Logger(), nil, ParamsRestAPI.DebugRequestLoggerEnabled)

		CoreComponent.LogInfo("Starting API server ...")

		setupRoutes(e, deps.LedgerService)

		go func() {
			CoreComponent.LogInfof("You can now access the API using: http://%s", ParamsRestAPI.BindAddress)
			if err := e.Start(ParamsRestAPI.BindAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
				CoreComponent.LogErrorfAndExit("Stopped REST-API server due to an error (%s)", err)
			}
		}()

		ctxRegister, cancelRegister := context.WithTimeout(ctx, 5*time.Second)

		advertisedAddress := ParamsRestAPI.BindAddress
		if ParamsRestAPI.AdvertiseAddress != "" {
			advertisedAddress = ParamsRestAPI.AdvertiseAddress
		}

		if err := deps.NodeBridge.RegisterAPIRoute(ctxRegister, APIRoute, advertisedAddress); err != nil {
			CoreComponent.LogErrorfAndExit("Registering INX api route failed: %s", err)
		}
		cancelRegister()

		CoreComponent.LogInfo("Starting API server ... done")
		<-ctx.Done()
		CoreComponent.LogInfo("Stopping API ...")

		ctxUnregister, cancelUnregister := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelUnregister()

		//nolint:contextcheck // false positive
		if err := deps.NodeBridge.UnregisterAPIRoute(ctxUnregister, APIRoute); err != nil {
			CoreComponent.LogWarnf("Unregistering INX api route failed: %s", err)
		}

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCtxCancel()

		//nolint:contextcheck // false positive
		if err := e.Shutdown(shutdownCtx); err != nil {
			CoreComponent.LogWarn(err)
		}

		CoreComponent.LogInfo("Stopping API ... done")
	}, daemon.PriorityStopRestAPI); err != nil {
		CoreComponent.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}

// loads Mnemonic phrases from the given environment variable.
func loadMnemonicFromEnvironment(name string) ([]string, error) {
	keys, exists := os.LookupEnv(name)
	if !exists {
		return nil, fmt.Errorf("environment variable '%s' not set", name)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("environment variable '%s' not set", name)
	}

	phrases := strings.Split(keys, " ")

	return phrases, nil
}
