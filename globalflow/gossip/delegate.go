package gossip

import (
	"encoding/json"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
)

type Delegate struct {
	Metadata    GossipMetadata
	MessageChan chan []byte
}

func (d *Delegate) NodeMeta(limit int) []byte {
	encoded, err := json.Marshal(d.Metadata)
	if err != nil {
		panic(err)
	}

	if len(encoded) > limit {
		panic("metadata too large")
	}

	return encoded
}

func (d *Delegate) NotifyMsg(bytes []byte) {
	logrus.Debug("Received message")

	// Copy the byte slice before sending it to the channel.
	d.MessageChan <- append([]byte{}, bytes...)
}

func (d *Delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return [][]byte{}
}

func (d *Delegate) LocalState(join bool) []byte {
	return []byte{}
}

func (d *Delegate) MergeRemoteState(buf []byte, join bool) {

}

var _ memberlist.Delegate = &Delegate{}
