package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nathan Drake, Treasure Hunter — Legendary Creature — Human Rogue {U}{B}{R},
// 3/2 (Secret Lair Drop, #5462):
//
//	"First strike
//	 You may spend mana as though it were mana of any color to cast spells
//	 you don't own or to activate abilities of permanents you control but
//	 don't own.
//	 Whenever Nathan Drake attacks, exile the top card of each player's
//	 library. You may cast a spell from among those cards."
//
// The mana rider is an any-colour spend grant for its controller that covers
// two payments (AnyColorSpendStatic.Covers): a CAST whose spell's owner is not
// the payer, and an ACTIVATION whose permanent's owner is not the controller
// paying (ManaSpendContext.SpellOwner, ObjectOwner). It touches nothing else
// the player pays. The attack trigger exiles each player's top card as one
// event, and once they are in exile the controller is offered one cast over
// them at the spell's printed cost (the Bilbo posture: one permission, spent
// by one cast, the cast made as the trigger resolves, CR 608.2g). Declining, or a
// land among them, leaves the cards in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed69b907-272c-4a87-a943-06aaad976fb1",
		Name:            "Nathan Drake, Treasure Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		AnyColorSpend: []game.AnyColorSpendStatic{{
			Label: "You may spend mana as though it were mana of any color to cast spells you don't own or to activate abilities of permanents you control but don't own.",
			Whose: game.AnyColorSpendYou,
			Covers: func(ctx game.ManaSpendContext) bool {
				switch ctx.Purpose {
				case game.SpendPurposeCast:
					return ctx.Payer != uuid.Nil && ctx.SpellOwner != uuid.Nil && ctx.Payer != ctx.SpellOwner
				case game.SpendPurposeActivate:
					return ctx.Controller != uuid.Nil && ctx.Controller == ctx.Payer && ctx.ObjectOwner != uuid.Nil && ctx.ObjectOwner != ctx.Controller
				}
				return false
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Nathan Drake — exile the top card of each player's library; you may cast a spell from among those cards",
				nathanDrakeAttack),
		},
	})
}

const nathanDrakeName = "Nathan Drake, Treasure Hunter"

func nathanDrakeAttack(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Library == nil {
			continue
		}
		if top, err := p.Library.Top(); err == nil {
			ids = append(ids, top.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, effects.NewContext(g, item).ItemTail(nathanDrakeExiled))
}

// nathanDrakeExiled offers the one cast over what reached exile. An on-disk
// key: never renamed, never reused.
var nathanDrakeExiled = effects.CardTail("nathan-drake/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g, you, source := ctx.Game, ctx.Controller(), ctx.Source()
	perm := game.CastPermission{
		Player:      you,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   1,
		LapseOnPass: game.LapseStaysInExile,
		Source:      source,
		SourceName:  nathanDrakeName,
		Label:       nathanDrakeName + " — cast a spell from among the exiled cards",
	}
	var cards []game.Card
	var offered []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok && perm.OffersASpell(c) {
			cards = append(cards, c)
			offered = append(offered, id)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(perm, cards)
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       source,
		Reason:       nathanDrakeName,
		Cards:        offered,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordPaidCast,
		AcceptLabel:  "Cast it",
		DeclineLabel: effects.DeclineLeaveInExile,
	})
	return nil
})
