package my

import (
	"net/http"

	"github.com/jmcanterafonseca-iota/inx-my/pkg/ledger"
	"github.com/labstack/echo/v4"

	"github.com/pkg/errors"

	"github.com/iotaledger/inx-app/pkg/httpserver"
	iotago "github.com/iotaledger/iota.go/v3"
)

const (
	APIRoute = "my/v1"

	// ParameterBlockID is used to identify a block by its ID.
	ParameterAuditTrailID = "auditTrailID"

	RouteReadAuditTrail   = "/audit-trails/:" + ParameterAuditTrailID
	RouteCreateAuditTrail = "/audit-trails"
)

func setupRoutes(e *echo.Echo, ledgerService *ledger.LedgerService) {

	e.GET(RouteReadAuditTrail, func(c echo.Context) error {
		req := &AuditTrailReadRequest{}

		if err := c.Bind(req); err != nil {
			return errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid request, error: %s", err)
		}

		CoreComponent.LogDebugf("Audit Trail ID %s", req.AuditTrailID)

		aliasID, err := iotago.DecodeHex(req.AuditTrailID)
		if err != nil {
			return errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid Audit Trail ID, error: %s", err)
		}

		if len(aliasID) != iotago.AliasIDLength {
			return errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid Audit Trail ID")
		}

		resp, err := readAuditTrail(c, ledgerService, (*iotago.AliasID)(aliasID))

		if err != nil {
			return err
		}

		return httpserver.JSONResponse(c, http.StatusOK, resp)
	})


	e.POST(RouteCreateAuditTrail, func(c echo.Context) error {
		req := &AuditTrailCreateRequest{}

		if err := c.Bind(req); err != nil {
			return errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid request, error: %s", err)
		}

		resp, err := createAuditTrail(c, ledgerService, req.Data)
		if err != nil {
			return err
		}

		return httpserver.JSONResponse(c, http.StatusOK, resp)
	})
}
