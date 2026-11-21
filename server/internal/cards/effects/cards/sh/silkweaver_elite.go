package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "3d0eb524-15e7-46f2-8f53-db08a07944d7",
		Name:            "Silkweaver Elite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersIfRevolt("Silkweaver Elite — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
