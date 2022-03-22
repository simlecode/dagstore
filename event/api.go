package dagstore_event

import (
	"context"

	types "github.com/filecoin-project/venus/venus-shared/types/gateway"
)

type IDagstoreEventAPI interface {
	ResponseDagstoreEvent(ctx context.Context, resp *types.ResponseEvent) error
	ListenDagstoreEvent(ctx context.Context, policy *DagstoreRegisterPolicy) (<-chan *types.RequestEvent, error)
}

type IDagstoreEvent interface {
	GetResource(ctx context.Context, resourceID string) ([]byte, error)
	HasResource(ctx context.Context, resourceID string) (bool, error)
}

type DagstoreEventAPI struct {
	dagstoreEvent *DagstoreEventStream
}

func (e *DagstoreEventAPI) ResponseDagstoreEvent(ctx context.Context, resp *types.ResponseEvent) error {
	return e.dagstoreEvent.ResponseEvent(ctx, resp)
}

func (e *DagstoreEventAPI) ListenDagstoreEvent(ctx context.Context, policy *DagstoreRegisterPolicy) (<-chan *types.RequestEvent, error) {
	return e.dagstoreEvent.ListenDagstoreEvent(ctx, policy)
}

var _ IDagstoreEventAPI = (*DagstoreEventAPI)(nil)
