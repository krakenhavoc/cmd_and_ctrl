package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Gargantua — Creature — Phyrexian Horror {4}{B}{B}, 4/4:
//
//	"When this creature enters, you draw two cards and you lose 2 life."
//
// One ability: the draw, then the life loss (Mephitic Draught's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3925dfd-62b0-44ef-8545-7bee92cad5a2",
		Name:         "Phyrexian Gargantua",
		Purpose:      game.Purpose{Draws: 2},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Phyrexian Gargantua — you draw two cards and you lose 2 life",
				Do(DrawCards{N: 2}, youLoseLife(2))),
		},
	})
}
