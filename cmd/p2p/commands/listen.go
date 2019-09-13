package commands

import (
	"context"
	"fmt"

	"github.com/dimchansky/go-p2p-forwarding/cmd/p2p/commands/flag"
	"github.com/dimchansky/go-p2p-forwarding/p2p"
	"github.com/dimchansky/go-p2p-forwarding/p2p/listener"
	"github.com/libp2p/go-libp2p"
)

type ListenCommand struct {
	PrivateKey    *flag.PrivateKey  `long:"identity"                       description:"Identity key file."`
	TargetAddress flag.MultiAddress `long:"target-address" required:"true" description:"Target address to forward connections to."`
	ClientAddress flag.MultiAddress `long:"client-address" required:"true" description:"Client p2p address to accept connections from."`
}

// Execute implements flags.Commander interface
func (c *ListenCommand) Execute(args []string) error {
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

	lst, err := listener.New(ctx, node, c.TargetAddress.AsMultiaddr(), c.ClientAddress.AsMultiaddr())
	if err != nil {
		return err
	}
	defer func() {
		if cErr := lst.Close(); cErr != nil && err == nil {
			err = cErr
		}
	}()

	fmt.Println("Listener started:", node.ID().Pretty())
	fmt.Println("Connections will be forwarded to:", c.TargetAddress.AsMultiaddr().String())

	<-ctx.Done()

	return nil
}
