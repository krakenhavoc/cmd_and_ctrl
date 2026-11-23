package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abzan Charm — Instant {W}{B}{G} (EDHREC rank 4433):
//
//	"Choose one —
//	 • Exile target creature with power 3 or greater.
//	 • You draw two cards and you lose 2 life.
//	 • Distribute two +1/+1 counters among one or two target
//	   creatures."
//
// The Khans charm that has stayed relevant: exile (not destroy) for
// the big thing, or a Sign in Blood when there is no big thing.
// Roadmap batch 42 (#449), "no new machinery".
//
// Mode 0 EXILES, which is the mode's whole point — indestructible,
// regeneration and death triggers all miss. "Power 3 or greater" is
// the layered power, checked at announce (CR 601.2c) and again at
// resolution (CR 608.2b), so shrinking the creature in response is a
// real answer.
//
// Mode 1 draws and then loses, in that order. It matters at 2 life or
// less: you draw the cards first and the loss can kill you, which is
// the printed card.
//
// Mode 2 distributes (#1563, CR 601.2d): one or two target
// creatures, and the caster announces the split with the targets —
// both counters on one, or one each. A target that leaves in response
// takes nothing and its counter is lost, not moved to the other.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4137a22c-f793-4e27-a9b2-72740a6e2121",
		Name:         "Abzan Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Exile target creature with power 3 or greater.",
				effects.TargetCreature("target creature with power 3 or greater", effects.PowerGE(3))),
			effects.Mode("You draw two cards and you lose 2 life."),
			effects.Mode("Distribute two +1/+1 counters among one or two target creatures.",
				effects.TargetCreature("one or two target creatures").WithCount(1, 2).Dividing(effects.Divide(2))),
		),
		// Each step is guarded by its mode; the life loss waits for a prompt the
		// draw asks (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				item := ctx.Item
				if !ctx.HasMode(0) || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return effects.ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
			}),
			effects.YouDrawIfMode(1, 2),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				return effects.GainLife{Player: ctx.Controller(), Amount: -2}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(2) {
					return nil
				}
				return effects.PutDividedCounters(ctx, game.CounterPlusOne)
			}),
		),
	})
}
