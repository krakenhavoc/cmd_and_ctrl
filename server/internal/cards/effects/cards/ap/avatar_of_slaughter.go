package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avatar of Slaughter — Creature — Avatar {6}{R}{R}, 8/8 (Commander
// Masters):
//
//	"All creatures have double strike and attack each combat if able."
//
// Absolute Grace's keyword grant over every creature at the table and Grand
// Melee's attack requirement over the same set (CR 508.1d): the Avatar's
// controller's creatures and every opponent's, the Avatar itself included.
// The requirement is a layer-6 ability read as attackers are declared, so a
// tapped or summoning-sick creature, or one that can't attack, is simply
// not required (CR 508.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fc6305e-e75e-478a-826b-49f655ed832c",
		Name:         "Avatar of Slaughter",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B29AllCreatures, "double strike"),
			effects.AttacksEachCombatWhere(effects.B29AllCreatures),
		},
	})
}
