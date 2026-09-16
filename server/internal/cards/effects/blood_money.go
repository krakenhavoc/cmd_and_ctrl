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
// The Then clause pays for `destroyed` minus the tokens. It cannot
// walk `swept` and count what left the battlefield: a commander in
// the wipe was destroyed, but CR 903.9 queues its owner's
// command-zone prompt and the card stays on the battlefield until
// they answer, which is after this clause runs. `destroyed` already
// counts that commander, since CR 903.9 replaces the zone change and
// not the destruction (Fumigate's note). A token can never be a
// commander, so a token that is off the battlefield is exactly a
// token destroyed this way, and it is the part to take back out.
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
						if IsToken(c) && !ctx.Game.Battlefield.Contains(c.InstanceID) {
							paid--
						}
					}
					return b13CreateTappedTreasures(ctx, ctx.Controller(), paid)
				},
			}.Apply(ctx)
		},
	})
}
