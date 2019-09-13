package flag

import (
	"github.com/dimchansky/go-p2p-forwarding/p2p"
	"github.com/libp2p/go-libp2p-core/crypto"
)

type PrivateKey struct {
	k crypto.PrivKey
}

// UnmarshalFlag implements flags.Unmarshaler interface
func (a *PrivateKey) UnmarshalFlag(value string) (err error) {
	a.k, err = p2p.ReadIdentity(value)
	return
}

// AsMultiaddr returns crypto.PrivKey
func (a *PrivateKey) AsPrivKey() crypto.PrivKey {
	return a.k
}
