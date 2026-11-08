package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Trampled Garden — Enchantment — Case {2}{G} (Murders at
// Karlov Manor, #4469):
//
//	"When this Case enters, distribute two +1/+1 counters among one or
//	 two target creatures you control.
//	 To solve — Creatures you control have total power 8 or greater.
//	 (If unsolved, solve at the beginning of your end step.)
//	 Solved — Whenever you attack, put a +1/+1 counter on target
//	 attacking creature. It gains trample until end of turn."
//
// The enters trigger is Jade Seedstones' divided counters (the division
// is made as the trigger is put on the stack, CR 601.2d). The solve
// condition is Mosswort Bridge's total power, current power with
// counters, read at the end step and again on resolution (CR 603.4).
// The solved line is Araña's "whenever you attack": one trigger per
// declaration (OncePerBatch), gated on the Case being solved, and the
// trample goes to the same target if it is still legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "575a0dc2-206a-406a-90b3-72a92f2ea87a",
		Name:         "Case of the Trampled Garden",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Case of the Trampled Garden — distribute two +1/+1 counters",
					func(g *game.Game, item *game.StackItem) error {
						return PutDividedCounters(NewContext(g, item), game.CounterPlusOne)
					}),
				TargetCreature("one or two target creatures you control", YouControl()).
					WithCount(1, 2).Dividing(Divide(2))),
			ToSolve("Case of the Trampled Garden — to solve: creatures you control have total power 8 or greater",
				func(g *game.Game, controller, _ uuid.UUID) bool {
					return creaturesYouControlTotalPowerAtLeast(8)(g, controller)
				}),
			WhenSolved(OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Case of the Trampled Garden — solved: put a +1/+1 counter on target attacking creature; it gains trample",
					Do(Instruction(counterOnFirstCreatureTarget), Instruction(trampledGardenTrample))),
				TargetCreature("target attacking creature", AttackingCreature())))),
		},
	})
}

// trampledGardenTrample is "It gains trample until end of turn": the
// creature the counter went on, while it is still a legal target.
func trampledGardenTrample(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   t.ID,
		Keywords: []string{"trample"},
		Label:    "Case of the Trampled Garden — trample",
	}.Apply(ctx)
}
