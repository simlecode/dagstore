package api

import (
	"context"

	"github.com/filecoin-project/dagstore"
	event "github.com/filecoin-project/dagstore/event"
	"github.com/filecoin-project/dagstore/shard"
	types "github.com/filecoin-project/venus/venus-shared/types/gateway"
	mh "github.com/multiformats/go-multihash"
)

type IDagStore interface {
	event.IDagstoreEventAPI
	RegisterShard(ctx context.Context, key shard.Key, opts dagstore.RegisterOpts) error     //perm:read
	GetShard(ctx context.Context, key shard.Key, opts dagstore.AcquireOpts) ([]byte, error) //perm:read
	RecoverShard(ctx context.Context, key shard.Key, opts dagstore.RecoverOpts) error       //perm:read
	GetShardInfo(k shard.Key) (dagstore.ShardInfo, error)                                   //perm:read
	AllShardsInfo() ShardsInfo                                                              //perm:read
	ShardsContainingMultihash(ctx context.Context, h mh.Multihash) ([]shard.Key, error)     //perm:read
}

var _ IDagStore = (*IDagStoreStruct)(nil)

type IDagStoreStruct struct {
	Internal struct {
		RegisterShard             func(ctx context.Context, key shard.Key, opts dagstore.RegisterOpts) error
		GetShard                  func(ctx context.Context, key shard.Key, opts dagstore.AcquireOpts) ([]byte, error)
		RecoverShard              func(ctx context.Context, key shard.Key, opts dagstore.RecoverOpts) error
		GetShardInfo              func(k shard.Key) (dagstore.ShardInfo, error)
		AllShardsInfo             func() ShardsInfo
		ShardsContainingMultihash func(ctx context.Context, h mh.Multihash) ([]shard.Key, error)

		ListenDagstoreEvent   func(ctx context.Context, policy event.DagstoreRegisterPolicy) (<-chan *types.RequestEvent, error)
		ResponseDagstoreEvent func(ctx context.Context, resp *types.ResponseEvent) error
	}
}

func (s *IDagStoreStruct) RegisterShard(ctx context.Context, key shard.Key, opts dagstore.RegisterOpts) error {
	return s.Internal.RegisterShard(ctx, key, opts)
}

func (s *IDagStoreStruct) GetShard(ctx context.Context, key shard.Key, opts dagstore.AcquireOpts) ([]byte, error) {
	return s.Internal.GetShard(ctx, key, opts)
}

func (s *IDagStoreStruct) RecoverShard(ctx context.Context, key shard.Key, opts dagstore.RecoverOpts) error {
	return s.Internal.RecoverShard(ctx, key, opts)
}

func (s *IDagStoreStruct) GetShardInfo(k shard.Key) (dagstore.ShardInfo, error) {
	return s.Internal.GetShardInfo(k)
}

func (s *IDagStoreStruct) AllShardsInfo() ShardsInfo {
	return s.Internal.AllShardsInfo()
}

func (s *IDagStoreStruct) ShardsContainingMultihash(ctx context.Context, h mh.Multihash) ([]shard.Key, error) {
	return s.Internal.ShardsContainingMultihash(ctx, h)
}

func (s *IDagStoreStruct) ListenDagstoreEvent(ctx context.Context, policy *event.DagstoreRegisterPolicy) (<-chan *types.RequestEvent, error) {
	return s.Internal.ListenDagstoreEvent(ctx, *policy)
}

func (s *IDagStoreStruct) ResponseDagstoreEvent(ctx context.Context, resp *types.ResponseEvent) error {
	return s.Internal.ResponseDagstoreEvent(ctx, resp)
}
