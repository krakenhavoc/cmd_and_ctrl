package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hazoret's Favor — {2}{R} Enchantment (Amonkhet, S82 old batch 246):
//
//	"At the beginning of combat on your turn, you may have target creature you
//	 control get +2/+0 and gain haste until end of turn. If you do, sacrifice it
//	 at the beginning of the next end step."
//
// One trigger: the target is chosen when it goes on the stack, the "may" is
// asked as it resolves, and only a yes pumps the creature and schedules the
// delayed sacrifice of that very object (CR 603.7), which does nothing if the
// creature has already left.
//
// No simplification.
func init() {
	const label = "Hazoret's Favor — you may have target creature you control get +2/+0 and gain haste until end of turn, then sacrifice it at the beginning of the next end step"
	effects.Register(effects.Spec{
		OracleID:     "2071b712-2539-4797-8359-f9adb31889ae",
		Name:         "Hazoret's Favor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true), label,
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.BoostUntilEOT{Target: id, Power: 2, Label: "Hazoret's Favor — +2/+0"}).Apply(ctx); err != nil {
							return err
						}
						if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Hazoret's Favor — haste"}).Apply(ctx); err != nil {
							return err
						}
						return effects.ScheduleDelayedTrigger{
							At:    game.StepEnd,
							Label: "Hazoret's Favor — sacrifice it",
							Cards: []uuid.UUID{id},
							Body:  effects.SacrificeListedCardsBody,
						}.Apply(ctx)
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				"Hazoret's Favor — have the target creature get +2/+0 and haste, and sacrifice it at the next end step?"),
		},
	})
}
