package globalflow

import (
	"context"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type Raft struct {
}

// StartRaft starts the raft node.
func (server *Server) StartRaft(ctx context.Context) error {
	go func(ctx context.Context) {
		if err := server.runRaft(ctx); err != nil {
			logrus.WithError(err).WithError(err).Error("failed to run raft")
		}
	}(ctx)

	return nil
}

// runRaft runs the raft node.
func (server *Server) runRaft(ctx context.Context) error {
	logrus.Info("Running raft")

	votes := 0
	term := 0
	leader := ""

	for {
		switch server.State {
		case NodeStateFollower:
			logrus.WithField("leader", leader).WithField("term", term).Info("Running as follower")

			select {
			case v := <-server.Channels.Heartbeat:
				logrus.Debug("Received heartbeat")
				leader = v.NodeID
				term = v.Term

			case <-time.After(10 * time.Second):
				server.State = NodeStateCandidate

				votes = 1
			case v := <-server.Channels.RequestVote:
				logrus.Debug("Received vote request")

				if v.Term > term {
					server.Broadcast(ctx, VoteGrantedMessage{NodeID: server.Container.Configuration.NodeID, Term: v.Term})
				}
			}

		case NodeStateCandidate:
			logrus.WithField("quorum", server.Quorum()).Info("Starting election as candidate")

			if err := server.Broadcast(ctx, RequestVoteMessage{NodeID: server.Container.Configuration.NodeID, Term: term + 1}); err != nil {
				logrus.WithError(err).Warn("failed to broadcast vote request")

				server.State = NodeStateFollower

				break
			}

			if votes >= server.Quorum() {
				server.State = NodeStateLeader
				leader = server.Container.Configuration.NodeID
				term = term + 1

				break
			}

			select {
			case <-time.After(time.Duration(rand.Intn(10)+3) * time.Second):
				logrus.Info("Election timeout")

				server.State = NodeStateFollower

			case v := <-server.Channels.Heartbeat:
				logrus.Debug("Received heartbeat")

				//if v.Term > term {
				server.State = NodeStateFollower
				leader = v.NodeID
				term = v.Term

				break
				//}

			case <-server.Channels.VoteGranted:
				votes++

				logrus.WithField("quorum", server.Quorum()).WithField("votes", votes).Debug("Received vote")

				if votes >= server.Quorum() {
					server.State = NodeStateLeader
					leader = server.Container.Configuration.NodeID
					term = term + 1

					break
				}
			}

		case NodeStateLeader:
			logrus.WithField("term", term).Info("Running as leader")

			if err := server.Broadcast(ctx, HeartbeatMessage{Term: term, NodeID: server.Container.Configuration.NodeID}); err != nil {
				logrus.WithError(err).Warning("failed to broadcast heartbeat")
			}

			select {
			case <-time.After(1 * time.Second):
				break

			case v := <-server.Channels.Heartbeat:
				logrus.Debug("Received heartbeat")

				if v.Term > term {
					server.State = NodeStateFollower

					break
				}
			}
		}
	}
}
