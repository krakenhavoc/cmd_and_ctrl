package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dense Foliage — Enchantment {2}{G} (Weatherlight, S82 older sets batch 778, #5629):
//
//	"Creatures can't be the targets of spells."
//
// A layer-6 static giving every creature on the battlefield, whoever controls it, a spell-only
// targeting bar (#6196), read at the one targeting gate, so a spell already aimed at a creature
// fizzles if Dense Foliage arrives before it resolves (CR 608.2b). Abilities may still target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb1369aa-25a2-4e2a-9cbe-6cfe7e238b1b",
		Name:         "Dense Foliage",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB778CreaturesCantBeTargetedBySpells("Creatures can't be the targets of spells."),
		},
	})
}
