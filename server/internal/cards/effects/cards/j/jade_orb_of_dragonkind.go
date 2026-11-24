package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jade Orb of Dragonkind — Artifact {2}{G} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP2 b05 #5948, #6088):
//
//	"{T}: Add {G}. When you spend this mana to cast a Dragon creature
//	 spell, it enters with an additional +1/+1 counter on it and gains
//	 hexproof until your next turn."
//
// Carnelian Orb of Dragonkind's shape: the mana carries two spend riders,
// both filtered to a Dragon creature spell. The counter is an
// enters-with-counters mark on the spell (CR 614.1c); the hexproof is an
// ability-adding effect on the spell made at the spend, which the permanent
// the spell becomes keeps until its caster's next turn (CR 400.7a, 611.2b;
// SpentSpellGainsUntilYourNextTurn). Mana spent on anything else does
// neither.
//
// No simplification.
func init() {
	dragonSpell := []string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature"), effects.ManaRestrictSubtype("Dragon")}
	effects.Register(effects.Spec{
		OracleID:       "5eef3d70-ef16-4722-8c3d-21a0311597bd",
		Name:           "Jade Orb of Dragonkind",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
			SpendRiders: []game.ManaSpendRider{
				effects.SpentEntersWithCounters(game.CounterPlusOne, 1, dragonSpell...),
				effects.SpentSpellGainsUntilYourNextTurn([]string{"hexproof"}, dragonSpell...),
			},
		}},
	})
}
