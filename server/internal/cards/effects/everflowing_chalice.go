package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Everflowing Chalice — Artifact {0} (EDHREC rank 254):
//
//	"Multikicker {2} (You may pay an additional {2} any number of
//	 times as you cast this spell.)
//	 This artifact enters with a charge counter on it for each time
//	 it was kicked.
//	 {T}: Add {C} for each charge counter on this artifact."
//
// Three primitives built for exactly this card and composed here:
// `Multikicker` (#664, additional_cost.go) for the repeatable cost,
// `CountersPerKick` (entry_counters.go — its own doc comment names
// Everflowing Chalice's charge counter as the reason it exists) for
// the entry count, and a `ProducedFunc` reading the counters back off
// the mana ability's own source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a79237e-0811-4a8a-bd4d-db3ca91bff22",
		Name:         "Everflowing Chalice",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			Multikicker("{2}", 20),
		},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersPerKick(game.CounterCharge, 1),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: everflowingChaliceProduced,
				Label:        "{T}: Add {C} for each charge counter on this artifact",
			},
		},
	})
}

// everflowingChaliceProduced is "Add {C} for each charge counter on
// this artifact" — read off the ability's own source, not the board.
func everflowingChaliceProduced(g *game.Game, _, source uuid.UUID) string {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return ""
	}
	return strings.Repeat("{C}", c.Counters[game.CounterCharge])
}
