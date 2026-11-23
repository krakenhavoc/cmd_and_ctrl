package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steadfast Cathar — Creature — Human Soldier {1}{W}, 2/1 (Eldritch Moon,
// #5541):
//
//	"Whenever this creature attacks, it gets +0/+2 until end of turn."
//
// Charging Paladin's attack trigger with +0/+2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00895741-f94f-4d80-8ca4-c436bf3c8d56",
		Name:         "Steadfast Cathar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Steadfast Cathar — it gets +0/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(0, 2, "Steadfast Cathar — +0/+2")),
		},
	})
}
