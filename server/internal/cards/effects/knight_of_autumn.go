package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of Autumn — Creature — Dryad Knight {1}{G}{W}, 2/1:
//
//	"When this creature enters, choose one —
//	 • Put two +1/+1 counters on this creature.
//	 • Destroy target artifact or enchantment.
//	 • You gain 4 life."
//
// One modal enters-the-battlefield trigger. The mode is chosen as the
// trigger is put on the stack (CR 700.2a), and only the destroy mode has a
// target (CR 700.2c). If the Knight has already left the battlefield when
// the first mode resolves, there is nothing to put counters on.
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Knight of Autumn — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Put two +1/+1 counters on this creature.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			if !ctx.Game.Battlefield.Contains(item.SourceCardID) {
				return nil
			}
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
		}),
		ModeDoing("Destroy target artifact or enchantment.",
			TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			titanOfIndustryDestroy), // the same "destroy the mode's target" body
		ModeDoing("You gain 4 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     "c4212ba4-8c43-40ca-aa36-b3b535659fa8",
		Name:         "Knight of Autumn",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
