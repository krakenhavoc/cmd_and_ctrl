package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Ranger — Creature — Elf Ranger {2}{G}, 4/1 (Alliances, #5634):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "092f2c34-1f75-4998-b219-2cf1ca73656d",
		Name:         "Elvish Ranger",
		Completeness: effects.CompletenessFull,
	})
}
