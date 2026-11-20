package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sabertooth Mauler — Creature — Cat {3}{G}, 3/3 (Core Set 2021, #5511):
//
//	"At the beginning of your end step, if a creature died this turn, put a
//	 +1/+1 counter on this creature and untap it."
//
// Bulette's end-step counter (AtYourEndStepIfAnyCreatureDied over the
// table-wide tally) followed by an untap of the same object. A Mauler that
// left the battlefield gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da4f4094-b0ce-4b5b-8ac3-8c0a5159220f",
		Name:         "Sabertooth Mauler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfAnyCreatureDied("Sabertooth Mauler — put a +1/+1 counter on this creature and untap it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, Instruction(func(ctx *Context) error {
					return UntapTarget{Target: ctx.Source()}.Apply(ctx)
				}))),
		},
	})
}
