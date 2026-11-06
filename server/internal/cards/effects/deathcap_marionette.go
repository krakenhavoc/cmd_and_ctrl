package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathcap Marionette — Creature — Fungus {1}{B}, 1/1 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Deathtouch
//	 When this creature enters, you may mill two cards."
//
// Deathtouch is the printed keyword (CR 702.2a). The enters trigger is
// Mineshaft Spider's: optional, asked as it resolves, and it mills its
// controller's top two cards (CR 701.17a); declining mills nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6c4247d0-6dc8-4139-bbf5-c83dfdc7be1f",
		Name:            "Deathcap Marionette",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Deathcap Marionette — mill two cards", Do(MillCards{N: 2})),
				"Deathcap Marionette — mill two cards?"),
		},
	})
}
