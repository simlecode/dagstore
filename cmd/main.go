package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/filecoin-project/dagstore"
	"github.com/filecoin-project/dagstore/api"
	event "github.com/filecoin-project/dagstore/event"
	"github.com/filecoin-project/dagstore/index"
	"github.com/filecoin-project/dagstore/mount"
	"github.com/filecoin-project/dagstore/shard"
	"github.com/filecoin-project/go-jsonrpc"
	gtypes "github.com/filecoin-project/venus/venus-shared/types/gateway"
	"github.com/gorilla/mux"
	gtypes2 "github.com/ipfs-force-community/venus-gateway/types"
	"github.com/ipfs/go-cid"
	badger "github.com/ipfs/go-ds-badger2"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/multiformats/go-multihash"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var log = logging.Logger("main")

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "repo",
				Value: "./.dagstore",
			},
			&cli.StringFlag{
				Name:  "listen-address",
				Value: "/ip4/127.0.0.1/tcp/7898",
			},
		},
		Commands: []*cli.Command{
			runCmd,
		},
	}

	app.Setup()
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

var runCmd = &cli.Command{
	Name: "run",
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		rootDir := cctx.String("repo")
		idx, err := index.NewFSRepo(rootDir)
		if err != nil {
			return err
		}
		if err := checkRemoteDir(rootDir); err != nil {
			return err
		}

		ds, err := badger.NewDatastore(rootDir, &badger.DefaultOptions)
		if err != nil {
			return xerrors.Errorf("failed to open datastore: %w", err)
		}
		inverted := index.NewInverted(ds)
		traceCh := make(chan dagstore.Trace, 28)
		failureCh := make(chan dagstore.ShardResult, 10)
		mountRegistry := mount.NewRegistry()

		cfg := dagstore.Config{
			TransientsDir:             rootDir,
			IndexRepo:                 idx,
			TopLevelIndex:             inverted,
			Datastore:                 ds,
			MountRegistry:             mountRegistry,
			TraceCh:                   traceCh,
			FailureCh:                 failureCh,
			MaxConcurrentIndex:        10,
			MaxConcurrentReadyFetches: 1,
			RecoverOnStart:            dagstore.RecoverOnAcquire,
		}
		dagStore, err := dagstore.NewDAGStore(cfg)
		if err != nil {
			return err
		}
		if err := dagStore.Start(ctx); err != nil {
			return err
		}

		go traceLoop(ctx, traceCh, failureCh)
		go gcLoop(ctx, dagStore, time.Minute*5)

		e := event.NewDagstoreEventStream(ctx, &gtypes2.Config{})

		rpcServer := jsonrpc.NewServer(jsonrpc.WithServerTimeout(time.Minute * 30))
		rpcServer.Register("DagStore", newDagStoreImp(rootDir, dagStore, mountRegistry, e))
		router := mux.NewRouter()
		router.Handle("/rpc/v0", rpcServer)

		srv := &http.Server{Handler: router}

		go func() {
			select {
			case <-cctx.Context.Done():
			}
			if err := dagStore.Close(); err != nil {
				log.Error(err)
			}
			if err := srv.Shutdown(context.TODO()); err != nil {
				log.Errorf("shutting down RPC server failed: %s", err)
			}
			log.Info("Graceful shutdown successful")
		}()

		addr, err := multiaddr.NewMultiaddr(cctx.String("listen-address"))
		if err != nil {
			return err
		}

		nl, err := manet.Listen(addr)
		if err != nil {
			return err
		}
		fmt.Printf("start rpc listen %s\n", addr)
		return srv.Serve(manet.NetListener(nl))
	},
}

func checkRemoteDir(rootDir string) error {
	remoteDir := filepath.Join(rootDir, "remote")
	f, err := os.Stat(remoteDir)
	if err != nil {
		return os.MkdirAll(remoteDir, os.ModePerm)
	}
	if !f.IsDir() {
		return xerrors.Errorf("%s must a dir", remoteDir)
	}

	return nil
}

