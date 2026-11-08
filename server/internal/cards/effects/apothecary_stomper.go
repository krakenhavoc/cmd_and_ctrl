package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Apothecary Stomper — Creature — Elephant {4}{G}{G}, 4/4 (Foundations,
// #4465):
//
//	"Vigilance
//	 When this creature enters, choose one —
//	 • Put two +1/+1 counters on target creature you control.
//	 • You gain 4 life."
//
// Charming Prince's modal enters trigger: the mode is chosen as the
// trigger goes on the stack, and the counters' target is chosen with it
// and re-checked at resolution (CR 608.2b). The Stomper itself is a
// legal target.
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Apothecary Stomper — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Put two +1/+1 counters on target creature you control.",
			TargetCreature("target creature you control", YouControl()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}),
		ModeDoing("You gain 4 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:        "6a381121-9134-489c-8c7e-44be1936b0e2",
		Name:            "Apothecary Stomper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
