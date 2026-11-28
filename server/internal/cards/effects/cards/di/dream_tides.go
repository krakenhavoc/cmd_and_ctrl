package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dream Tides — Enchantment {2}{U}{U} (Visions, S82 OLD, #5631):
//
//	"Creatures don't untap during their controllers' untap steps.
//	 At the beginning of each player's upkeep, that player may choose any
//	 number of tapped nongreen creatures they control and pay {2} for each
//	 creature chosen this way. If the player does, untap those creatures."
//
// Rising Waters' untap-step restriction over creatures, and an every-upkeep
// trigger: the active player picks the creatures, then is offered the total
// {2}-per-creature payment as one prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "d9583aab-3eff-4d88-84d5-7219ab982636",
		Name:                  "Dream Tides",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.Creature())},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Dream Tides — that player may choose any number of tapped nongreen creatures they control and pay {2} for each; untap those creatures", effects.S82OldB788DreamTidesUpkeep),
		},
	})
}
