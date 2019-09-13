package listener

import (
	"context"
	"sync"
	"time"

	"github.com/dimchansky/go-p2p-forwarding/p2p"
	"github.com/dimchansky/go-p2p-forwarding/p2p/async"
	"github.com/dimchansky/go-p2p-forwarding/p2p/logging"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

const ID = "/ipfs/port-forwarding-listener/0.0.1"

var logger = logging.Logger("listener")

type Listener struct {
	closeOnce sync.Once
	ctx       context.Context
	ctxCancel func()
	wg        sync.WaitGroup

	h              host.Host
	targetAddr     multiaddr.Multiaddr
	clientPeerAddr peer.AddrInfo
}

func New(ctx context.Context, h host.Host, targetAddr multiaddr.Multiaddr, clientAddr multiaddr.Multiaddr) (*Listener, error) {
	clientPeerAddr, err := peer.AddrInfoFromP2pAddr(clientAddr)
	if err != nil {
		return nil, err
	}

	listenerCtx, ctxCancel := context.WithCancel(ctx)
	listener := &Listener{
		ctx:            listenerCtx,
		ctxCancel:      ctxCancel,
		h:              h,
		targetAddr:     targetAddr,
		clientPeerAddr: *clientPeerAddr,
	}

	h.SetStreamHandler(ID, listener.handleStream)
	listener.keepClientConnectionAsync()

	return listener, nil
}

func (l *Listener) Close() (err error) {
	l.closeOnce.Do(func() {
		err = l.close()
	})
	return
}

func (l *Listener) close() error {
	logger.Info("closing listener...")
	defer logger.Info("listener closed.")

	l.h.RemoveStreamHandler(ID)

	l.ctxCancel()
	l.wg.Wait()

	return nil
}

func (l *Listener) handleStream(remote network.Stream) {
	remoteConn := remote.Conn()
	if remotePeerID := remoteConn.RemotePeer(); l.clientPeerAddr.ID != remotePeerID {
		logger.Warningf("unauthorized peer rejected: %v (%v)", remoteConn.RemotePeer(), remoteConn.RemoteMultiaddr())
		_ = remote.Reset()
		return
	}

	local, err := l.dialWithTimeout(l.targetAddr, 30*time.Second)
	if err != nil {
		_ = remote.Reset()
		return
	}

	logger.Debugf("forwarding %v (%v) to %v...", remoteConn.RemotePeer(), remoteConn.RemoteMultiaddr(), local.RemoteAddr())
	defer logger.Debugf("stopped forwarding %v (%v) to %v.", remoteConn.RemotePeer(), remoteConn.RemoteMultiaddr(), local.RemoteAddr())
	p2p.FullDuplexCopy(l.ctx, local, remote)
}

func (l *Listener) dialWithTimeout(target multiaddr.Multiaddr, timeout time.Duration) (manet.Conn, error) {
	ctx, cancel := context.WithTimeout(l.ctx, timeout)
	defer cancel()

	return (&manet.Dialer{}).DialContext(ctx, target)
}

func (l *Listener) keepClientConnectionAsync() {
	async.RunPeriodically(&l.wg, l.ctx, 5*time.Second, func(ctx context.Context) error {
		if err := p2p.EnsureConnectedToPeerWithTimeout(ctx, l.h, l.clientPeerAddr, time.Second*30); err != nil {
			logger.Debugf("failed to connect to client peer: %v", err)
		}
		return nil
	})
}
