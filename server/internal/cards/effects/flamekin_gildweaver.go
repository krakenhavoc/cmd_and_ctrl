package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flamekin Gildweaver — Creature — Elemental Sorcerer {3}{R}, 4/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Trample
//	 When this creature enters, create a Treasure token. (It's an artifact
//	 with "{T}, Sacrifice this token: Add one mana of any color.")"
//
// Professional Wrestler's entry trigger on a trampler: the shared
// TreasureToken template.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bc8427fb-e5ee-4b44-9a4f-6d9f01b4b896",
		Name:            "Flamekin Gildweaver",
		Purpose:         game.Purpose{Tokens: 1},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Flamekin Gildweaver — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
