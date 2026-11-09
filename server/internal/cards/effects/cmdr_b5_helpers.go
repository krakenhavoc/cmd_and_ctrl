package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdrB5Dragon is the Dragon creature type, read through the effective
// subtype list so a changeling or a granted type counts.
func cmdrB5Dragon() CardPredicate { return OfCreatureType("Dragon") }

// cmdrB5CountBattlefield counts the battlefield permanents, of any
// controller, that pass pred, read through the layered view.
func cmdrB5CountBattlefield(g *game.Game, pred func(game.Card) bool) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if pred(c) {
			n++
		}
	}
	return n
}

// cmdrB5CountYouControl counts the permanents `controller` controls that
// pass pred.
func cmdrB5CountYouControl(g *game.Game, controller uuid.UUID, pred func(game.Card) bool) int {
	return cmdrB5CountBattlefield(g, func(c game.Card) bool {
		return c.Controller == controller && pred(c)
	})
}

// cmdrB5CountAllGraveyards counts the cards in every graveyard that pass
// pred.
func cmdrB5CountAllGraveyards(g *game.Game, pred func(game.Card) bool) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if pred(c) {
				n++
			}
		}
	}
	return n
}

// cmdrB5DivideGreatestPower is "X damage divided as you choose, where X is
// the greatest power among creatures you control as you cast this spell"
// (Monstrous Onslaught): the amount is read as the targets are announced,
// so it is fixed for the cast (CR 601.2d, 601.2i).
var cmdrB5DivideGreatestPower = game.RegisterDivideAmount("greatest-power-you-control",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		return game.DivideSpec{Total: b42GreatestPowerControlledBy(g, a.Controller)}
	})

// cmdrB5OfferBehold is the question of "you may behold a <subtype>"
// written as an effect (CR 701.4a): choose a <subtype> you control or
// reveal a <subtype> card from your hand. The candidates are the shared
// reveal-cost walk's (the matching cards in hand, then the matching
// permanents), so a cast's behold and this one can't disagree about who
// can behold. With nothing to behold nothing is asked. Declining is a real
// answer (Min 0): the clause is a "may". `then` is a continuation
// registered by the card; it runs with the pick and must call
// cmdrB5RevealBeheld first, as the cost does (payBeholdLocked).
func cmdrB5OfferBehold(g *game.Game, item *game.StackItem, subtype string, then game.ChooseCardsThen, question string) {
	cands := g.RevealCostOptionsForEffect(item.Controller, item.SourceCardID, &game.RevealCost{Subtype: subtype, Behold: true})
	if len(cands) == 0 {
		return
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: question,
		Cards:    cands,
		Min:      0,
		Max:      1,
		ThenKey:  then,
	})
}

// cmdrB5RevealBeheld reveals the beheld card to the table when it was a
// card in the chooser's hand; a chosen permanent shows nothing the
// battlefield did not. It reports whether anything was beheld.
func cmdrB5RevealBeheld(g *game.Game, r game.CardsChosen, subtype string) bool {
	if !r.Answered || len(r.Picked) == 0 {
		return false
	}
	p := g.PlayerByIDForEffect(r.Chooser)
	if p == nil {
		return false
	}
	if p.Hand != nil && p.Hand.Contains(r.Picked[0]) {
		reason := "behold a " + subtype
		if c, ok := g.LookupCardForEffect(r.Source); ok {
			reason = c.Name + " — " + reason
		}
		g.RevealForEffect(game.RevealSpec{Player: r.Chooser, Source: r.Source, Reason: reason, Cards: r.Picked[:1]})
	}
	return true
}
