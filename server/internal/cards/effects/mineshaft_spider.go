package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mineshaft Spider — Creature — Spider {3}{G}, 3/4 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Reach
//	 When this creature enters, you may mill two cards."
//
// Reach is the printed keyword (CR 702.17a). The enters trigger is
// optional, asked as it resolves, and mills its controller's top two
// cards (CR 701.17a); declining mills nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e5aa32b6-2eaf-4b48-81cc-cd3db976513d",
		Name:            "Mineshaft Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Mineshaft Spider — mill two cards", Do(MillCards{N: 2})),
				"Mineshaft Spider — mill two cards?"),
		},
	})
}
