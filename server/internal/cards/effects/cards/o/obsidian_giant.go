package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obsidian Giant — Creature — Giant {4}{R}, 4/4 (Portal Second Age,
// #5625):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7a96612-4fc5-4944-8b74-9eb6f24190b4",
		Name:         "Obsidian Giant",
		Completeness: effects.CompletenessFull,
	})
}
