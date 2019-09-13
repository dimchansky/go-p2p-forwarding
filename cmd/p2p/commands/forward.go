package commands

import (
	"context"
	"fmt"

	"github.com/dimchansky/go-p2p-forwarding/cmd/p2p/commands/flag"
	"github.com/dimchansky/go-p2p-forwarding/cmd/p2p/commands/types/p2pservice"
	"github.com/dimchansky/go-p2p-forwarding/p2p"
	"github.com/dimchansky/go-p2p-forwarding/p2p/forwarder"
	"github.com/dimchansky/go-p2p-forwarding/p2p/listener"
	"github.com/dimchansky/go-p2p-forwarding/p2p/socks5"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type ForwardCommand struct {
	PrivateKey        *flag.PrivateKey    `long:"identity"                       description:"Identity key file."`
	ListenAddress     flag.MultiAddress   `long:"listen-address" required:"true" description:"Listen address to accept incoming connections."`
	TargetAddress     flag.MultiAddress   `long:"target-address" required:"true" description:"Target p2p address to forward connections to."`
	TargetServiceType flag.P2PServiceType `long:"target-service" required:"true" description:"Target service type (socks5, portforwarder)."`
}

// Execute implements flags.Commander interface
func (c *ForwardCommand) Execute(args []string) error {
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

	var targetProtocolID protocol.ID
	switch c.TargetServiceType.AsP2PService() {
	case p2pservice.PortForwarder:
		targetProtocolID = listener.ID
	case p2pservice.Socks5:
		targetProtocolID = socks5.ID
	default:
		return fmt.Errorf("unsupported p2p service type: %v", c.TargetServiceType)
	}

	fwd, err := forwarder.New(ctx, node, c.ListenAddress.AsMultiaddr(), c.TargetAddress.AsMultiaddr(), targetProtocolID)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := fwd.Close(); cErr != nil && err == nil {
			err = cErr
		}
	}()

	fmt.Println("Forwarder started:", node.ID().Pretty())
	fmt.Println("Forwarder listens on:", c.ListenAddress.AsMultiaddr().String())
	fmt.Println("Connections will be forwarded to:", c.TargetAddress.AsMultiaddr().String())

	<-ctx.Done()

	return nil
}
