package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Portent of Calamity — Sorcery {X}{U} (Duskmourn: House of Horror, #4466):
//
//	"Reveal the top X cards of your library. For each card type, you may
//	 exile a card of that type from among them. Put the rest into your
//	 graveyard. You may cast a spell from among the exiled cards without
//	 paying its mana cost if you exiled four or more cards this way. Then
//	 put the rest of the exiled cards into your hand."
//
// Atraxa, Grand Unifier's flow with exile in place of the hand: the X are
// revealed (they stay on the library, public), and the controller is asked
// once per card type in order (artifact, battle, creature, enchantment,
// instant, kindred, land, planeswalker, sorcery), for a card of that type
// still among the revealed cards, optional. A chosen card is exiled at once
// and cannot be chosen again, so a card with two types fills one slot (the
// rulings). When the last type has been asked, the rest still in the library
// go to the graveyard (not a mill, so nothing that watches a mill sees it).
// Then the cards that landed in exile are counted: with four or more the
// controller may cast one spell from among them without paying its mana cost
// as part of the resolution (CR 608.2g; OfferResolutionCastForEffect, Abstract
// Performance's flow: ONE free cast, timing restrictions ignored, a card with
// {X} uses X = 0, additional costs still owed). Then the rest of the exiled
// cards still in exile go to the hand, whether or not a spell was cast. With
// fewer than four, they all go to the hand. Every prompt is keyed and plain
// data, so a table waiting on one is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d49187a-1c37-4c3a-afc2-fac2bbc6b29a",
		Name:         "Portent of Calamity",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, x,
				"Portent of Calamity — reveal the top X cards of your library")
			if len(revealed) == 0 {
				return nil
			}
			return portentAsk(ctx, revealed, 0)
		},
	})
}

// portentCardTypes are the nine card types, in order.
var portentCardTypes = []string{"artifact", "battle", "creature", "enchantment", "instant", "kindred", "land", "planeswalker", "sorcery"}

// portentAsk asks for the card of the first type at or after `from` that has
// a candidate among the revealed cards still on the library, and finishes
// once no type is left.
func portentAsk(ctx *Context, revealed []uuid.UUID, from int) error {
	g, you := ctx.Game, ctx.Controller()
	left := cardsStillInALibrary(g, revealed)
	for i := from; i < len(portentCardTypes); i++ {
		var candidates []uuid.UUID
		for _, id := range left {
			if c, ok := g.LookupCardForEffect(id); ok && c.HasCardType(portentCardTypes[i]) {
				candidates = append(candidates, id)
			}
		}
		if len(candidates) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    you,
			FromPlayer: you,
			Source:     ctx.Source(),
			Question:   "Portent of Calamity — you may exile a " + portentCardTypes[i] + " card from among the revealed cards",
			Cards:      candidates,
			Min:        0,
			Max:        1,
			LiveZone:   game.ZoneLibrary,
			ThenKey:    portentPicked,
			Carry:      revealed,
			CarryNums:  []int{i},
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return portentFinish(ctx, revealed, left)
}

// portentPicked is one type's answer: the card to exile, then the next
// type. An on-disk key.
var portentPicked game.ChooseCardsThen

func init() {
	portentPicked = askedPick("choose-cards/portent-of-calamity-card-type", func(ctx *Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
				continue
			}
			if err := ctx.Game.ExileCardForEffect(id); err != nil {
				return err
			}
		}
		next := 0
		if len(r.Nums) > 0 {
			next = r.Nums[0] + 1
		}
		return portentAsk(ctx, r.Carry, next)
	})
}

// portentFinish puts the rest into the graveyard, then offers the free cast
// and puts what is left in exile into the hand.
func portentFinish(ctx *Context, revealed, rest []uuid.UUID) error {
	next := ctx.ItemTail(portentAfterGraveyard)
	next.Carry = revealed
	if len(rest) == 0 {
		return next.Run(ctx.Game, nil)
	}
	return ctx.Game.PutCardsIntoGraveyardKeyedForEffect(rest, next)
}

// portentAfterGraveyard counts the exiled cards, offers the one free cast
// when there are four or more, and puts the rest into the hand. An on-disk
// key.
var portentAfterGraveyard = CardTail("portent-of-calamity/after-graveyard", func(ctx *Context, o game.TailOutcome) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	var exiled []uuid.UUID
	for _, id := range o.Tail.Carry {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			exiled = append(exiled, id)
		}
	}
	if len(exiled) == 0 {
		return nil
	}
	if len(exiled) >= 4 {
		perm := game.CastPermission{
			Player:      controller,
			Zone:        game.ZoneExile,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        "{0}",
			Timing:      game.TimingFlash,
			CastOnly:    true,
			LapseOnPass: game.LapseStaysInExile,
			Source:      source,
			SourceName:  "Portent of Calamity",
			Label:       "Portent of Calamity — cast it without paying its mana cost",
		}
		if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
			Chooser:      controller,
			Source:       source,
			Reason:       "Portent of Calamity",
			Cards:        exiled,
			Permission:   &perm,
			Keyword:      game.MayCastKeywordFreeCast,
			AcceptLabel:  "Cast it free",
			DeclineLabel: "Put them into your hand",
			Then:         portentRestToHand,
			Carry:        exiled,
		}) {
			return nil
		}
	}
	return gloriousPurposePutRest(g, controller, source, exiled)
})

// portentRestToHand is "Then put the rest of the exiled cards into your
// hand", once the cast is answered: Carry is the exiled cards. An on-disk
// key.
var portentRestToHand = game.RegisterChooseCardsThen("resolution-cast/portent-of-calamity-rest-to-hand",
	func(g *game.Game, r game.CardsChosen) error {
		return gloriousPurposePutRest(g, r.Chooser, r.Source, r.Carry)
	})
