package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Money — Sorcery {5}{B}{B} (EDHREC rank 1652):
//
//	"Destroy all creatures. For each nontoken creature destroyed
//	 this way, you create a tapped Treasure token."
//
// A seven-mana Damnation that pays you a Treasure per real creature
// it killed, so the turn after the wipe is the biggest of the game.
//
// The batched sweep. DestroyAllMatching destroys every creature as
// one simultaneous event (CR 700.4), so a "whenever another creature
// dies" watcher caught in the wipe sees every death. The sweep also
// leaves indestructible creatures out of `swept` (#446 / #470), so a
// survivor pays nothing.
//
// The Then clause pays for `destroyed` minus the tokens. `swept` is
// the pre-move copies of exactly the creatures that were destroyed
// this way (#815), so the subtraction is the whole of the "nontoken"
// clause — a creature the CR 614 window saved or sent somewhere other
// than a graveyard is in neither number.
//
// A commander in the wipe WAS destroyed and is counted, since CR 903.9
// replaces the zone change and not the destruction (Fumigate's note).
// Its owner is asked first, so the Treasures are made when they
// answer rather than on this line. This clause used to walk `swept`
// looking for cards still on the battlefield, because it ran while
// that prompt was still open and could not tell a commander waiting
// on CR 903.9 from a creature that never left.
func init() {
	Register(Spec{
		OracleID:     "75f5d372-4ff9-430c-8302-72472439e0d2",
		Name:         "Blood Money",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: Creature(),
				Then: func(ctx *Context, swept []game.Card, destroyed int) error {
					paid := destroyed
					for _, c := range swept {
						if IsToken(c) {
							paid--
						}
					}
					return b13CreateTappedTreasures(ctx, ctx.Controller(), paid)
				},
			}.Apply(ctx)
		},
	})
}
