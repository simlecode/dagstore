package log

import logging "github.com/ipfs/go-log/v2"

func init() {
	logging.SetAllLoggers(logging.LevelDebug)
}
