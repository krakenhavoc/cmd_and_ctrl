package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	ab := PowerUp(
		"{4}{G}: Put two +1/+1 counters on Goliath. Then draw a card for each creature you control with power 4 or greater.",
		ManaCost("{4}{G}"),
		Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2}, Instruction(func(ctx *Context) error {
			return DrawCards{N: msc13CreaturesWithPowerAtLeast(ctx.Game, ctx.Controller(), 4)}.Apply(ctx)
		})))
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:        "d06d7065-27e5-4119-a226-6f6d990d72c7",
		Name:            "Goliath, Mass Manipulator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []ActivatedAbility{ab},
	})
}
