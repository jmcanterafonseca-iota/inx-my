package my

import (
	"encoding/json"

	"reflect"

	"github.com/pkg/errors"

	"encoding/binary"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/jmcanterafonseca-iota/inx-my/pkg/ledger"
	"github.com/labstack/echo/v4"

	// import implementation.
	_ "golang.org/x/crypto/blake2b"
)

const (
	magicString    string = "DID"
	didVersion     uint8  = 1
	didEncoding    uint8  = 0
	didTemplateStr string = "did:0:0"
)

func createAuditTrail(c echo.Context, ledgerService *ledger.LedgerService, data string) (*AuditTrailCreateResponse, error) {
	CoreComponent.LogDebug("Create Audit Trail Function")

	aliasID, err := ledgerService.MintAlias(CoreComponent.Daemon().ContextStopped(), data)

	if err != nil {
		CoreComponent.LogErrorf("Error creating Audit Trail: %w", err)
		return nil, err
	}

	return &AuditTrailCreateResponse{AuditTrailID: aliasID.String()}, nil
}

func readAuditTrail(c echo.Context, ledgerService *ledger.LedgerService, aliasID *iotago.AliasID) (*AuditTrailReadResponse, error) {
	CoreComponent.LogDebug("Read Audit Trail Function")

	aliasOutput, err := ledgerService.ReadAlias(CoreComponent.Daemon().ContextStopped(), aliasID)

	if err != nil {
		CoreComponent.LogErrorf("Error reading Audit Trail: %w", err)
		return nil, err
	}

	return &AuditTrailReadResponse{Data: string(aliasOutput.StateMetadata)}, nil
}

func readIdentity(c echo.Context, ledgerService *ledger.LedgerService, didAlias *iotago.AliasID, did string) (IdentityReadResponse, error) {
	CoreComponent.LogDebug("Read Identity Function")

	offset := 0

	aliasOutput, err := ledgerService.ReadAlias(CoreComponent.Daemon().ContextStopped(), didAlias)

	if err != nil {
		CoreComponent.LogErrorf("Error reading DID: %w", err)
		return IdentityReadResponse(make(map[string]any)), err
	}

	// We need to verify that this is a valid DID
	// First three bytes should be "DID"
	aliasContent := aliasOutput.StateMetadata
	if !(aliasContent[offset] == magicString[0] && aliasContent[offset+1] == magicString[1] &&
		aliasContent[offset+2] == magicString[2]) {
		CoreComponent.LogErrorf("Invalid Alias Content!! for %s", didAlias.String())
		return IdentityReadResponse(make(map[string]any)), errors.New("Invalid_Alias_Content")
	}

	offset += len(magicString)

	if uint8(aliasContent[offset]) != didVersion {
		CoreComponent.LogErrorf("Invalid Alias Content:  DID Version!! for %s", didAlias.String())
		return IdentityReadResponse(make(map[string]any)), errors.New("Invalid_DID_Version")
	}

	offset++

	if uint8(aliasContent[offset]) != didEncoding {
		CoreComponent.LogErrorf("Invalid Alias Content:  DID encoding not supported!! for %s", didAlias.String())
		return IdentityReadResponse(make(map[string]any)), errors.New("DID_Encoding_Not_Supported")
	}
	offset++

	// The first two bytes are the length of the byte array that contains the DID JSON
	length := binary.LittleEndian.Uint16(aliasContent[offset : offset+2])
	offset += 2

	didContent := aliasContent[offset : offset+int(length)]

	var result map[string]any
	json.Unmarshal(didContent, &result)

	parseDIDDocument(&result, did)

	return IdentityReadResponse(result), nil
}

func parseDIDDocument(doc *map[string]any, did string) {
	for key, value := range *doc {
		if reflect.TypeOf(value).Kind() == reflect.String {
			str := value.(string)
			if str == didTemplateStr {
				(*doc)[key] = did
			}
		} else if reflect.TypeOf(value).Kind() == reflect.Map {
			refMap := value.(map[string]any)
			parseDIDDocument(&refMap, did)
		}
	}
}

func validateDIDDocument(doc *map[string]any) {
	
}

func writeDIDDocument(doc *map[string]any) {

}
