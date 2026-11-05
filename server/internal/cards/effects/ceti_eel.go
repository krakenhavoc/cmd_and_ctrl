package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ceti Eel — Creature — Worm {3}{B}, 2/2 (Star Trek Commander, #4974):
//
//	"When this creature enters, mill two cards, then put a +1/+1 counter
//	 on this creature for each artifact and/or creature card in your
//	 graveyard."
//
// The mill happens first, so the two milled cards count; the count is
// read as the counters are put, over the whole graveyard (an artifact
// creature card counts once). An Eel that left the battlefield before
// resolution gets nothing (CR 400.7), but the mill still happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1914241-6b2b-4b3e-8626-aeb0b8cbc12d",
		Name:         "Ceti Eel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ceti Eel — mill two cards, then put a +1/+1 counter on it for each artifact and/or creature card in your graveyard",
				Do(MillCards{N: 2}, Instruction(func(ctx *Context) error {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Graveyard == nil {
						return nil
					}
					n := 0
					for _, c := range p.Graveyard.Cards {
						if c.IsArtifact() || c.IsCreature() {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}))),
		},
	})
}
