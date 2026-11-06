package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meticulous Artisan — Creature — Djinn Artificer {3}{R}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 When this creature enters, create a Treasure token. (It's an
//	 artifact with "{T}, Sacrifice this token: Add one mana of any
//	 color.")"
//
// Prowess is the keyword the engine turns into its trigger (CR 702.108a).
// The Treasure is the predefined token (CR 111.10a), Corsair Captain's
// entry trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2b2b3b16-389c-4581-b889-b8a2e76f83a2",
		Name:            "Meticulous Artisan",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1},
		PrintedKeywords: []string{game.KeywordProwess},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Meticulous Artisan — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			})),
		},
	})
}
