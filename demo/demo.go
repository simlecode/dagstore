package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	event "github.com/filecoin-project/dagstore/event"
	"github.com/filecoin-project/go-jsonrpc"
	gtypes "github.com/filecoin-project/venus/venus-shared/types/gateway"
)

const piecePath = "/tmp/mnt/piece"

type dagstoreClient struct {
	ListenDagstoreEvent   func(ctx context.Context, policy event.DagstoreRegisterPolicy) (<-chan *gtypes.RequestEvent, error)
	ResponseDagstoreEvent func(ctx context.Context, resp *gtypes.ResponseEvent) error
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()
	cli := &dagstoreClient{}
	closer, err := jsonrpc.NewMergeClient(ctx, "ws://127.0.0.1:7898/rpc/v0", "DagStore", []interface{}{cli}, http.Header{})
	checkErr(err)
	defer closer()

	for {
		reqEvent, err := cli.ListenDagstoreEvent(ctx, event.DagstoreRegisterPolicy{})
		if err != nil {
			fmt.Println(err)
			return
		}
		checkErr(err)

		for req := range reqEvent {
			switch req.Method {
			case "GetResource":
				fmt.Println("get resource")
				r := event.ResourceRequest{}
				err := json.Unmarshal(req.Payload, &r)
				checkErr(err)

				_, err = os.Stat(filepath.Join(piecePath, r.ResourceID))
				checkErr(err)
				data, err := ioutil.ReadFile(filepath.Join(piecePath, r.ResourceID))
				checkErr(err)
				resp := gtypes.ResponseEvent{
					ID:      req.ID,
					Payload: data,
				}
				start := time.Now()
				fmt.Printf("start: %v, length: %v\n", start.String(), len(data))
				fmt.Println(cli.ResponseDagstoreEvent(ctx, &resp))
				fmt.Printf("end: %v\n", time.Since(start).Seconds())
			case "HasResource":
				r := event.ResourceRequest{}
				err := json.Unmarshal(req.Payload, &r)
				checkErr(err)
				_, err = os.Stat(filepath.Join(piecePath, r.ResourceID))
				checkErr(err)
				d, err := json.Marshal(true)
				checkErr(err)
				fmt.Println("has: ", d)
				resp := gtypes.ResponseEvent{
					ID:      req.ID,
					Payload: d,
				}
				checkErr(cli.ResponseDagstoreEvent(ctx, &resp))
			}
		}
	}
}
