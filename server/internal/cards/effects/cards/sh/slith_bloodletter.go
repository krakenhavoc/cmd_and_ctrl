package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slith Bloodletter — Creature — Slith {B}{B}, 1/1 (Mirrodin, #5606):
//
//	"Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it.
//	 {1}{B}: Regenerate this creature."
//
// Hexgold Slith's trigger (mrdB5SlithGrows): the counter goes on the
// creature only while it is still the object that dealt the damage
// (CR 400.7).
//
// The regenerate row is the shared one (regenerate_this.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbad6038-93ed-4a28-b1f0-ada54608758b",
		Name:         "Slith Bloodletter",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.MrdB5SlithGrows("Slith Bloodletter")},
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{B}")},
	})
}
