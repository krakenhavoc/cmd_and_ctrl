package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hokori, Dust Drinker — Legendary Creature — Spirit {2}{W}{W}, 2/2 (Secret
// Lair Drop):
//
//	"Lands don't untap during their controllers' untap steps.
//	 At the beginning of each player's upkeep, that player untaps a land they
//	 control."
//
// The first sentence is Back to Basics' untap-step restriction over every
// land. The upkeep trigger is on every player's upkeep and the player whose
// upkeep it is chooses the land (CR 608.2d); a player with no tapped land
// does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "d2aaf485-8c74-4a27-b80b-f600c20d8b9c",
		Name:                  "Hokori, Dust Drinker",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.Land())},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Hokori, Dust Drinker — that player untaps a land they control", effects.S82OldB021UntapALand),
		},
	})
}
