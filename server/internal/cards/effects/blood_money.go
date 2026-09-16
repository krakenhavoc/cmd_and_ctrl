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
// The Then clause cannot use the `destroyed` count, because that
// count includes tokens and the card pays only for nontoken
// creatures. It walks `swept` instead, and counts a card only when
// it is no longer on the battlefield, so a creature whose move failed
// is never paid for. A commander sent to the command zone (CR 903.9)
// was still destroyed and counts, as Fumigate notes. The hand-rolled
// loop this replaced made the same battlefield check.
func init() {
	Register(Spec{
		OracleID:     "75f5d372-4ff9-430c-8302-72472439e0d2",
		Name:         "Blood Money",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: Creature(),
				Then: func(ctx *Context, swept []game.Card, _ int) error {
					paid := 0
					for _, c := range swept {
						if !IsToken(c) && !ctx.Game.Battlefield.Contains(c.InstanceID) {
							paid++
						}
					}
					return b13CreateTappedTreasures(ctx, ctx.Controller(), paid)
				},
			}.Apply(ctx)
		},
	})
}
