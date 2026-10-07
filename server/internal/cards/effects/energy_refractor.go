package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Energy Refractor — Artifact {2}:
//
//	"When this artifact enters, draw a card.
//	 {2}: Add one mana of any color."
//
// A mana ability with a mana cost and no tap (Prophetic Prism without
// the {T}): it filters two mana of any kind into one of a colour, and
// can be used again and again. No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd9dde4c-bacd-48d8-b4c9-a6a04f0b3338",
		Name:         "Energy Refractor",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{samiDrawOnETB("Energy Refractor")},
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Mana: "{2}"}, "{2}: Add one mana of any color"),
		},
	})
}
