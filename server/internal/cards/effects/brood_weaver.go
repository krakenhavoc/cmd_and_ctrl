package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brood Weaver — Creature — Spider {3}{G}, 2/4 (Innistrad: Midnight Hunt, S82):
//
//	"Reach
//	 When this creature dies, create a 1/2 green Spider creature token with reach."
//
// A dies trigger making the 1/2 Spider with reach.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f0a7ef59-7897-4e84-a85e-1fcbe3dc9f6a",
		Name:            "Brood Weaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Brood Weaver — create a 1/2 green Spider creature token with reach",
				Do(CreateToken{Template: TokenCard("1/2 green Spider with reach"), N: 1})),
		},
	})
}
