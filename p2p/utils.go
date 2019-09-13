package p2p

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	swarm "github.com/libp2p/go-libp2p-swarm"
)

// EnsureConnectedToPeer ensures host is connected to target peer, if not tries to connect to target peer with removing
// all previously stored addresses and backoff records for the given peer.
func EnsureConnectedToPeer(ctx context.Context, h host.Host, targetPeerAddr peer.AddrInfo) error {
	targetPeerID := targetPeerAddr.ID
	if h.Network().Connectedness(targetPeerID) != network.Connected {
		logger.Debugf("connecting to peer: %v", targetPeerID)

		h.Peerstore().ClearAddrs(targetPeerID)

		if sw, ok := h.Network().(*swarm.Swarm); ok {
			sw.Backoff().Clear(targetPeerID)
		}

		if err := h.Connect(ctx, targetPeerAddr); err != nil {
			logger.Debugf("failed to connect to peer %v: %v", targetPeerID, err)
			return err
		}

		logger.Debugf("successfully connected to peer %v", targetPeerID)
	}

	return nil
}

// EnsureConnectedToPeerWithTimeout does the same as EnsureConnectedToPeer, but cancels operation after timeout.
func EnsureConnectedToPeerWithTimeout(ctx context.Context, h host.Host, targetPeerAddr peer.AddrInfo, timeout time.Duration) error {
	ctx2, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return EnsureConnectedToPeer(ctx2, h, targetPeerAddr)
}
