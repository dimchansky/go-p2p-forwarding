package commands

import (
	"context"
	"fmt"

	"github.com/dimchansky/go-p2p-forwarding/cmd/p2p/commands/flag"
	"github.com/dimchansky/go-p2p-forwarding/p2p"
	"github.com/dimchansky/go-p2p-forwarding/p2p/socks5"
	"github.com/libp2p/go-libp2p"
)

type Socks5Command struct {
	PrivateKey    *flag.PrivateKey  `long:"identity"                       description:"Identity key file."`
	ClientAddress flag.MultiAddress `long:"client-address" required:"true" description:"Client p2p address to accept connections from."`
}

// Execute implements flags.Commander interface
func (c *Socks5Command) Execute(args []string) error {
	ctx, cancel := context.WithCancel(createCtrlCContext())
	defer cancel()

	var opts []libp2p.Option
	if pk := c.PrivateKey; pk != nil {
		opts = append(opts, libp2p.Identity(pk.AsPrivKey()))
	}

	node, err := p2p.NewNode(ctx, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := node.Close(); cErr != nil && err == nil {
			err = cErr
		}
	}()

	lst, err := socks5.New(ctx, node.Host, c.ClientAddress.AsMultiaddr())
	if err != nil {
		return err
	}
	defer func() {
		if cErr := lst.Close(); cErr != nil && err == nil {
			err = cErr
		}
	}()

	fmt.Println("Socks5 started:", node.ID().Pretty())

	<-ctx.Done()

	return nil
}
