package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// True Conviction — Enchantment {3}{W}{W}{W} (EDHREC rank 1829):
//
//	"Creatures you control have double strike and lifelink."
//
// A six-mana team grant, one Layer 6 static over "creatures you
// control" (b16GrantKeywords). Both keywords are engine-honoured:
// double strike runs the two combat damage substeps and lifelink
// gains on every point dealt, so the enchantment does exactly what
// it says the moment it lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc299c1c-50f3-492a-b6b8-a3664bb72ab7",
		Name:         "True Conviction",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B16CreaturesYouControl, "double strike", "lifelink"),
		},
	})
}
