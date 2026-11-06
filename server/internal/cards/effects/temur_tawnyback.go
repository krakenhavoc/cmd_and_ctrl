package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temur Tawnyback — Creature — Beast {2/G}{2/U}{2/R}, 4/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, draw a card, then discard a card."
//
// Each monocolored hybrid symbol is paid with one mana of its colour or
// two mana of any type (CR 107.4e), and counts 2 toward its mana value
// of 6 (CR 202.3f). The loot is
// Overwhelmed Archivist's: the discard is chosen after the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "889d26d6-9e6c-4cc0-8e8e-dd932f705a20",
		Name:         "Temur Tawnyback",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Temur Tawnyback — draw a card, then discard a card", b35LootOne),
		},
	})
}
