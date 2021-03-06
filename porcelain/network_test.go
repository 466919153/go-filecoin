package porcelain_test

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/go-filecoin/net"
	. "github.com/filecoin-project/go-filecoin/porcelain"
	th "github.com/filecoin-project/go-filecoin/testhelpers"
)

type ntwkPingPlumbing struct {
	self peer.ID       // pinging this will fail immediately
	rtt  time.Duration // pinging all other ids will resolve after rtt
}

func (npp *ntwkPingPlumbing) NetworkPing(ctx context.Context, pid peer.ID) (<-chan time.Duration, error) {
	if pid == npp.self {
		return nil, net.ErrPingSelf
	}
	c := make(chan time.Duration)

	go func() {
		<-time.After(npp.rtt)
		c <- npp.rtt
	}()
	return c, nil
}

func newNtwkPingPlumbing(rtt time.Duration, self peer.ID) *ntwkPingPlumbing {
	return &ntwkPingPlumbing{
		rtt:  rtt,
		self: self,
	}
}

func TestPingSuccess(t *testing.T) {
	self := th.RequireRandomPeerID(t)
	plumbing := newNtwkPingPlumbing(100*time.Millisecond, self)
	pid := th.RequireRandomPeerID(t)
	ctx := context.Background()

	assert.NoError(t, PingMinerWithTimeout(ctx, pid, time.Second, plumbing))
}

func TestPingSelfFails(t *testing.T) {
	self := th.RequireRandomPeerID(t)
	plumbing := newNtwkPingPlumbing(100*time.Millisecond, self)
	ctx := context.Background()

	assert.Error(t, PingMinerWithTimeout(ctx, self, time.Second, plumbing))
}

func TestPingTimeout(t *testing.T) {
	self := th.RequireRandomPeerID(t)
	plumbing := newNtwkPingPlumbing(300*time.Millisecond, self)
	pid := th.RequireRandomPeerID(t)
	ctx := context.Background()

	assert.Error(t, PingMinerWithTimeout(ctx, pid, 100*time.Millisecond, plumbing))
}
