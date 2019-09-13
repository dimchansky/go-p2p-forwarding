package p2p

import (
	"context"
	"io"
	"sync"

	"github.com/dimchansky/go-p2p-forwarding/p2p/async"
	"github.com/libp2p/go-libp2p-core/network"
	manet "github.com/multiformats/go-multiaddr-net"
)

// FullDuplexCopy copies bytes from local to remote and vice versa
func FullDuplexCopy(ctx context.Context, local manet.Conn, remote network.Stream) {
	var wg sync.WaitGroup

	localRemoteCh := make(chan struct{})
	async.Run(&wg, func() {
		defer close(localRemoteCh)
		_, _ = io.Copy(local, remote)
	})

	remoteLocalCh := make(chan struct{})
	async.Run(&wg, func() {
		defer close(remoteLocalCh)
		_, _ = io.Copy(remote, local)
	})

	select {
	case <-localRemoteCh:
	case <-remoteLocalCh:
	case <-ctx.Done():
	}

	_ = local.Close()
	_ = remote.Reset()

	wg.Wait()
}
