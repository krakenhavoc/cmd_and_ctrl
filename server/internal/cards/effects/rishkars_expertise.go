package effects

import (
	"github.com/google/uuid"
)

// Rishkar's Expertise — Sorcery {4}{G}{G} (EDHREC rank 243):
//
//	"Draw cards equal to the greatest power among creatures you
//	 control.
//	 You may cast a spell with mana value 5 or less from your hand
//	 without paying its mana cost."
//
// The draw count is `b42GreatestPowerControlledBy`, the same scalar
// Avatar Kyoshi's mana ability and Orcish Siegemaster's trigger read.
// The free cast is a `choose_cards` pick over the hand — the same
// door Brainstorm's tuck-back and Malcolm's discard offer use —
// over the cards with a spell of mana value 5 or less, floor zero
// ("you MAY cast"), then the pass-closed free cast over whichever card
// was picked (offerFreeHandCast, #3270): the cast opens at once and
// ignores timing, CR 608.2g. The cap is the SPELL's, judged on the face
// being cast (#4509): a modal DFC whose back face costs more is cast free
// only as its front, and an Adventure that fits is cast free though its
// creature does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97407cd0-2bd2-4074-94d3-4ec3d243fa78",
		Name:         "Rishkar's Expertise",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The free cast waits for a prompt the draw asks (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: b42GreatestPowerControlledBy(ctx.Game, ctx.Controller())}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return rishkarsExpertiseOfferFreeCast(ctx, ctx.Controller()) }),
		),
	})
}

// rishkarsExpertiseOfferFreeCast is "you may cast a spell with mana
// value 5 or less from your hand without paying its mana cost".
//
// Not handCardsMatching: that helper is put_from_hand.go's "permanent
// card" filter, and "a spell" reaches instants and sorceries too —
// allHandCardIDs (brainstorm.go) is the type-agnostic list, narrowed to
// the cards with a spell of mana value 5 or less (#4509).
func rishkarsExpertiseOfferFreeCast(ctx *Context, controller uuid.UUID) error {
	offerFreeHandCast(ctx, controller, rishkarsExpertiseFreeCast,
		"a spell with mana value 5 or less", spellWithManaValueAtMost(5))
	return nil
}

// rishkarsExpertiseFreeCast is the offer's keyed continuation (#4246).
var rishkarsExpertiseFreeCast = registerFreeHandCastOffer("choose-cards/rishkars-expertise-cast", "Rishkar's Expertise")
