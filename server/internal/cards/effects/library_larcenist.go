package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Library Larcenist — Creature — Merfolk Rogue {2}{U}, 1/2 (Core Set 2021, #5511):
//
//	"Whenever this creature attacks, draw a card."
//
// Svyelun's attack trigger: a Do of one draw, so the row declares its
// footprint.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb0e6f63-2861-4392-ac0e-64380fbbd8ae",
		Name:         "Library Larcenist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Library Larcenist — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
