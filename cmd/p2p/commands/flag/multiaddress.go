package flag

import (
	"github.com/multiformats/go-multiaddr"
)

type MultiAddress struct {
	ma multiaddr.Multiaddr
}

// UnmarshalFlag implements flags.Unmarshaler interface
func (a *MultiAddress) UnmarshalFlag(value string) (err error) {
	a.ma, err = multiaddr.NewMultiaddr(value)
	return
}

// AsMultiaddr returns multiaddr.Multiaddr
func (a *MultiAddress) AsMultiaddr() multiaddr.Multiaddr {
	return a.ma
}
