package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slith Predator — Creature — Slith {G}{G}, 1/1 (Mirrodin, #5606):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it."
//
// Printed trample, and Hexgold Slith's trigger (mrdB5SlithGrows): the
// counter goes on the creature only while it is still the object that
// dealt the damage (CR 400.7).
// Damage trampled over to the player counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fd9c984e-daa2-414c-bb90-ecc41aaa896d",
		Name:            "Slith Predator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{effects.MrdB5SlithGrows("Slith Predator")},
	})
}
