package dagstore_event

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	sharedTypes "github.com/filecoin-project/venus/venus-shared/types"
	types2 "github.com/filecoin-project/venus/venus-shared/types/gateway"
	"github.com/ipfs-force-community/venus-gateway/types"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("dagstore")

type DagstoreEventStream struct {
	connLk       sync.RWMutex
	channelStore *channelStore
	cfg          *types.Config
	*types.BaseEventStream
}

func NewDagstoreEventStream(ctx context.Context, cfg *types.Config) *DagstoreEventStream {
	return &DagstoreEventStream{
		connLk:          sync.RWMutex{},
		channelStore:    newChannelStore(),
		cfg:             cfg,
		BaseEventStream: types.NewBaseEventStream(ctx, cfg),
	}
}

func (e *DagstoreEventStream) ListenDagstoreEvent(ctx context.Context, policy *DagstoreRegisterPolicy) (chan *types2.RequestEvent, error) {
	ip, exist := jwtclient.CtxGetTokenLocation(ctx)
	if !exist {
		// return nil, xerrors.Errorf("ip not exist")
	}

	out := make(chan *types2.RequestEvent, e.cfg.RequestQueueSize)
	channel := types.NewChannelInfo(ip, out)

	_ = e.channelStore.addChanel(channel)
	log.Infof("add new connections %s", channel.ChannelId)
	go func() {
		connectBytes, err := json.Marshal(types2.ConnectedCompleted{
			ChannelId: channel.ChannelId,
		})
		if err != nil {
			close(out)
			log.Errorf("marshal failed %v", err)
			return
		}

		out <- &types2.RequestEvent{
			ID:         sharedTypes.NewUUID(),
			Method:     "InitConnect",
			Payload:    connectBytes,
			CreateTime: time.Now(),
			Result:     nil,
		} // not response
		for { // nolint:gosimple
			select {
			case <-ctx.Done():
				_ = e.channelStore.removeChannel(channel)
				log.Infof("remove connections %s", channel.ChannelId)
				return
			}
		}
	}()
	return out, nil
}

func (e *DagstoreEventStream) getChannels() ([]*types.ChannelInfo, error) {
	return e.channelStore.channelList()
}

func (e *DagstoreEventStream) HasResource(ctx context.Context, resourceID string) (bool, error) {
	reqBody := ResourceRequest{
		ResourceID: resourceID,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return false, err
	}

	channels, err := e.getChannels()
	if err != nil {
		return false, err
	}

	var result bool
	err = e.SendRequest(ctx, channels, "HasResource", payload, &result)

	return result, err
}

func (e *DagstoreEventStream) GetResource(ctx context.Context, resourceID string) ([]byte, error) {
	reqBody := ResourceRequest{
		ResourceID: resourceID,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	channels, err := e.getChannels()
	if err != nil {
		return nil, err
	}

	var result []byte
	err = e.SendRequest(ctx, channels, "GetResource", payload, &result)

	return result, err
}
