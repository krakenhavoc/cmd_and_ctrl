package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lowland Giant — Creature — Giant {2}{R}{R}, 4/3 (Tempest, S82 OLD, #5628):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see hill_giant.go). The
// cost, types and body ride game.Card from Scryfall, so the Spec is the key and
// the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "041cc2d0-eb67-484e-8da1-e91ddb2b3f70",
		Name:         "Lowland Giant",
		Completeness: effects.CompletenessFull,
	})
}
