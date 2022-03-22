package api

import "github.com/filecoin-project/dagstore"

type ShardsInfo map[string]dagstore.ShardInfo

func ToShardsInfo(all dagstore.AllShardsInfo) ShardsInfo {
	shardsInfo := make(map[string]dagstore.ShardInfo, len(all))
	for key, info := range all {
		shardsInfo[key.String()] = info
	}

	return shardsInfo
}
