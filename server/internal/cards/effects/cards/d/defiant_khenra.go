package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiant Khenra — Creature — Jackal Warrior {1}{R}, 2/2 (Hour of Devastation, #5534):
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
		OracleID:     "0e6fe04f-b20d-4df9-a7ac-edebc3bcc9be",
		Name:         "Defiant Khenra",
		Completeness: effects.CompletenessFull,
	})
}
