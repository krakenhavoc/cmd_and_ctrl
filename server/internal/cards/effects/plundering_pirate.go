package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plundering Pirate — Creature — Orc Pirate {2}{R}, 3/2 (The Lost Caverns
// of Ixalan, #4470):
//
//	"When this creature enters, create a Treasure token. (It's an artifact
//	 with "{T}, Sacrifice this token: Add one mana of any color.")"
//
// Flamekin Gildweaver's entry trigger without trample: the shared
// TreasureToken template.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ab13412-1eaf-40c0-905d-8a3ca0c51be5",
		Name:         "Plundering Pirate",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Plundering Pirate — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
