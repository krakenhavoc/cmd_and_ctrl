package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kari Zev's Expertise — Sorcery {1}{R}{R}:
//
//	"Gain control of target creature or Vehicle until end of turn.
//	 Untap it. It gains haste until end of turn.
//	 You may cast a spell with mana value 2 or less from your hand
//	 without paying its mana cost."
//
// The theft is Act of Treason's three primitives (#756) over a wider
// target — "creature or Vehicle" is Agonasaur Rex's clause,
// Or(Creature(), Subtype("Vehicle")) — and the free cast is Rishkar's
// Expertise's second sentence with the mana-value ceiling lowered from
// five to two: a choose_cards pick over the hand, floor zero ("you
// MAY"), over the cards with a spell of mana value 2 or less (judged on
// the face being cast, #4509: Valki, God of Lies is cast free as Valki and
// never as Tibalt, Cosmic Impostor), then the
// pass-closed free cast over whichever card was picked (offerFreeHandCast,
// #3270): the cast opens at once and ignores timing, CR 608.2g.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0de0121-6185-45c5-9d88-a415b07347e0",
		Name:         "Kari Zev's Expertise",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or Vehicle", Or(Creature(), Subtype("Vehicle"))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Kari Zev's Expertise — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Kari Zev's Expertise — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return kariZevsExpertiseOfferFreeCast(ctx, item.Controller)
		},
	})
}

// kariZevsExpertiseOfferFreeCast is "you may cast a spell with mana
// value 2 or less from your hand without paying its mana cost" —
// Rishkar's Expertise's rishkarsExpertiseOfferFreeCast with the
// ceiling lowered to two.
func kariZevsExpertiseOfferFreeCast(ctx *Context, controller uuid.UUID) error {
	offerFreeHandCast(ctx, controller, kariZevsExpertiseFreeCast,
		"a spell with mana value 2 or less", spellWithManaValueAtMost(2))
	return nil
}

// kariZevsExpertiseFreeCast is the offer's keyed continuation (#4246).
var kariZevsExpertiseFreeCast = registerFreeHandCastOffer("choose-cards/kari-zevs-expertise-cast", "Kari Zev's Expertise")
