package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Itazura, Lingering Wick — Legendary Creature — Spirit 1/5 {2}{U}{R}
// (Mystery Booster 2 / MB Commander):
//
//	"At the beginning of your upkeep, exile the top three cards of your
//	 library. Each opponent secretly chooses a number 0 or greater. Then
//	 those numbers are revealed. Choose an opponent with the highest
//	 number. Itazura deals that much damage to them, then they may cast a
//	 spell from among those cards without paying its mana cost. You put a
//	 card from among them that wasn't cast this way into your hand."
//
// The three cards are exiled face up. Then a secret choice round (#5948,
// ADR 0180 amendment 2026-11-24) asks each opponent, in turn order, for a
// number 0 or greater with no ceiling; nobody sees another's number until
// the last is in, and then all are revealed together. The controller
// chooses among the opponents tied for the highest (no prompt when there
// is one). Itazura deals that opponent that much damage (none for 0; it is
// still Itazura's damage if Itazura has left, CR 608.2h), and then that
// opponent is offered one cast of a spell from among the three still in
// exile without paying its mana cost, during the resolution (CR 608.2g).
// Once they have cast or declined, the controller chooses one of the
// exiled cards that wasn't cast and puts it into their hand; the rest stay
// in exile. With no opponent left to choose a number nobody is damaged or
// casts, and the controller still takes a card.
//
// A bot opponent names 0: the number comes back as damage to whoever
// names the highest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55b45ac9-9ce7-4d17-9efb-3c0c51482950",
		Name:         "Itazura, Lingering Wick",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Itazura, Lingering Wick — exile the top three cards of your library; each opponent secretly chooses a number",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), 3, ctx.Tail(itazuraExiled))
				}),
		},
	})
}

// itazuraExiled asks the opponents for their numbers. An on-disk key.
var itazuraExiled = effects.CardTail("itazura-lingering-wick/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	var asks []game.VoteAsk
	for _, opp := range effects.OpponentsInTurnOrder(ctx.Game, ctx.Controller()) {
		asks = append(asks, game.VoteAsk{
			Voter:    opp,
			Question: "Itazura, Lingering Wick — secretly choose a number 0 or greater (the highest is dealt that much damage, then may cast one of the exiled cards free)",
			Number:   &game.VoteNumber{Min: 0, NoMax: true, SelfDamage: true},
		})
	}
	return effects.SecretChoices{
		Question: "Itazura, Lingering Wick — secretly choose a number 0 or greater",
		Asks:     asks,
		Then:     itazuraRevealed,
		Carry:    o.IDs,
	}.Apply(ctx)
})

// itazuraRevealed finds the highest number and who named it. An on-disk
// key.
var itazuraRevealed = effects.VoteResultThen("secret/itazura-lingering-wick", func(ctx *effects.Context, r game.VoteResult) error {
	best := -1
	var tied []uuid.UUID
	for _, b := range r.Ballots {
		if b.Option != game.VoteNumberIndex || !effects.VoterStillIn(ctx.Game, b.Voter) {
			continue
		}
		switch {
		case b.Number > best:
			best, tied = b.Number, []uuid.UUID{b.Voter}
		case b.Number == best:
			tied = append(tied, b.Voter)
		}
	}
	exiled := r.Carry
	switch len(tied) {
	case 0:
		return itazuraToHand(ctx.Game, ctx.Controller(), ctx.Source(), exiled)
	case 1:
		return itazuraDamageThenOffer(ctx, tied[0], best, exiled)
	}
	options := make([]game.ChoiceOption, 0, len(tied))
	for _, id := range tied {
		if p := ctx.Game.PlayerByIDForEffect(id); p != nil {
			options = append(options, game.ChoiceOption{Label: p.Name, Player: id})
		}
	}
	return effects.PickOption{
		Question: "Itazura, Lingering Wick — choose an opponent with the highest number",
		Options:  options,
		ThenKey:  itazuraChosen,
		Carry:    exiled,
		Nums:     []int{best},
	}.Apply(ctx)
})

// itazuraChosen is the tie broken. An on-disk key.
var itazuraChosen = effects.OptionPickThen("option-pick/itazura-lingering-wick-opponent", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Player == uuid.Nil || len(r.Nums) == 0 {
		return itazuraToHand(ctx.Game, ctx.Controller(), ctx.Source(), r.Carry)
	}
	return itazuraDamageThenOffer(ctx, r.Option.Player, r.Nums[0], r.Carry)
})

// itazuraDamageThenOffer deals the damage and offers the chosen opponent
// the free cast; the card for the hand follows the offer.
func itazuraDamageThenOffer(ctx *effects.Context, opp uuid.UUID, n int, exiled []uuid.UUID) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	if n > 0 {
		if err := g.DealDamageToPlayerForEffect(source, opp, n); err != nil {
			return err
		}
	}
	still := effects.AbstractPerformanceStillExiled(g, exiled)
	if len(still) == 0 || !effects.VoterStillIn(g, opp) {
		return itazuraToHand(g, controller, source, still)
	}
	perm := game.CastPermission{
		Player:      opp,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      source,
		SourceName:  "Itazura, Lingering Wick",
		Label:       "Itazura, Lingering Wick — cast a spell from among them without paying its mana cost",
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      opp,
		Source:       source,
		Reason:       "Itazura, Lingering Wick",
		Cards:        still,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Don't cast",
		Then:         itazuraAfterCast,
		Carry:        append([]uuid.UUID{controller}, still...),
	}) {
		return nil
	}
	return itazuraToHand(g, controller, source, still)
}

// itazuraAfterCast is "You put a card from among them that wasn't cast
// this way into your hand", once the cast is answered: Carry is the
// controller, then the cards. An on-disk key.
var itazuraAfterCast = game.RegisterChooseCardsThen("resolution-cast/itazura-lingering-wick-to-hand",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return itazuraToHand(g, r.Carry[0], r.Source, r.Carry[1:])
	})

// itazuraToHand asks the controller for one card among `ids` still in
// exile (none was cast), or takes the only one.
func itazuraToHand(g *game.Game, controller, source uuid.UUID, ids []uuid.UUID) error {
	still := effects.AbstractPerformanceStillExiled(g, ids)
	switch {
	case len(still) == 0 || !effects.VoterStillIn(g, controller):
		return nil
	case len(still) == 1:
		_, err := g.ReturnFromAnyZoneToHandForEffect(controller, source, still[0])
		return err
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "Itazura, Lingering Wick — put a card from among them that wasn't cast into your hand",
		Cards:    still,
		Min:      1,
		Max:      1,
		ThenKey:  itazuraPicked,
		LiveZone: game.ZoneExile,
	})
	return nil
}

// itazuraPicked puts the chosen card into its chooser's hand. An on-disk
// key.
var itazuraPicked = game.RegisterChooseCardsThen("choose-cards/itazura-lingering-wick-to-hand",
	func(g *game.Game, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
				continue
			}
			if _, err := g.ReturnFromAnyZoneToHandForEffect(r.Chooser, r.Source, id); err != nil {
				return err
			}
		}
		return nil
	})
