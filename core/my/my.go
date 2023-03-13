package my

import (
	"encoding/json"
	"strings"

	"reflect"

	"time"

	"github.com/pkg/errors"

	"encoding/binary"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/jmcanterafonseca-iota/inx-my/pkg/ledger"
	"github.com/labstack/echo/v4"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"

	// import implementation.
	_ "golang.org/x/crypto/blake2b"

	"golang.org/x/exp/slices"
)

const (
	magicString    string = "DID"
	didVersion     uint8  = 1
	didEncoding    uint8  = 0
	didTemplateStr string = "did:0:0"
	idKey          string = "id"
)

func whiteListedKeys() []string {
	return []string{idKey}
}

func createAuditTrail(c echo.Context, ledgerService *ledger.LedgerService, data string) (*AuditTrailCreateResponse, error) {
	CoreComponent.LogDebug("Create Audit Trail Function")

	aliasID, err := ledgerService.MintAlias(CoreComponent.Daemon().ContextStopped(), []byte(data))

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
	meta := result["meta"].(map[string]any)

	// Adding the stateController and the governorController
	conditions, _ := aliasOutput.Conditions.Set()
	meta["stateControllerAddress"] = conditions.StateControllerAddress().Address.Bech32(ledgerService.Bech32HRP())
	meta["governorAddress"] = conditions.GovernorAddress().Address.Bech32(ledgerService.Bech32HRP())

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

// Redacts the DID as per the IOTA Identity Spec
func prepareForWriteDID(doc *map[string]any) {
	whiteListedKeys := whiteListedKeys()

	for key, value := range *doc {
		redactedKey := slices.Contains(whiteListedKeys, key)

		if redactedKey {
			if reflect.TypeOf(value).Kind() == reflect.String {
				// Only redact it until the fragment
				str := value.(string)
				elements := strings.Split(str, "#")
				var redactedValue = didTemplateStr
				if len(elements) == 2 {
					redactedValue += elements[1]
				}
				(*doc)[key] = redactedValue
			}
		} else if reflect.TypeOf(value).Kind() == reflect.Map {
			refMap := value.(map[string]any)
			prepareForWriteDID(&refMap)
		}
	}
}

func createIdentity(c echo.Context, ledgerService *ledger.LedgerService, doc interface{}, stateController ...string) (string, error) {
	if err := validateDIDDocument(doc); err != nil {
		CoreComponent.LogErrorf("Error validating DID: %w", err)
		return "", err
	}

	CoreComponent.LogDebugf("DID: %v", doc)

	finalDoc := make(map[string]any)
	finalDoc["doc"] = doc
	meta := make(map[string]string)
	t := time.Now()
	meta["created"] = t.Format(time.RFC3339)
	meta["updated"] = meta["created"]
	finalDoc["meta"] = meta

	docAsMap := doc.(map[string]any)
	// The DID is redacted as mandated by the IOTA DID Spec
	prepareForWriteDID(&docAsMap)

	data, err := json.Marshal(finalDoc)
	if err != nil {
		return "", err
	}

	CoreComponent.LogDebugf("Length of data %v", len(data))

	headerLen := len(magicString) + 2
	headerBytes := make([]byte, headerLen)
	copy(headerBytes, []byte(magicString))
	headerBytes[len(magicString)] = didVersion
	headerBytes[len(magicString)+1] = didEncoding

	// 2 bytes for the length of the encoded data
	stateMetadata := make([]byte, headerLen+2+len(data))
	copy(stateMetadata, headerBytes)
	dataLength := make([]byte, 2)
	binary.LittleEndian.PutUint16(dataLength, uint16(len(data)))
	copy(stateMetadata[headerLen:], dataLength)

	copy(stateMetadata[headerLen+2:], data)

	aliasID, err := ledgerService.MintAlias(CoreComponent.Daemon().ContextStopped(), stateMetadata, stateController[0])
	if err != nil {
		return "", err
	}

	CoreComponent.LogDebugf("New DID %s", "did:iota:"+string(ledgerService.Bech32HRP())+":"+aliasID.String())

	return "did:iota:" + string(ledgerService.Bech32HRP()) + ":" + aliasID.String(), nil
}

func validateDIDDocument(doc interface{}) error {
	const schemaURL = "https://raw.githubusercontent.com/jmcanterafonseca-iota/inx-my/main/core/schema/did-document-schema.json"

	schema, err := jsonschema.Compile(schemaURL)
	if err != nil {
		return err
	}

	if err := schema.Validate(doc); err != nil {
		return err
	}

	return nil
}
