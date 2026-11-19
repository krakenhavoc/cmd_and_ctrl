package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Audacious Thief — Creature — Human Rogue {2}{B}, 2/2 (Foundations
// Jumpstart, #5929):
//
//	"Whenever this creature attacks, you draw a card and you lose 1 life."
//
// One attack trigger whose body is Barbed Servitor's draw-then-lose-one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5905c55a-f1d1-4021-a9af-f85b02ae21f3",
		Name:         "Audacious Thief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Audacious Thief — you draw a card and you lose 1 life", b36DrawAndLoseOne),
		},
	})
}
