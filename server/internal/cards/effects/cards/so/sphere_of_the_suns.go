package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sphere of the Suns — Artifact {2} (Mirrodin Besieged, S82 OLD batch 464,
// #5576):
//
//	"This artifact enters tapped and with three charge counters on it.
//	 {T}, Remove a charge counter from this artifact: Add one mana of any
//	 color."
//
// SelfEntersTappedWithCounters for the entry and Pentad Prism's
// charge-counter mana ability with a tap added to its cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c709413-dacf-48af-8ac9-8f61b2699bf1",
		Name:         "Sphere of the Suns",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedWithCounters(game.CounterCharge, 3)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, RemoveCounters: effects.RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Remove a charge counter from this artifact: Add one mana of any color.",
		}},
	})
}
