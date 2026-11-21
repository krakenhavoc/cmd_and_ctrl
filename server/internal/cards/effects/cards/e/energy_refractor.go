package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Energy Refractor — Artifact {2}:
//
//	"When this artifact enters, draw a card.
//	 {2}: Add one mana of any color."
//
// A mana ability with a mana cost and no tap (Prophetic Prism without
// the {T}): it filters two mana of any kind into one of a colour, and
// can be used again and again. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd9dde4c-bacd-48d8-b4c9-a6a04f0b3338",
		Name:         "Energy Refractor",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.SamiDrawOnETB("Energy Refractor")},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Mana: "{2}"}, "{2}: Add one mana of any color"),
		},
	})
}
