package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yue, the Moon Spirit — Legendary Creature — Spirit Ally {3}{U}, 3/3:
//
//	"Flying, vigilance
//	 Waterbend {5}, {T}: You may cast a noncreature spell from your hand
//	 without paying its mana cost."
//
// The cost is WaterbendCost("{5}") plus the tap (CR 701.67a): each of
// the five generic may be paid by tapping an untapped artifact or
// creature you control, Yue herself excepted since she taps for the {T}.
// The effect is Brazen Bluff's free hand cast over "a noncreature spell"
// (offerFreeHandCast): a pick among the noncreature spells in your hand,
// cast during the resolution without paying its mana cost, so X is 0 and
// no alternative cost applies, while additional costs may still be paid
// (the rulings). Declining casts nothing.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "961f1883-2f97-4249-90f9-416279d7a9dd",
		Name:            "Yue, the Moon Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Waterbend {5}, {T}: You may cast a noncreature spell from your hand without paying its mana cost.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.WaterbendCost("{5}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				effects.OfferFreeHandCast(ctx, ctx.Controller(), yueFreeCast, "a noncreature spell",
					effects.FreeCastSpell{Filter: game.PermissionFilter{NoncreatureOnly: true}})
				return nil
			},
		}},
	})
}

var yueFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/yue-the-moon-spirit-cast", "Yue, the Moon Spirit")
