package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twinblade Assassins — Creature — Elf Assassin {3}{B}{G}, 5/4 (Core Set 2021, #5511):
//
//	"At the beginning of your end step, if a creature died this turn, draw
//	 a card."
//
// Warlock Class's end-step trigger (AtYourEndStepIfAnyCreatureDied): an
// intervening "if" (CR 603.4) over the table-wide tally, any controller's
// creature, tokens included, checked as the step begins and again as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66093486-c9c3-42c5-97d1-ec0fe54f65b0",
		Name:         "Twinblade Assassins",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfAnyCreatureDied("Twinblade Assassins — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
