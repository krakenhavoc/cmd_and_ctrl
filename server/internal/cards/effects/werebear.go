package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Werebear — Creature — Human Bear Druid {1}{G}, 1/1 (Odyssey, #5613):
//
//	"{T}: Add {G}.
//	 Threshold — This creature gets +3/+3 as long as there are seven or
//	 more cards in your graveyard."
//
// Llanowar Elves' mana ability, so summoning sickness applies to it (CR
// 302.6), and the layer 7c threshold self-modifier (CR 613.4c) Krosan
// Beast uses, at +3/+3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96cf3320-c730-4618-8768-28a179f4d9c0",
		Name:         "Werebear",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Static: []game.StaticAbility{fraThresholdSelfPT(3, 3)},
	})
}
