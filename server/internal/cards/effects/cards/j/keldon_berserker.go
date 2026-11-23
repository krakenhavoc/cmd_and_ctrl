package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Berserker — Creature — Human Soldier Berserker {3}{R}, 2/3
// (Prophecy, #5794):
//
//	"Whenever this creature attacks, if you control no untapped lands, it
//	 gets +3/+0 until end of turn."
//
// An intervening "if" (CR 603.4, no_untapped_lands.go): checked as the
// Berserker is declared as an attacker and again as the trigger
// resolves, so a land untapped in response leaves it 2/3. The pump is
// Soltari Trooper's thisGetsUntilEndOfTurn, ending in the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e0b9e67-3103-4299-934b-46d8fe96f553",
		Name:         "Keldon Berserker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacksIfNoUntappedLands("Keldon Berserker — it gets +3/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(3, 0, "Keldon Berserker — +3/+0")),
		},
	})
}
