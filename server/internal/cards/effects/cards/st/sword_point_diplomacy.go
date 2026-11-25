package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sword-Point Diplomacy — Sorcery {2}{B} (Ixalan, S82 older sets):
//
//	"Reveal the top three cards of your library. For each of those cards, put
//	 that card into your hand unless any opponent pays 3 life. Then exile the
//	 rest."
//
// The three cards are revealed and named once (they stay on the library until
// something moves them). Then, card by card in the order revealed, the offer
// goes to your opponents in turn order (AnyPlayerPays with Opponents: the
// controller is not asked, the first opponent who pays ends that card's
// offer, CR 119.4: only a player with 3 life can). A card nobody paid for goes
// to your hand; a paid-for card stays on the library. After the third, every
// revealed card still on the library is exiled. Each card's offer is a link
// of one chain: the carried list is the cards still to offer, a separator and
// then all three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac5585ce-2959-4eaf-af56-5dca8cdfb2d6",
		Name:         "Sword-Point Diplomacy",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var revealed []uuid.UUID
			if err := (effects.RevealTopOfLibrary{N: 3, Reason: "Sword-Point Diplomacy", Revealed: &revealed}).Apply(ctx); err != nil {
				return err
			}
			return swordPointOffer(ctx, revealed, revealed)
		},
	})
}

// swordPointOffer offers the first of the pending cards to the opponents, or,
// with none left, exiles the revealed cards that are still on the library.
func swordPointOffer(ctx *effects.Context, pending, all []uuid.UUID) error {
	g := ctx.Game
	if len(pending) == 0 {
		for _, id := range all {
			if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
				if err := g.ExileCardForEffect(id); err != nil {
					return err
				}
			}
		}
		return nil
	}
	carry := append([]uuid.UUID(nil), pending...)
	carry = append(carry, uuid.Nil)
	carry = append(carry, all...)
	return effects.AnyPlayerPays{
		Pay:       effects.LifePayment(3),
		Opponents: true,
		Question:  "Sword-Point Diplomacy — pay 3 life so the controller doesn't put this revealed card into their hand?",
		Then:      swordPointAnswered,
		Carry:     carry,
	}.Apply(ctx)
}

// swordPointAnswered is one card's outcome: put it into the hand if nobody
// paid, then offer the next one. An on-disk key.
var swordPointAnswered game.AnyPlayerThen

// Assigned in init: a var initialiser would be an initialisation cycle
// (the consequence offers the next card, which names the consequence).
func init() {
	swordPointAnswered = effects.AnyPlayerConsequences("sword-point-diplomacy/card",
		func(ctx *effects.Context, r game.AnyPlayerAnswer) error { return swordPointNext(ctx, r, false) },
		func(ctx *effects.Context, r game.AnyPlayerAnswer) error { return swordPointNext(ctx, r, true) },
		nil)
}

func swordPointNext(ctx *effects.Context, r game.AnyPlayerAnswer, nobodyPaid bool) error {
	var pending, all []uuid.UUID
	seenSep := false
	for _, id := range r.Carry {
		switch {
		case id == uuid.Nil:
			seenSep = true
		case seenSep:
			all = append(all, id)
		default:
			pending = append(pending, id)
		}
	}
	if len(pending) == 0 {
		return nil
	}
	g := ctx.Game
	if nobodyPaid {
		if z := g.FindCardZoneForEffect(pending[0]); z != nil && z.Kind == game.ZoneLibrary {
			if err := g.BounceToHandForEffect(pending[0]); err != nil {
				return err
			}
		}
	}
	return swordPointOffer(ctx, pending[1:], all)
}
