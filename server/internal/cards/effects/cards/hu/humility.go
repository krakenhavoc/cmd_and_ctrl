package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Humility — Enchantment {2}{W}{W} (Tempest, S82 OLD batch 765, #5628):
//
//	"All creatures lose all abilities and have base power and toughness 1/1."
//
// Two layers over every creature on the battlefield, whoever controls it: the
// removal of every ability in layer 6 (CR 613.1f) and the base 1/1 in layer
// 7b (CR 613.4b). A +1/+1 counter, an anthem or an Aura's pump still lands on
// top of the 1/1, and an ability granted later by a timestamp-later effect
// still works (CR 613.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed7bdb3e-5c51-4547-9266-76a791e0b2b0",
		Name:         "Humility",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			humilityLoseAbilities(),
			{
				Label:     "All creatures have base power and toughness 1/1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: humilityApplies,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 1, 1
				},
			},
		},
	})
}

func humilityApplies(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsCreature()
}

func humilityLoseAbilities() game.StaticAbility {
	s := effects.LoseAllAbilities()
	s.Label = "All creatures lose all abilities."
	s.AppliesTo = humilityApplies
	return s
}
