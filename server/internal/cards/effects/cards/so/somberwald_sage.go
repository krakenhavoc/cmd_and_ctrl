package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Somberwald Sage — Creature — Human Druid {2}{G}, 0/1:
//
//	"{T}: Add three mana of any one color. Spend this mana only to cast
//	 creature spells."
//
// Gilded Lotus's one pick of three (OneColorOfAmount) with Food
// Chain's restriction: a cast, of a creature spell. The tag pays any
// part of that spell's total cost, kicker and alternative costs
// included, and never an activated ability — unearth and ninjutsu
// abilities, or a noncreature spell that makes creature tokens (the
// 2012-05-01 rulings). It is a creature's {T} ability, so it can't be
// activated the turn the Sage arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f249db15-09fa-4a1c-9461-9ae803c219b6",
		Name:         "Somberwald Sage",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     effects.OneColorOfAmount(3),
			Label:        "Add three mana of any one color. Spend this mana only to cast creature spells",
			Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature")},
		}},
	})
}
