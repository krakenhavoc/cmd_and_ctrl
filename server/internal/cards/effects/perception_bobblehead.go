package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perception Bobblehead — Artifact — Bobblehead {3} (Fallout, #5948):
//
//	"{T}: Add one mana of any color.
//	 {3}, {T}: Look at the top X cards of your library, where X is the
//	 number of Bobbleheads you control. You may cast a spell with mana value
//	 3 or less from among them without paying its mana cost. Put the rest
//	 on the bottom of your library in a random order."
//
// Cosmic Cube's look-and-cast (lookAtTopAndOfferFreeCast) with X counted as
// the ability resolves and a mana value cap of 3 judged on the face being
// cast. Everything not cast, the declined cards included, goes to the
// bottom in a random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0bcdc43-c43d-4ce7-a283-b8dcab4fc497",
		Name:         "Perception Bobblehead",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Look at the top X cards of your library, where X is the number of Bobbleheads you control. You may cast a spell with mana value 3 or less from among them without paying its mana cost. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := pipBobbleheadsYouControl(g, ctx.Controller())
				if x <= 0 {
					return nil
				}
				return lookAtTopAndOfferFreeCast(g, ctx.Controller(), ctx.Source(), perceptionBobbleheadFreeCast, x, spellWithManaValueAtMost(3))
			},
		}},
	})
}

var perceptionBobbleheadFreeCast = registerFreeLibraryCastOffer("choose-cards/perception-bobblehead-cast", "Perception Bobblehead")
