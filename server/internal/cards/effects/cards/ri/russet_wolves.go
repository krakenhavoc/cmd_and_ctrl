package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Russet Wolves — Creature — Wolf {3}{R}, 3/3 (Dark Ascension, #5571):
//
//	(no rules text)
//
// Vanilla, as Memnite: registered so the card is audited and published
// as complete. The body, cost and type line are printed card data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e788e9dd-449d-495b-adad-6d49e408b003",
		Name:         "Russet Wolves",
		Completeness: effects.CompletenessFull,
	})
}
