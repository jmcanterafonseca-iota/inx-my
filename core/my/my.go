package my

import (
	"strings"

	"github.com/labstack/echo/v4"

	// import implementation.
	_ "golang.org/x/crypto/blake2b"
)

func createProof(c echo.Context) (*MessageResponse, error) {
	var builder strings.Builder

	builder.WriteString("Hello World")
	
	return &MessageResponse{Message: builder.String()}, nil
}
