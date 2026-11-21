package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slith Firewalker — Creature — Slith {R}{R}, 1/1 (Mirrodin, #5606):
//
//	"Haste
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it."
//
// Printed haste, and Hexgold Slith's trigger (mrdB5SlithGrows): the
// counter goes on the creature only while it is still the object that
// dealt the damage (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea1af948-686d-4b40-ad4e-cb48a0258269",
		Name:            "Slith Firewalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{effects.MrdB5SlithGrows("Slith Firewalker")},
	})
}
