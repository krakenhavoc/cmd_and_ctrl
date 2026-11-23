package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pernicious Deed — Enchantment {1}{B}{G} (Apocalypse, #5614):
//
//	"{X}, Sacrifice this enchantment: Destroy each artifact, creature,
//	 and enchantment with mana value X or less."
//
// Blast Zone's announced-X activated cost (ctx.X() at resolution) with
// Wrath of the Skies' sweep: DestroyAllMatching over artifacts,
// creatures and enchantments of mana value X or less, every player's.
// The Deed is sacrificed as part of the cost, so it is not among them.
// X = 0 is a legal announcement and destroys only the mana-value-0 ones
// (tokens among them). Lands and planeswalkers that are none of the
// three types are spared, and an indestructible permanent survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62e44e0d-eda0-4275-8367-49dab9a087c3",
		Name:         "Pernicious Deed",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, Sacrifice this enchantment: Destroy each artifact, creature, and enchantment with mana value X or less.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.DestroyAllMatching{Match: effects.And(effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment()), effects.ManaValueLE(ctx.X()))}.Apply(ctx)
			},
		}},
	})
}
