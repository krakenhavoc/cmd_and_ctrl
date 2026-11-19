package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tendo Ice Bridge — Land (Modern Horizons 3 Commander, #5948):
//
//	"This land enters with a charge counter on it.
//	 {T}: Add {C}.
//	 {T}, Remove a charge counter from this land: Add one mana of any
//	 color."
//
// The counter goes on through one entry replacement (b10EntersWithCounters),
// on the card before any trigger or state check sees it. The second mana
// ability's cost has a counter component (Channeler Initiate's shape): the
// counter is removed as the ability is announced and an empty Bridge can't
// activate it (CR 602.5). It stays a mana ability (CR 605.1a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e006a4b-8dde-4416-8cb4-8401562d0fd5",
		Name:         "Tendo Ice Bridge",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterCharge, 1, "Tendo Ice Bridge: enters with a charge counter"),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, RemoveCounters: RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}, Remove a charge counter from this land: Add one mana of any color",
			},
		},
	})
}
