package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spinehorn Minotaur — Creature — Minotaur Berserker {2}{R}, 2/3 (Modern Horizons, #5519):
//
//	"As long as you've drawn two or more cards this turn, this creature has
//	 double strike."
//
// The shared drawn-two keyword static, with double strike.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db46b8f4-22bb-4042-abe8-dd83b2426142",
		Name:         "Spinehorn Minotaur",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{s82BroB07SelfKeywordWhileDrawnTwo("double strike")},
	})
}
