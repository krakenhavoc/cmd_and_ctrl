package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "e0de0121-6185-45c5-9d88-a415b07347e0",
		Name:         "Kari Zev's Expertise",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or Vehicle", effects.Or(effects.Creature(), effects.Subtype("Vehicle"))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Kari Zev's Expertise — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Kari Zev's Expertise — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return KariZevsExpertiseOfferFreeCast(ctx, item.Controller)
		},
	})
}

// kariZevsExpertiseOfferFreeCast is "you may cast a spell with mana
// value 2 or less from your hand without paying its mana cost" —
// Rishkar's Expertise's rishkarsExpertiseOfferFreeCast with the
// ceiling lowered to two.
func KariZevsExpertiseOfferFreeCast(ctx *effects.Context, controller uuid.UUID) error {
	effects.OfferFreeHandCast(ctx, controller, kariZevsExpertiseFreeCast,
		"a spell with mana value 2 or less", effects.SpellWithManaValueAtMost(2))
	return nil
}

// kariZevsExpertiseFreeCast is the offer's keyed continuation (#4246).
var kariZevsExpertiseFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/kari-zevs-expertise-cast", "Kari Zev's Expertise")
