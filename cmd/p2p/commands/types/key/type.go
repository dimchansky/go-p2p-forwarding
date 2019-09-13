//go:generate enumer -type=Type

package key

import "github.com/libp2p/go-libp2p-core/crypto"

type Type int

const (
	// RSA is an enum for the supported RSA key type
	RSA Type = crypto.RSA
	// Ed25519 is an enum for the supported Ed25519 key type
	Ed25519 Type = crypto.Ed25519
	// Secp256k1 is an enum for the supported Secp256k1 key type
	Secp256k1 Type = crypto.Secp256k1
	// ECDSA is an enum for the supported ECDSA key type
	ECDSA Type = crypto.ECDSA
)
