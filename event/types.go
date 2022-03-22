package dagstore_event

import (
	"sync"

	sharedTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/venus-gateway/types"
)

type DagstoreRegisterPolicy struct {
}

type ResourceRequest struct {
	ResourceID string
}

type channelStore struct {
	channels map[sharedTypes.UUID]*types.ChannelInfo
	lk       sync.RWMutex
}

func newChannelStore() *channelStore {
	return &channelStore{
		channels: make(map[sharedTypes.UUID]*types.ChannelInfo),
		lk:       sync.RWMutex{},
	}
}

func (cs *channelStore) addChanel(ch *types.ChannelInfo) error {
	cs.lk.Lock()
	defer cs.lk.Unlock()

	cs.channels[ch.ChannelId] = ch
	return nil
}

func (cs *channelStore) removeChannel(ch *types.ChannelInfo) error {
	cs.lk.Lock()
	defer cs.lk.Unlock()

	delete(cs.channels, ch.ChannelId)
	return nil
}

func (cs *channelStore) channelList() ([]*types.ChannelInfo, error) {
	var channels []*types.ChannelInfo
	for _, channel := range cs.channels {
		channels = append(channels, channel)
	}
	return channels, nil
}
