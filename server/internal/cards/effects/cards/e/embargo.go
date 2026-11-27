package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embargo — Enchantment {3}{U} (Mercadian Masques, #5619):
//
//	"Nonland permanents don't untap during their controllers' untap steps.
//	 At the beginning of your upkeep, you lose 2 life."
//
// Back to Basics' untap restriction over every nonland permanent, and an
// upkeep life loss for Embargo's controller only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "edadd0bf-15c8-4e9c-810e-2edfd77a9d01",
		Name:                  "Embargo",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.Nonland())},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Embargo — you lose 2 life", func(g *game.Game, item *game.StackItem) error {
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -2)
			}),
		},
	})
}
