package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spinehorn Minotaur — Creature — Minotaur Berserker {2}{R}, 2/3 (Modern Horizons, #5519):
//
//	"As long as you've drawn two or more cards this turn, this creature has
//	 double strike."
//
// The shared drawn-two keyword static, with double strike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db46b8f4-22bb-4042-abe8-dd83b2426142",
		Name:         "Spinehorn Minotaur",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.S82BroB07SelfKeywordWhileDrawnTwo("double strike")},
	})
}
