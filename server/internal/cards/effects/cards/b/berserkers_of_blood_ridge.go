package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Berserkers of Blood Ridge — Creature — Human Berserker {4}{R}, 4/4 (Magic 2010, #5585):
//
//	"This creature attacks each combat if able."
//
// Bloodrock Cyclops' attack requirement (CR 508.1d): the engine refuses to
// finish declaring attackers while it could attack and doesn't.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e8a2d53-3145-421a-aad1-8f2b2f6eaec8",
		Name:         "Berserkers of Blood Ridge",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
