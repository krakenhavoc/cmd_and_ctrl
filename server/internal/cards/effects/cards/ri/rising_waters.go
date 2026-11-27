package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rising Waters — Enchantment {3}{U} (Nemesis, S82 OLD, #5618):
//
//	"Lands don't untap during their controllers' untap steps.
//	 At the beginning of each player's upkeep, that player untaps a land they
//	 control."
//
// Hokori, Dust Drinker's two lines: Back to Basics' untap-step restriction over
// every land, and an every-upkeep trigger whose player chooses the land they
// untap (CR 608.2d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "d4cbf380-1a05-48ea-9187-de931b00866d",
		Name:                  "Rising Waters",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.Land())},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Rising Waters — that player untaps a land they control", effects.S82OldB021UntapALand),
		},
	})
}
