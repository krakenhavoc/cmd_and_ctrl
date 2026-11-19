package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venomous Hierophant — Creature — Gorgon Cleric {3}{B}, 3/3 (Ravnica: Clue
// Edition, S82):
//
//	"Deathtouch
//	 When this creature enters, mill three cards."
//
// Colossal Grave-Reaver's mill body on an enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e235d1ca-3099-46c1-8f33-798286292d71",
		Name:            "Venomous Hierophant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Venomous Hierophant — mill three cards", Do(MillCards{N: 3})),
		},
	})
}
