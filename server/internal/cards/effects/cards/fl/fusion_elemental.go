package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fusion Elemental — Creature — Elemental {W}{U}{B}{R}{G}, 8/8
// (Alara Reborn reprint):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it. The cost, types
// and body ride game.Card from Scryfall, so the Spec is the key and
// the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a152674f-be29-40ab-8dd9-376fd4eb3bb8",
		Name:         "Fusion Elemental",
		Completeness: effects.CompletenessFull,
	})
}
