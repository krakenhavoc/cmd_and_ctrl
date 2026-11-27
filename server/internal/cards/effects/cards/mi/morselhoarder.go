package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morselhoarder — Creature — Elemental {4}{R/G}{R/G}, 6/4 (Shadowmoor, S82 OLD
// b532, #5591):
//
//	"This creature enters with two -1/-1 counters on it.
//	 Remove a -1/-1 counter from this creature: Add one mana of any color."
//
// Sphere of the Suns' counter-removal mana ability: it adds mana, targets
// nothing and uses no stack (CR 605.1a), so it can pay for a spell mid-cast. The
// counter is the whole cost, so a creature with none can't activate it. The
// entry counters are CR 614.1c's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fe687aa-d134-4a29-9c93-64086c4a1977",
		Name:         "Morselhoarder",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB532MinusOneCountersAtEntry("Morselhoarder", 2)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{RemoveCounters: effects.RemoveCountersFromThis(game.CounterMinusOne, 1).RemoveCounters},
			Produced: "{W|U|B|R|G}",
			Label:    "Remove a -1/-1 counter from this creature: Add one mana of any color.",
		}},
	})
}
