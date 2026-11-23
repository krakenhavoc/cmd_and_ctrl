package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secluded Steppe — Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 Cycling {W} ({W}, Discard this card: Draw a card.)"
//
// Canyon Slough's shape with one colour: enters tapped, one mana
// ability, and the row below.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8dec6fcf-1254-4b1b-ba23-7a3e492a7241",
		Name:         "Secluded Steppe",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{W}")},
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
