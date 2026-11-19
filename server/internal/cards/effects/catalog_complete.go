package effects

import "sync/atomic"

// The catalog completeness guard (#5906, ADR 0224).
//
// After the split, card files register from the shards under cards/, and only
// the aggregator package internal/cards/effects/cards imports them all. A
// binary that imported this package without the aggregator would read a
// partial catalog without a word: Lookup would miss, and a snapshot restore
// that names a card's effect key would not find it ("a restore never drops a
// stack item"). The guard makes that loud instead: the aggregator's init calls
// MarkCatalogComplete once every shard has registered, and Lookup, All and Has
// panic if they run before it.
//
// catalogSharded stays false until the split itself (scripts/split-effects)
// sets it to true. Until then every card is in this package, importing it
// registers them all, and the guard compiles to nothing.
const catalogSharded = false

var catalogComplete atomic.Bool

// MarkCatalogComplete records that every card shard has registered. Only the
// aggregator, internal/cards/effects/cards, calls it, from its init.
func MarkCatalogComplete() { catalogComplete.Store(true) }

// requireCompleteCatalog panics when the catalog is read before every shard
// registered: the binary imports the core but not the aggregator.
func requireCompleteCatalog(op string) {
	if catalogSharded && !catalogComplete.Load() {
		panic("effects." + op + ": the card catalog was read before its shards registered. " +
			`Link them all with import _ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards" ` +
			"in the main package, or in an _test.go file of the package under test (ADR 0224, #5906)")
	}
}
