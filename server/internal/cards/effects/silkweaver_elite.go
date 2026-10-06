package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silkweaver Elite — Creature — Elf Archer {2}{G}, 2/2:
//
//	"Reach
//	 Revolt — When this creature enters, if a permanent left the
//	 battlefield under your control this turn, draw a card."
//
// Revolt is an intervening if (revolt.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d0eb524-15e7-46f2-8f53-db08a07944d7",
		Name:            "Silkweaver Elite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersIfRevolt("Silkweaver Elite — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