func traceLoop(ctx context.Context, trace chan dagstore.Trace, fail chan dagstore.ShardResult) {
	select {
	case <-ctx.Done():
		return
	case t := <-trace:
		log.Debugf("trace key %s, op: %v, shard info: %v", t.Key, t.Op, t.After)
	case t := <-fail:
		log.Warnf("result failed key: %s, error: %v", t.Key, t.Error)
	}
}

func gcLoop(ctx context.Context, ds *dagstore.DAGStore, gcInterval time.Duration) {
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = ds.GC(ctx)
			ticker.Reset(gcInterval)
		}
	}
}

type dagStoreImp struct {
	rootDir       string
	ds            *dagstore.DAGStore
	mountRegistry *mount.Registry
	e             *event.DagstoreEventStream
}

func newDagStoreImp(rootDir string, ds *dagstore.DAGStore, mountRegistry *mount.Registry, e *event.DagstoreEventStream) *dagStoreImp {
	return &dagStoreImp{
		rootDir:       rootDir,
		ds:            ds,
		mountRegistry: mountRegistry,
		e:             e,
	}
}

func (imp *dagStoreImp) RegisterShard(ctx context.Context, key shard.Key, opts dagstore.RegisterOpts) error {
	remoteMount, err := mount.NewRemoteMount(ctx, key.String(), imp.e)
	if err != nil {
		return err
	}
	if err := imp.mountRegistry.Register(key.String(), remoteMount); err != nil {
		return err
	}
	result := make(chan dagstore.ShardResult)
	if err := imp.ds.RegisterShard(ctx, key, remoteMount, result, opts); err != nil {
		return err
	}

	_, err = waitResult(ctx, result)

	return err
}

func (imp *dagStoreImp) GetShard(ctx context.Context, key shard.Key, opts dagstore.AcquireOpts) ([]byte, error) {
	result := make(chan dagstore.ShardResult)
	if err := imp.ds.AcquireShard(ctx, key, result, opts); err != nil {
		return nil, err
	}

	res, err := waitResult(ctx, result)
	if err != nil {
		return nil, err
	}

	bs, err := res.Accessor.Blockstore()
	if err != nil {
		return nil, err
	}
	c, err := cid.Parse(key.String())
	if err != nil {
		return nil, err
	}
	blk, err := bs.Get(ctx, c)
	if err != nil {
		return nil, err
	}

	return blk.RawData(), err
}

func (imp *dagStoreImp) RecoverShard(ctx context.Context, key shard.Key, opts dagstore.RecoverOpts) error {
	result := make(chan dagstore.ShardResult)
	if err := imp.ds.RecoverShard(ctx, key, result, opts); err != nil {
		return err
	}

	_, err := waitResult(ctx, result)

	return err
}

func (imp *dagStoreImp) GetShardInfo(key shard.Key) (dagstore.ShardInfo, error) {
	return imp.ds.GetShardInfo(key)
}

func (imp *dagStoreImp) AllShardsInfo() api.ShardsInfo {
	return api.ToShardsInfo(imp.ds.AllShardsInfo())
}

func (imp *dagStoreImp) ShardsContainingMultihash(ctx context.Context, h multihash.Multihash) ([]shard.Key, error) {
	return imp.ds.ShardsContainingMultihash(ctx, h)
}

func (imp *dagStoreImp) ListenDagstoreEvent(ctx context.Context, policy *event.DagstoreRegisterPolicy) (<-chan *gtypes.RequestEvent, error) {
	return imp.e.ListenDagstoreEvent(ctx, policy)
}

func (imp *dagStoreImp) ResponseDagstoreEvent(ctx context.Context, resp *gtypes.ResponseEvent) error {
	return imp.e.ResponseEvent(ctx, resp)
}

var _ api.IDagStore = (*dagStoreImp)(nil)

func waitResult(ctx context.Context, result chan dagstore.ShardResult) (*dagstore.ShardResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-result:
		return &res, res.Error
	}
}
