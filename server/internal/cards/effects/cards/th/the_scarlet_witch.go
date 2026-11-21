package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Scarlet Witch — {2}{R} Legendary Creature — Mutant Warlock Hero 2/3
// (Marvel Super Heroes, #4455):
//
//	"Instant and sorcery spells you cast with mana value 4 or greater cost
//	 {X} less to cast, where X is The Scarlet Witch's power."
//
// Vadrik's computed reduction (the cost query's Source is the live
// permanent, so a pump or a counter deepens the discount) behind a
// mana-value gate. The mana value is the spell's own, with {X} as announced
// (CR 202.3e), and the reduction spends generic mana only (CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8ae8c38-501e-44ea-9faf-cf9d11ba6616",
		Name:         "The Scarlet Witch",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int { return q.Source.CurrentPower() },
				"Instant and sorcery spells you cast with mana value 4 or greater cost {X} less to cast, where X is The Scarlet Witch's power.",
				effects.YourSpell(), effects.InstantOrSorcerySpell(), effects.SpellManaValueAtLeast(4)),
		},
	})
}
