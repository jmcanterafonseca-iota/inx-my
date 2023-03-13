package my

import (
	"net/http"
	"net/url"
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
	ParameterDID          = "DID"

	RouteReadAuditTrail            = "/audit-trails/:" + ParameterAuditTrailID
	RouteCreateAuditTrail          = "/audit-trails"
	RouteCreateIdentity            = "/identities"
	RouteReadDecentralizedIdentity = "/identities/:" + ParameterDID
)

func setupRoutes(e *echo.Echo, ledgerService *ledger.LedgerService) {

	e.GET(RouteReadAuditTrail, func(c echo.Context) error {
		aliasID, err := parseAuditTrailIDParam(c, ParameterAuditTrailID)

		if err != nil {
			return err
		}

		resp, err := readAuditTrail(c, ledgerService, aliasID)

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

	e.GET(RouteReadDecentralizedIdentity, func(c echo.Context) error {
		DID, DIDAlias, err := parseDIDParam(c, ledgerService, ParameterDID)

		if err != nil {
			return err
		}

		resp, err := readIdentity(c, ledgerService, &DIDAlias, DID)

		if err != nil {
			return err
		}

		return httpserver.JSONResponse(c, http.StatusOK, resp)
	})

	e.POST(RouteCreateIdentity, func(c echo.Context) error {
		req := &IdentityCreateRequest{}

		if err := c.Bind(req); err != nil {
			return errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid request, error: %s", err)
		}

		if req.Doc == nil {
			return errors.WithMessage(httpserver.ErrInvalidParameter, "invalid request, no DID Doc provided")
		}

		resp, err := createIdentity(c, ledgerService, req.Doc)
		if err != nil {
			return err
		}

		return httpserver.JSONResponse(c, http.StatusOK, resp)
	})
}

func parseAuditTrailIDParam(c echo.Context, paramName string) (*iotago.AliasID, error) {
	return httpserver.ParseAliasIDParam(c, paramName)
}

func parseDIDParam(c echo.Context, ledgerService *ledger.LedgerService, paramName string) (string, iotago.AliasID, error) {
	DID := c.Param(paramName)

	CoreComponent.LogDebugf("DID %s", DID)

	_, err := url.ParseRequestURI(DID)

	if err != nil {
		return "", iotago.AliasID{}, errors.WithMessagef(httpserver.ErrInvalidParameter,
			"invalid DID: %s, error: %s", DID, err)
	}

	if !strings.HasPrefix(DID, "did:iota:"+string(ledgerService.Bech32HRP())+":") {
		return "", iotago.AliasID{}, errors.WithMessagef(httpserver.ErrInvalidParameter,
			"invalid DID: %s, error: %s", DID, err)
	}

	aliasIDComponents := strings.Split(DID, ":")
	aliasIDHex := aliasIDComponents[len(aliasIDComponents)-1]

	aliasID, err := parseAliasID(aliasIDHex)

	return DID, aliasID, err
}

func parseAliasID(aliasIDHex string) (iotago.AliasID, error) {
	aliasIDBytes, err := iotago.DecodeHex(aliasIDHex)

	if err != nil {
		return iotago.AliasID{}, errors.WithMessagef(httpserver.ErrInvalidParameter,
			"invalid Alias ID: %s, error: %s", aliasIDBytes, err)
	}

	if len(aliasIDBytes) < iotago.AliasIDLength {
		return iotago.AliasID{}, errors.WithMessagef(httpserver.ErrInvalidParameter,
			"invalid Trail ID: %s", aliasIDHex)
	}

	var aliasID iotago.AliasID
	copy(aliasID[:], aliasIDBytes)

	return aliasID, nil
}
