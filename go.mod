module github.com/filecoin-project/dagstore

go 1.16

require (
	github.com/filecoin-project/go-jsonrpc v0.1.5
	github.com/filecoin-project/venus v1.2.2
	github.com/filecoin-project/venus-auth v1.3.2
	github.com/gorilla/mux v1.8.0
	github.com/ipfs-force-community/venus-gateway v1.2.0
	github.com/ipfs/go-block-format v0.0.3
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/go-datastore v0.5.1
	github.com/ipfs/go-ds-badger2 v0.1.2
	github.com/ipfs/go-ds-leveldb v0.5.0
	github.com/ipfs/go-ipfs-blocksutil v0.0.1
	github.com/ipfs/go-log/v2 v2.4.0
	github.com/ipld/go-car/v2 v2.1.1
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.4.1
	github.com/multiformats/go-multicodec v0.3.1-0.20210902112759-1539a079fd61
	github.com/multiformats/go-multihash v0.1.0
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20210713220151-be142a5ae1a8
	golang.org/x/exp v0.0.0-20210715201039-d37aa40e8013
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/filecoin-project/go-jsonrpc => github.com/ipfs-force-community/go-jsonrpc v0.1.4-0.20210731021807-68e5207079bc
