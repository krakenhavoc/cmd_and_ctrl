package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sickle Dancer — Creature — Human Warrior {2}{B}, 3/2 (Battlebond, #5529):
//
//	"Whenever this creature attacks, if your team controls another
//	 Warrior, this creature gets +1/+1 until end of turn."
//
// "Your team" is "you" (CR 102.4). The intervening "if" is the shared
// Battlebond Warrior check (bbdWheneverAttacksWithAnotherWarrior); the pump
// is on the Dancer itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "836bfcad-014a-40cc-ab79-906fcab723a0",
		Name:         "Sickle Dancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.BbdWheneverAttacksWithAnotherWarrior("Sickle Dancer — gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Sickle Dancer — +1/+1 until end of turn")),
		},
	})
}
