package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Preening Champion — Creature — Bird Knight {2}{U}, 2/2 (March of the
// Machine):
//
//	"Flying
//	 When this creature enters, create a 1/1 blue and red Elemental
//	 creature token."
//
// An ordinary enters trigger that makes the shared 1/1 blue and red
// Elemental token row (Prismari Pianist's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e1c458ca-10f6-4da8-a93d-3adfb86fe97d",
		Name:            "Preening Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Preening Champion — create a 1/1 blue and red Elemental token",
				Do(CreateToken{Template: TokenCard("1/1 blue and red Elemental"), N: 1})),
		},
	})
}
