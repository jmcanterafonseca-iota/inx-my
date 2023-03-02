package my

import (
	"net/http"
	"strings"

	"github.com/jmcanterafonseca-iota/inx-my/pkg/ledger"
	"github.com/labstack/echo/v4"

	"github.com/pkg/errors"

	"github.com/iotaledger/inx-app/pkg/httpserver"
	iotago "github.com/iotaledger/iota.go/v3"
)

const (
	APIRoute = "my/v1"

	// ParameterBlockID is used to identify a block by its ID.
	ParameterAuditTrailID = "AuditTrailID"

	RouteReadAuditTrail   = "/audit-trails/:" + ParameterAuditTrailID
	RouteCreateAuditTrail = "/audit-trails"
)

func setupRoutes(e *echo.Echo, ledgerService *ledger.LedgerService) {

	e.GET(RouteReadAuditTrail, func(c echo.Context) error {
		aliasID, err := parseAuditTrailIDParam(c, ParameterAuditTrailID)

		if err != nil {
			return err
		}

		resp, err := readAuditTrail(c, ledgerService, &aliasID)

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

func parseAuditTrailIDParam(c echo.Context, paramName string) (iotago.AliasID, error) {
	auditTrailIDHex := strings.ToLower(c.Param(paramName))

	CoreComponent.LogDebugf("Audit Trail ID %s", auditTrailIDHex)

	aliasIDBytes, err := iotago.DecodeHex(auditTrailIDHex)
	if err != nil {
		return iotago.AliasID{}, errors.WithMessagef(httpserver.ErrInvalidParameter, 
			"invalid Trail ID: %s, error: %s", auditTrailIDHex, err)
	}

	if len(aliasIDBytes) < iotago.AliasIDLength {
		return iotago.AliasID{}, errors.WithMessagef(httpserver.ErrInvalidParameter, 
			"invalid Trail ID: %s", auditTrailIDHex)
	}

	var aliasID iotago.AliasID
	copy(aliasID[:], aliasIDBytes)

	return aliasID, nil
}
