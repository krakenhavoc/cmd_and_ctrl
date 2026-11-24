package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sol Talisman — Artifact with no mana cost (Modern Horizons 2):
//
//	"Suspend 3—{1} (Rather than cast this card from your hand, pay {1} and exile
//	 it with three time counters on it. At the beginning of your upkeep, remove a
//	 time counter. When the last is removed, you may cast it without paying its
//	 mana cost.)
//	 {T}: Add {C}{C}."
//
// Lotus Bloom's shape: suspend is the whole of its playability, and an artifact
// is not held back by summoning sickness, so it taps the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "867f0cdf-cf75-40f0-bf1e-ccb92e55e9a3",
		Name:         "Sol Talisman",
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Suspend(3, "{1}"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
