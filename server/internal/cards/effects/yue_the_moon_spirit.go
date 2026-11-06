package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "961f1883-2f97-4249-90f9-416279d7a9dd",
		Name:            "Yue, the Moon Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "Waterbend {5}, {T}: You may cast a noncreature spell from your hand without paying its mana cost.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(WaterbendCost("{5}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				offerFreeHandCast(ctx, ctx.Controller(), yueFreeCast, "a noncreature spell",
					freeCastSpell{filter: game.PermissionFilter{NoncreatureOnly: true}})
				return nil
			},
		}},
	})
}

var yueFreeCast = registerFreeHandCastOffer("choose-cards/yue-the-moon-spirit-cast", "Yue, the Moon Spirit")
