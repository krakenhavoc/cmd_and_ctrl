package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borderland Marauder — Creature — Human Warrior {1}{R}, 1/2 (Magic 2015,
// #5556):
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// Charging Bandits' attack trigger, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11fd2650-2150-4e62-8b64-e820d734a231",
		Name:         "Borderland Marauder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Borderland Marauder — it gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Borderland Marauder — +2/+0")),
		},
	})
}
