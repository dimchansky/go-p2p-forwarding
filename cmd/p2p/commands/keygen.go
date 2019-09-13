package commands

import (
	"fmt"

	"github.com/dimchansky/go-p2p-forwarding/cmd/p2p/commands/flag"
	"github.com/dimchansky/go-p2p-forwarding/p2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

type KeyGenCommand struct {
	File    string       `short:"f" long:"identity" required:"true" description:"Output key file"`
	KeyType flag.KeyType `short:"t" long:"key-type" required:"true" description:"key type; rsa, ed25519, secp256k1 or ecdsa"`
	Bits    int          `short:"b" long:"bits"                     description:"key size in bits (for rsa)" default:"2048"`
}

// Execute implements flags.Commander interface
func (c *KeyGenCommand) Execute(args []string) error {
	priv, pub, err := crypto.GenerateKeyPair(c.KeyType.AsInt(), c.Bits)
	if err != nil {
		return err
	}

	id, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return err
	}

	fmt.Printf("Peer ID: %s\n", id.Pretty())

	return p2p.WriteIdentity(priv, c.File)
}
