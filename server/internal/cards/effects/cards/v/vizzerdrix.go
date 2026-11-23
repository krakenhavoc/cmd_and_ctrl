package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vizzerdrix — Creature — Rabbit Beast {6}{U}, 6/6
// (Starter 1999, #5620):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (Grizzly Bears'
// shape). The cost, types and body ride game.Card from Scryfall, so the
// Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6649a99-c9cd-474e-91f2-cc8e75496864",
		Name:         "Vizzerdrix",
		Completeness: effects.CompletenessFull,
	})
}
