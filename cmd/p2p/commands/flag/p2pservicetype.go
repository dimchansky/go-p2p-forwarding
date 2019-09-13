package flag

import (
	"fmt"
	"strings"

	"github.com/dimchansky/go-p2p-forwarding/cmd/p2p/commands/types/p2pservice"
)

var (
	p2pServiceTypes      = make(map[string]p2pservice.Type)
	p2pServiceTypeSetStr string
)

func init() {
	keys := make([]string, 0, len(p2pservice.TypeValues()))

	for _, k := range p2pservice.TypeValues() {
		keyStr := strings.ToLower(k.String())
		p2pServiceTypes[keyStr] = k
		keys = append(keys, keyStr)
	}

	p2pServiceTypeSetStr = strings.Join(keys, ", ")
}

type P2PServiceType p2pservice.Type

// UnmarshalFlag implements flags.Unmarshaler interface
func (a *P2PServiceType) UnmarshalFlag(value string) error {
	dataType, ok := p2pServiceTypes[strings.ToLower(value)]
	if !ok {
		return fmt.Errorf("unsupported key type '%v', use one of: %v", value, p2pServiceTypeSetStr)
	}

	*a = P2PServiceType(dataType)

	return nil
}

func (a *P2PServiceType) AsP2PService() p2pservice.Type {
	return p2pservice.Type(*a)
}
