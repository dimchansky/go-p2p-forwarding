package p2p

import (
	"context"
	"sync"
	"time"

	"github.com/dimchansky/go-p2p-forwarding/p2p/async"
	"github.com/dimchansky/go-p2p-forwarding/p2p/logging"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dhtopts "github.com/libp2p/go-libp2p-kad-dht/opts"
)

var logger = logging.Logger("p2p")

func init() {
	crypto.MinRsaKeyBits = 512 // bootstrap workaround: failed to negotiate security protocol: rsa keys must be >= 2048 bits to be useful
}

type Node struct {
	closeOnce sync.Once
	ctx       context.Context
	ctxCancel func()
	wg        sync.WaitGroup

	host.Host
	*dht.IpfsDHT
}

func NewNode(ctx context.Context, opts ...libp2p.Option) (node *Node, err error) {
	nodeCtx, ctxCancel := context.WithCancel(ctx)
	n := &Node{
		ctx:       nodeCtx,
		ctxCancel: ctxCancel,
	}

	opts = append(opts,
		libp2p.NATPortMap(),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(),
		libp2p.Routing(n.routingFactory(nodeCtx, dhtopts.Client(true))),
	)

	n.Host, err = libp2p.New(nodeCtx, opts...)
	if err != nil {
		return
	}
	// close node in case of error
	defer func() {
		if err != nil {
			_ = n.Close()
		}
	}()

	if err := n.bootstrap(); err != nil {
		return nil, err
	}

	node = n
	return
}

func (n *Node) bootstrap() error {
	peerAddrInfos, err := defaultBootstrapPeerAddresses()
	if err != nil {
		return err
	}

	n.addBootstrapNodesAsPermanentToPeerstore(peerAddrInfos)
	n.connectToBootstrapPeers(peerAddrInfos)
	n.keepBootstrapConnectionsAsync(peerAddrInfos)

	if nDht := n.IpfsDHT; nDht != nil {
		return nDht.Bootstrap(n.ctx)
	}

	return nil
}

func (n *Node) addBootstrapNodesAsPermanentToPeerstore(peers []peer.AddrInfo) {
	for _, addrInfo := range peers {
		n.Host.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)
	}
}

func (n *Node) connectToBootstrapPeers(peers []peer.AddrInfo) {
	logger.Info("connecting to bootstrap peers...")

	var wg sync.WaitGroup
	for _, pi := range peers {
		if n.Host.Network().Connectedness(pi.ID) == network.Connected {
			continue
		}

		n.connectToBootstrapPeerAsync(&wg, pi)
	}

	wg.Wait()
}

func (n *Node) connectToBootstrapPeerAsync(wg *sync.WaitGroup, pi peer.AddrInfo) {
	async.Run(wg, func() { n.connectToBootstrapPeer(pi) })
}

func (n *Node) connectToBootstrapPeer(pi peer.AddrInfo) {
	if err := n.connectWithTimeout(pi, 10*time.Second); err != nil {
		logger.Debugf("error connecting to bootstrap peer %s: %s", pi.ID, err.Error())
	} else {
		n.Host.ConnManager().TagPeer(pi.ID, "bootstrap", 1)
	}
}

func (n *Node) keepBootstrapConnectionsAsync(peers []peer.AddrInfo) {
	async.Run(&n.wg, func() { n.keepBootstrapConnections(peers) })
}

func (n *Node) keepBootstrapConnections(peers []peer.AddrInfo) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			n.connectToBootstrapPeers(peers)
		case <-n.ctx.Done():
			// closing node
			return
		}
	}
}

func (n *Node) connectWithTimeout(pi peer.AddrInfo, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(n.ctx, timeout)
	defer cancel()

	return n.Host.Connect(ctx, pi)
}

func (n *Node) routingFactory(ctx context.Context, opts ...dhtopts.Option) func(host.Host) (routing.PeerRouting, error) {
	return func(h host.Host) (routing.PeerRouting, error) {
		dhtInst, err := dht.New(ctx, h, opts...)
		if err != nil {
			return nil, err
		}
		n.IpfsDHT = dhtInst
		return dhtInst, nil
	}
}

func (n *Node) Close() (err error) {
	n.closeOnce.Do(func() {
		err = n.close()
	})
	return
}

func (n *Node) close() error {
	logger.Info("closing node...")

	n.ctxCancel()
	n.wg.Wait()

	return n.Host.Close()
}

func defaultBootstrapPeerAddresses() ([]peer.AddrInfo, error) {
	bootstrapPeers := dht.DefaultBootstrapPeers
	pis := make([]peer.AddrInfo, 0, len(bootstrapPeers))
	for _, a := range bootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(a)
		if err != nil {
			return nil, err
		}
		pis = append(pis, *pi)
	}
	return pis, nil
}
