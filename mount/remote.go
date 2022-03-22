package mount

import (
	"bytes"
	"context"
	"net/url"
)

type Resourcer interface {
	GetResource(ctx context.Context, resourceID string) ([]byte, error)
	HasResource(ctx context.Context, resourceID string) (bool, error)
}

var _ Mount = (*RemoteMount)(nil)

type RemoteMount struct {
	resourceID string
	resourcer  Resourcer
	size       int64
	exist      bool
}

func NewRemoteMount(ctx context.Context, resourceID string, resourcer Resourcer) (*RemoteMount, error) {
	var err error
	m := &RemoteMount{
		resourceID: resourceID,
		resourcer:  resourcer,
	}
	m.exist, err = resourcer.HasResource(ctx, resourceID)

	return m, err
}

func (r *RemoteMount) Close() error {
	return nil
}

func (r *RemoteMount) Fetch(ctx context.Context) (Reader, error) {
	data, err := r.resourcer.GetResource(ctx, r.resourceID)
	if err != nil {
		return nil, err
	}
	r.size = int64(len(data))
	r.exist = true

	return &readerCloser{bytes.NewReader(data)}, nil
}

func (r *RemoteMount) Info() Info {
	return Info{
		Kind:             KindRemote,
		AccessSequential: true,
	}
}

func (r *RemoteMount) Stat(ctx context.Context) (Stat, error) {
	return Stat{
		Exists: r.exist,
		Size:   r.size,
		Ready:  false,
	}, nil
}

func (r *RemoteMount) Serialize() *url.URL {
	u := &url.URL{}
	u.Scheme = r.resourceID
	return u
}

func (r *RemoteMount) Deserialize(u *url.URL) error {
	r.resourceID = u.Scheme
	return nil
}

var _ Reader = (*readerCloser)(nil)

type readerCloser struct {
	*bytes.Reader
}

func (r *readerCloser) Close() error {
	return nil
}
