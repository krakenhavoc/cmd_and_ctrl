package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barren Moor — Land:
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 Cycling {B} ({B}, Discard this card: Draw a card.)"
//
// Canyon Slough's shape with a single-colour mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "326ba371-124c-4949-a048-3a0c8962e567",
		Name:         "Barren Moor",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{B}")},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
