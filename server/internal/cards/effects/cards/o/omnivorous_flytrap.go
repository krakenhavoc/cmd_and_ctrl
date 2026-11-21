package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Omnivorous Flytrap — Creature — Plant {2}{G}, 2/4 (Duskmourn: House of
// Horror, #4466):
//
//	"Delirium — Whenever this creature enters or attacks, if there are
//	 four or more card types among cards in your graveyard, distribute two
//	 +1/+1 counters among one or two target creatures. Then if there are
//	 six or more card types among cards in your graveyard, double the
//	 number of +1/+1 counters on those creatures."
//
// One ability on two event kinds (WhenThisEntersOrAttacks's shape) with an
// intervening "if" (CR 603.4): the four-type count is read as the event is
// seen (AppliesTo) and again as the ability resolves (InterveningIf), and
// the ability does nothing if it no longer holds. The counters are Abzan
// Charm's division (PutDividedCounters, CR 601.2d: each target at least
// one, the announced split honoured, a departed target's counters lost).
// The second sentence is read at its own moment, after the counters have
// landed and any replacement-order prompt on them is settled: the
// six-type count is read then, and "those creatures" are the targets still
// legal, each putting as many +1/+1 counters on itself as it has (Primordial
// Hydra's doubling, so a counter doubler applies to it, as the ruling says).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0aa2ab46-c5ad-462b-b441-dbbc4b2c3532",
		Name:         "Omnivorous Flytrap",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIf(effects.B25CondGraveyardTypes{AtLeast: 4},
					effects.OnAny([]game.EventKind{game.EventETB, game.EventAttack},
						effects.AllOf(effects.Self, effects.B25YourGraveyardTypesAtLeast(4)),
						"Omnivorous Flytrap — distribute two +1/+1 counters among one or two target creatures; then with six or more card types, double them",
						effects.Do(
							effects.Instruction(func(ctx *effects.Context) error {
								return effects.PutDividedCounters(ctx, game.CounterPlusOne)
							}),
							effects.Instruction(func(ctx *effects.Context) error {
								if effects.B16CardTypesInGraveyard(ctx.Game, ctx.Controller()) < 6 {
									return nil
								}
								for _, t := range ctx.LegalTargets() {
									if t.Kind != game.TargetCard {
										continue
									}
									if err := effects.B23DoublePlusOneCountersOn(ctx, t.ID); err != nil {
										return err
									}
								}
								return nil
							}),
						))),
				effects.TargetCreature("one or two target creatures").WithCount(1, 2).Dividing(effects.Divide(2)),
			),
		},
	})
}
