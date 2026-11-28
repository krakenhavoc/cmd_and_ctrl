package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thelon's Curse — Enchantment {G}{G} (Fallen Empires, S82 OLD b856, #5637):
//
//	"Blue creatures don't untap during their controllers' untap steps.
//	 At the beginning of each player's upkeep, that player may choose any number of
//	 tapped blue creatures they control and pay {U} for each creature chosen this
//	 way. If the player does, untap those creatures."
//
// Dream Tides' two lines at {U} a creature over blue creatures: Root Cage's
// untap-step restriction, whoever controls them, and an every-upkeep trigger
// whose player picks the creatures and is then offered the total payment as one
// prompt. The colour is read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "f6880f71-a05a-4752-ba33-ffbc450606bf",
		Name:                  "Thelon's Curse",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.And(effects.Creature(), effects.OfColor("U")))},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Thelon's Curse — that player may choose any number of tapped blue creatures they control and pay {U} for each; untap those creatures", effects.S82OldB856ThelonsCurseUpkeep),
		},
	})
}
