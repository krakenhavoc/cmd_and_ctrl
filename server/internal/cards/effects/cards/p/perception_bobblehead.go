package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f0bcdc43-c43d-4ce7-a283-b8dcab4fc497",
		Name:         "Perception Bobblehead",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Look at the top X cards of your library, where X is the number of Bobbleheads you control. You may cast a spell with mana value 3 or less from among them without paying its mana cost. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.PipBobbleheadsYouControl(g, ctx.Controller())
				if x <= 0 {
					return nil
				}
				return effects.LookAtTopAndOfferFreeCast(g, ctx.Controller(), ctx.Source(), perceptionBobbleheadFreeCast, x, effects.SpellWithManaValueAtMost(3))
			},
		}},
	})
}

var perceptionBobbleheadFreeCast = effects.RegisterFreeLibraryCastOffer("choose-cards/perception-bobblehead-cast", "Perception Bobblehead")
