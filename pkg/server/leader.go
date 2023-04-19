package server

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// monitorLeadership monitors the raft agent's leadership channel to detected when this node has been promoted
// to a leader.  After taking over leadership, it begins running the leader loop.  If it loses leadership
// then it kills the leader loop and stops running it.
func (n *server) monitorLeadership(ctx context.Context) error {

	leaderCh := n.raftNotifyCh
	var lCtx context.Context
	var lCan context.CancelFunc
	var leaderLoop sync.WaitGroup
	for {
		select {
		case isLeader := <-leaderCh:
			switch {
			case isLeader.bool:
				if lCtx != nil {
					n.logger.Warn("attempted to start the leader loop while running", zap.String("id", n.config.ID()))
					continue
				}
				lCtx, lCan = context.WithCancel(ctx)
				defer lCan()
				leaderLoop.Add(1)
				go func(ctx context.Context) {
					defer leaderLoop.Done()
					n.leaderLoop(ctx, isLeader.uint64)
				}(lCtx)
				n.logger.Info("cluster leadership acquired", zap.String("id", n.config.ID()))

			default:
				if lCtx == nil {
					//n.logger.Warn("attempted to stop the leader loop while not running", zap.String("id", n.config.ID()))
					continue
				}
				n.logger.Info("shutting down leader loop", zap.String("id", n.config.ID()))
				lCan()
				leaderLoop.Wait()
				lCtx = nil
				n.logger.Info("cluster leadership lost", zap.String("id", n.config.ID()))
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// leaderLoop is the code you want the leader to run.  For example, if you need a job
// started ever N mins, this would be a good place to put it.
func (n *server) leaderLoop(ctx context.Context, shardID uint64) error {
	// We are the leader, do leader stuff here.
	<-ctx.Done()
	n.logger.Info("leader loop exiting", zap.Uint64("shard", shardID))
	return nil
}
