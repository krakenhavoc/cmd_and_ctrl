package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goliath, Mass Manipulator — {1}{G} Legendary Creature — Human Scientist
// Hero 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Power-up — {4}{G}: Put two +1/+1 counters on Goliath. Then draw a card
//	 for each creature you control with power 4 or greater. (Activate each
//	 power-up ability only once. Reduce the cost by his mana cost if he
//	 entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The counters go on first, so
// Goliath counts himself once he reaches power 4; the draw counts the
// creatures as the ability resolves, with the layers recomputed (CR 608.2).
//
// No simplification.
func init() {
	ab := effects.PowerUp(
		"{4}{G}: Put two +1/+1 counters on Goliath. Then draw a card for each creature you control with power 4 or greater.",
		effects.ManaCost("{4}{G}"),
		effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}, effects.Instruction(func(ctx *effects.Context) error {
			return effects.DrawCards{N: effects.Msc13CreaturesWithPowerAtLeast(ctx.Game, ctx.Controller(), 4)}.Apply(ctx)
		})))
	ab.Purpose.Answers = game.AnswerPump
	effects.Register(effects.Spec{
		OracleID:        "d06d7065-27e5-4119-a226-6f6d990d72c7",
		Name:            "Goliath, Mass Manipulator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []effects.ActivatedAbility{ab},
	})
}
