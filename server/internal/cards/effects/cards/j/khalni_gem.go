package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Khalni Gem — Artifact {4} (Zendikar, S82 OLD b494):
//
//	"When this artifact enters, return two lands you control to their owner's hand.
//	 {T}: Add two mana of any one color."
//
// The mana ability is Gilded Lotus's, for two. The enters trigger is a choice on
// resolution and not a target (the bounce lands' shape, B04BounceLand) of two
// lands; with fewer than two lands, all of them go (CR 609.3). The Gem itself is
// not a land and is never a candidate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0a96e47-91da-43ee-93dd-f866721029c6",
		Name:         "Khalni Gem",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(2),
			Label:    "Add two mana of any one color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Khalni Gem — return two lands you control to their owner's hand", effects.Old494ReturnTwoLands),
		},
	})
}
