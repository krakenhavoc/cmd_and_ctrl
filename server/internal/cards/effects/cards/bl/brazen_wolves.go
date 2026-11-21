package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brazen Wolves — Creature — Wolf {2}{R}, 2/3 (Eldritch Moon, #5541):
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// Charging Bandits' attack trigger, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b5a3d7b-6b2c-4e91-8218-2b957277855c",
		Name:         "Brazen Wolves",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Brazen Wolves — it gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Brazen Wolves — +2/+0")),
		},
	})
}
