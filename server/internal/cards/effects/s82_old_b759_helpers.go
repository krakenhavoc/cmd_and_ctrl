package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB759PowerNoGreaterThanCounters is "with power less than or equal to
// the number of <kind> counters on this" (Legacy's Allure): the candidate's
// power against the source's counters of `kind`, last-known once the source
// has left (CR 608.2h).
func S82OldB759PowerNoGreaterThanCounters(kind string) game.SourceComparison {
	return game.SourceComparison{Stat: game.SourcePowerVsCounters, Cmp: game.CmpLessEq, Counter: kind}
}

// S82OldB759PutABountyCounter is "put a bounty counter on target creature" as
// an activated ability's whole effect (Bounty Board, Bounty Hunter): the first
// still-legal target gets the counter, and nothing happens if it has gone
// (CR 608.2b).
func S82OldB759PutABountyCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: "bounty", N: 1}.Apply(ctx)
}
