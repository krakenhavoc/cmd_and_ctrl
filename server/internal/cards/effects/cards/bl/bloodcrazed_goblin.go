package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodcrazed Goblin — Creature — Goblin Berserker {R}, 2/2 (Magic 2011,
// #5579):
//
//	"This creature can't attack unless an opponent has been dealt damage this
//	 turn."
//
// A layer-6 restriction on the Goblin's own declaration (CR 508.1c), read
// live when attackers are declared from this turn's damage events: damage
// that was really dealt to a player other than the Goblin's controller, so
// prevented damage and life loss do not count. A creature the Goblin is not
// the controller's of is judged by whoever controls it then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a10e5b31-c9c1-45ca-85cc-4a50ed8c4946",
		Name:         "Bloodcrazed Goblin",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.S82OldB476CantAttackUnlessAnOpponentWasDealtDamage()},
	})
}
