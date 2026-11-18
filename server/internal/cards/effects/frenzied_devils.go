package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frenzied Devils — Creature — Devil {4}{R}, 3/3 (Innistrad: Crimson
// Vow, #5495):
//
//	"Haste
//	 Whenever you cast a noncreature spell, this creature gets +2/+2
//	 until end of turn."
//
// Slickshot Show-Off's cast trigger at +2/+2; haste is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c912e9a-21f9-4058-b28c-43f319e35d09",
		Name:            "Frenzied Devils",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Frenzied Devils — gets +2/+2 until end of turn",
				thisGetsUntilEndOfTurn(2, 2, "Frenzied Devils — +2/+2")),
		},
	})
}
