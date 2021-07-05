package dagstore

type ShardState byte

const (
	// ShardStateNew indicates that a shard has just been registered and is
	// about to be processed for activation.
	ShardStateNew ShardState = iota

	// ShardStateInitializing indicates that the shard is being initialized
	// by being fetched from the mount and being indexed.
	ShardStateInitializing

	// ShardStateAvailable indicates that the shard has been initialized and is
	// active for serving queries. There are no active shard readers.
	ShardStateAvailable

	// ShardStateServing indicates the shard has active readers and thus is
	// currently actively serving requests.
	ShardStateServing

	// ShardStateErrored indicates that an unexpected error was encountered
	// during a shard operation, and therefore the shard needs to be recovered.
	ShardStateErrored = 0xf0

	// ShardStateUnknown indicates that it's not possible to determine the state
	// of the shard. This state is currently unused, but it's reserved.
	ShardStateUnknown = 0xff
)