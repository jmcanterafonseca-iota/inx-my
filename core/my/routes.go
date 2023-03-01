package my

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/inx-app/pkg/httpserver"

	"github.com/iotaledger/inx-app/pkg/nodebridge"

)

const (
	APIRoute = "my/v1"

	// ParameterBlockID is used to identify a block by its ID.
	ParameterBlockID = "blockID"

	RouteCreateProof   = "/create/:" + ParameterBlockID
	RouteValidateProof = "/validate"
)

func setupRoutes(e *echo.Echo, nodeBridge *nodebridge.NodeBridge) {

	e.GET(RouteCreateProof, func(c echo.Context) error {
		resp, err := createProof(c, nodeBridge)

		if err != nil {
			return err
		}

		CoreComponent.LogDebugf(string(nodeBridge.ProtocolParameters().Bech32HRP))

		return httpserver.JSONResponse(c, http.StatusOK, resp)
	})

	/*
	e.POST(RouteValidateProof, func(c echo.Context) error {
		resp, err := validateProof(c)
		if err != nil {
			return err
		}

		return httpserver.JSONResponse(c, http.StatusOK, resp)
	}) */
}
