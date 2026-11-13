package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lava Hounds — Creature — Dog {2}{R}{R}, 4/4 (Weatherlight, #5629):
//
//	"Haste
//	 When this creature enters, it deals 4 damage to you."
//
// Haste is the printed keyword (CR 702.10b). The enters trigger is
// damage from the Hounds to their controller, which loses that much life
// (CR 120.3a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86c0dc85-d146-4dfa-819e-f78835bfffff",
		Name:            "Lava Hounds",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lava Hounds — it deals 4 damage to you", Do(DamageController{N: 4})),
		},
	})
}
