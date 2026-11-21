package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prophetic Prism — Artifact {2}:
//
//	"When this artifact enters, draw a card.
//	 {1}, {T}: Add one mana of any color."
//
// The draw is an ETB trigger (samiDrawOnETB); the filter is a mana
// ability whose {1} is paid from the pool and whose colour pick is the
// ordinary pipe slot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "134a9877-5cfb-4a2a-a0f0-930dce45f58b",
		Name:         "Prophetic Prism",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.SamiDrawOnETB("Prophetic Prism")},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Tap: true, Mana: "{1}"}, "{1}, {T}: Add one mana of any color"),
		},
	})
}
