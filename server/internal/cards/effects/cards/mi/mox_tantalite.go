package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mox Tantalite — Artifact with no mana cost (Modern Horizons, #5519):
//
//	"Suspend 3—{0} (Rather than cast this card from your hand, pay {0} and
//	 exile it with three time counters on it. At the beginning of your
//	 upkeep, remove a time counter. When the last is removed, you may cast
//	 it without paying its mana cost.)
//	 {T}: Add one mana of any color."
//
// Lotus Bloom's suspend at 3 for {0}; with no mana cost it can only be
// cast through suspend. The mana is Alloy Myr's any-colour ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc319819-68f9-4c1e-8d60-25a5e3e50669",
		Name:         "Mox Tantalite",
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Suspend(3, "{0}"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
