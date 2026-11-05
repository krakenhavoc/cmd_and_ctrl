package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battle of Wolf 359 — Sorcery {3}{B}{B} (Star Trek Commander, #4974):
//
//	"All nonartifact creatures get -X/-X until end of turn, where X is
//	 the number of artifact creatures you control."
//
// X is counted once, as the spell resolves, over the artifact creatures
// its controller controls then (post-layer types, so a crewed Vehicle
// counts). The affected set is locked at the same moment (CR 611.2c): a
// nonartifact creature that enters later this turn is not shrunk, and one
// that becomes an artifact later keeps its -X/-X. Every player's
// nonartifact creatures are affected, the caster's included. With no
// artifact creature, X is 0 and nothing changes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bb203cf5-1545-43c7-8278-3720616c22f4",
		Name:         "Battle of Wolf 359",
		Completeness: CompletenessFull,
		// A -X/-X sweep that spares artifact creatures; X is counted, as
		// Nuclear Fallout's and Martyr of Ashes' are.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && c.IsArtifact() && c.IsCreature() {
					x++
				}
			}
			return BoostUntilEOT{
				Match:     And(Creature(), Not(Artifact())),
				Power:     -x,
				Toughness: -x,
				Label:     "Battle of Wolf 359 — -X/-X",
			}.Apply(ctx)
		},
	})
}
