package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mass of Ghouls — Creature — Zombie Warrior {3}{B}{B}, 5/3
// (Future Sight, #5594):
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
		OracleID:     "29646827-7b69-48c5-8273-4e66076e2384",
		Name:         "Mass of Ghouls",
		Completeness: effects.CompletenessFull,
	})
}
