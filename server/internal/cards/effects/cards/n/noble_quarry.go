package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noble Quarry — Enchantment Creature — Unicorn {2}{G}, 1/1 (Commander 2015,
// S82 OLD #5546):
//
//	"Bestow {5}{G} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 All creatures able to block this creature or enchanted creature do so.
//	 Enchanted creature gets +1/+1."
//
// Boon Satyr's bestow, with Lure's requirement over the permanent itself while
// it is a creature and over the creature it is attached to while it is a
// bestowed Aura (SelfCreatureOrAttached), and the +1/+1 on the host alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d549abd0-fea5-417f-93f9-255cb87ec989",
		Name:         "Noble Quarry",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{5}{G}"),
		},
		Static: []game.StaticAbility{
			effects.BlockRequirementWhere(game.BlockRequirementLure, effects.SelfCreatureOrAttached),
			effects.PumpAttached(1, 1),
		},
	})
}
