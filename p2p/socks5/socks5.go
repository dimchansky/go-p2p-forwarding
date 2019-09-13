package socks5

import (
	"context"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/armon/go-socks5"
	"github.com/dimchansky/go-p2p-forwarding/p2p"
	"github.com/dimchansky/go-p2p-forwarding/p2p/async"
	"github.com/dimchansky/go-p2p-forwarding/p2p/logging"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

const ID = "/ipfs/port-forwarding-socks5/0.0.1"

var logger = logging.Logger("socks5")

type Socks5 struct {
	closeOnce sync.Once
	ctx       context.Context
	ctxCancel func()
	wg        sync.WaitGroup

	h              host.Host
	s5             *socks5.Server
	clientPeerAddr peer.AddrInfo
}

func New(ctx context.Context, h host.Host, clientAddr multiaddr.Multiaddr) (*Socks5, error) {
	clientPeerAddr, err := peer.AddrInfoFromP2pAddr(clientAddr)
	if err != nil {
		return nil, err
	}

	s5, err := socks5.New(&socks5.Config{
		Logger: log.New(ioutil.Discard, "", 0),
	})
	if err != nil {
		return nil, err
	}

	socksCtx, ctxCancel := context.WithCancel(ctx)
	socks := &Socks5{
		ctx:            socksCtx,
		ctxCancel:      ctxCancel,
		h:              h,
		s5:             s5,
		clientPeerAddr: *clientPeerAddr,
	}
	h.SetStreamHandler(ID, socks.handleStream)

	socks.keepClientConnectionAsync()

	return socks, nil
}

func (l *Socks5) Close() (err error) {
	l.closeOnce.Do(func() {
		err = l.close()
	})
	return
}

func (l *Socks5) close() error {
	logger.Info("closing listener...")
	defer logger.Info("listener closed.")

	l.h.RemoveStreamHandler(ID)

	// TODO: stop handling all socks5 requests

	l.ctxCancel()
	l.wg.Wait()

	return nil
}

func (l *Socks5) handleStream(remote network.Stream) {
	remoteConn := remote.Conn()
	if remotePeerID := remoteConn.RemotePeer(); l.clientPeerAddr.ID != remotePeerID {
		logger.Warningf("unauthorized peer rejected: %v (%v)", remoteConn.RemotePeer(), remoteConn.RemoteMultiaddr())
		_ = remote.Reset()
		return
	}

	// TODO: save remote stream to reset it on Close()

	if err := l.s5.ServeConn(p2p.NewNetConn(remote)); err != nil {
		logger.Debugf("socks5 serving error: %v", err)
		_ = remote.Reset()
	}
}

func (l *Socks5) keepClientConnectionAsync() {
	async.RunPeriodically(&l.wg, l.ctx, 5*time.Second, func(ctx context.Context) error {
		if err := p2p.EnsureConnectedToPeerWithTimeout(ctx, l.h, l.clientPeerAddr, time.Second*30); err != nil {
			logger.Debugf("failed to connect to client peer: %v", err)
		}
		return nil
	})
}
