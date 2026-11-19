package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tithebearer Giant — Creature — Giant Warrior {5}{B}, 4/5 (Ravnica: Clue
// Edition, S82):
//
//	"When this creature enters, you draw a card and you lose 1 life."
//
// Discreet Retreat's draw-then-lose body on an enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "475aaf7c-b0f8-40f3-ad91-3e7d414de933",
		Name:         "Tithebearer Giant",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tithebearer Giant — you draw a card and you lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
	})
}
