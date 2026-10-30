package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venomized Cat — Creature — Symbiote Cat Villain {2}{B}, 2/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"Deathtouch
//	 When this creature enters, mill two cards."
//
// Black Widow, Daring Operative's mill-on-enter body with two cards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eb54d807-624b-471e-8c44-07322003903e",
		Name:            "Venomized Cat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Venomized Cat — mill two cards", Do(MillCards{N: 2})),
		},
	})
}
