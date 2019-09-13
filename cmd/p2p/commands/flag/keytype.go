package flag

import (
	"fmt"
	"strings"

	"github.com/dimchansky/go-p2p-forwarding/cmd/p2p/commands/types/key"
)

var (
	keyTypes      = make(map[string]key.Type)
	keyTypeSetStr string
)

func init() {
	keys := make([]string, 0, len(key.TypeValues()))

	for _, k := range key.TypeValues() {
		keyStr := strings.ToLower(k.String())
		keyTypes[keyStr] = k
		keys = append(keys, keyStr)
	}

	keyTypeSetStr = strings.Join(keys, ", ")
}

type KeyType key.Type

// UnmarshalFlag implements flags.Unmarshaler interface
func (a *KeyType) UnmarshalFlag(value string) error {
	dataType, ok := keyTypes[strings.ToLower(value)]
	if !ok {
		return fmt.Errorf("unsupported key type '%v', use one of: %v", value, keyTypeSetStr)
	}

	*a = KeyType(dataType)

	return nil
}

// AsInt returns  int
func (a KeyType) AsInt() int {
	return int(a)
}
