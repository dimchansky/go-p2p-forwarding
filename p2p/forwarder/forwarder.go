package forwarder

import (
	"context"
	"sync"
	"time"

	"github.com/dimchansky/go-p2p-forwarding/p2p"
	"github.com/dimchansky/go-p2p-forwarding/p2p/async"
	"github.com/dimchansky/go-p2p-forwarding/p2p/logging"
	tec "github.com/jbenet/go-temp-err-catcher"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

var logger = logging.Logger("forwarder")

type Forwarder struct {
	closeOnce sync.Once
	ctx       context.Context
	ctxCancel func()
	wg        sync.WaitGroup

	h                host.Host
	listener         manet.Listener // listener accepts connections
	targetPeerAddr   peer.AddrInfo  // and forwards them to targetPeerAddr
	targetProtocolID protocol.ID    // using specified protocol ID
}

func New(ctx context.Context, h host.Host, bindAddr multiaddr.Multiaddr, targetAddr multiaddr.Multiaddr, protocolID protocol.ID) (forwarder *Forwarder, err error) {
	targetPeerAddr, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		return
	}

	h.Peerstore().AddAddrs(targetPeerAddr.ID, targetPeerAddr.Addrs, peerstore.TempAddrTTL)

	maListener, err := manet.Listen(bindAddr)
	if err != nil {
		return
	}

	forwarderCtx, ctxCancel := context.WithCancel(ctx)
	forwarder = &Forwarder{
		ctx:              forwarderCtx,
		ctxCancel:        ctxCancel,
		h:                h,
		listener:         maListener,
		targetPeerAddr:   *targetPeerAddr,
		targetProtocolID: protocolID,
	}

	forwarder.acceptConnectionsAsync()

	return
}

func (f *Forwarder) Close() (err error) {
	f.closeOnce.Do(func() {
		err = f.close()
	})
	return
}

func (f *Forwarder) close() error {
	logger.Info("closing forwarder...")
	defer logger.Info("forwarder closed.")

	_ = f.listener.Close()

	f.ctxCancel()
	f.wg.Wait()

	return nil
}

func (f *Forwarder) acceptConnectionsAsync() {
	async.Run(&f.wg, f.acceptConnections)
}

func (f *Forwarder) acceptConnections() {
	defer logger.Info("stopped accepting connections...")

	for {
		local, err := f.listener.Accept()
		if err != nil {
			if tec.ErrIsTemporary(err) {
				continue
			}
			return
		}

		f.handleStreamToTargetPeerAsync(local)
	}
}

func (f *Forwarder) handleStreamToTargetPeerAsync(local manet.Conn) {
	async.Run(&f.wg, func() { f.handleStreamToTargetPeer(local) })
}

func (f *Forwarder) handleStreamToTargetPeer(local manet.Conn) {
	remote, err := f.newStreamToTargetPeer()
	if err != nil {
		logger.Warningf("failed to create stream to target peer: %v", err)
		_ = local.Close()
		return
	}

	remoteConn := remote.Conn()
	logger.Debugf("forwarding %v to %v (%v)...", local.RemoteAddr(), remoteConn.RemotePeer(), remoteConn.RemoteMultiaddr())
	defer logger.Debugf("stopped forwarding %v to %v (%v).", local.RemoteAddr(), remoteConn.RemotePeer(), remoteConn.RemoteMultiaddr())
	p2p.FullDuplexCopy(f.ctx, local, remote)
}

func (f *Forwarder) newStreamToTargetPeer() (network.Stream, error) {
	ctx, cancel := context.WithTimeout(f.ctx, time.Second*30)
	defer cancel()

	if err := p2p.EnsureConnectedToPeer(ctx, f.h, f.targetPeerAddr); err != nil {
		return nil, err
	}

	return f.h.NewStream(ctx, f.targetPeerAddr.ID, f.targetProtocolID)
}
