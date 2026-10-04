package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Traverse the Outlands — Sorcery {4}{G}:
//
//	"Search your library for up to X basic land cards, where X is the
//	 greatest power among creatures you control. Put those cards onto
//	 the battlefield tapped, then shuffle."
//
// X is read as the spell resolves (b42GreatestPowerControlledBy:
// layered power, counters included; zero with no creature). "Up to"
// is the search's own limit, and the lands enter tapped together
// before the shuffle (Explosive Vegetation's shape).
//
// X of zero is still a search and still a shuffle: the search matches
// nothing rather than being skipped, because SearchLibrary reads a
// zero limit as one and the printed shuffle is unconditional.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d49cf51-af1f-4c17-9cf4-b82bc7c2c72e",
		Name:         "Traverse the Outlands",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := b42GreatestPowerControlledBy(ctx.Game, item.Controller)
			match := IsBasicLand
			if x <= 0 {
				match = func(game.Card) bool { return false }
			}
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     match,
				Dest:          game.ZoneBattlefield,
				Limit:         x,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Traverse the Outlands — up to X basic lands, tapped",
			}.Apply(ctx)
		},
	})
}
