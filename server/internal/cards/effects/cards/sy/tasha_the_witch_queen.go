package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tasha, the Witch Queen — {3}{U}{B} Legendary Planeswalker — Tasha, loyalty 4
//
//	"Whenever you cast a spell you don't own, create a 3/3 black Demon
//	 creature token.
//	 +1: Draw a card. For each opponent, exile up to one target instant or
//	 sorcery card from that player's graveyard and put a page counter on it.
//	 −3: You may cast a spell from among cards in exile with page counters
//	 on them without paying its mana cost.
//	 Tasha, the Witch Queen can be your commander."
//
// The trigger counts the spell's OWNER, not who cast it: a card cast from
// an opponent's graveyard or exile counts, your own never does.
//
// +1's targets are one "up to one" clause per opponent, bound to that
// opponent's graveyard (PerOpponentGraveyardClauses), re-checked as the
// ability resolves (CR 608.2b). Each card still there is exiled and, once
// it has landed in exile, gets a page counter. Counters on a card in exile
// are per object (CR 122.2), so the marker lasts exactly as long as the
// card stays exiled.
//
// −3 reads EVERY card in exile with a page counter, whoever put it there
// (another Tasha's included): the cards are not linked to this one. The
// cast is made during the resolution (CR 608.2g) through the
// resolution_cast prompt, one spell, free (a {0} price, X locked at 0); a
// land is never cast and is not offered. Declining leaves the cards in
// exile. Spectral Arcanist's shape (CastsLeft 1 over the set).
//
// "Can be your commander" is the deck importer's, from the printed line.
//
// No simplification.
const tashaPageCounter = "page"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "863a148c-596a-492b-b68e-3e1cbfcb140f",
		Name:            "Tasha, the Witch Queen",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, tashaCastASpellYouDontOwn,
				"Tasha, the Witch Queen — create a 3/3 black Demon creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 black Demon"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card. For each opponent, exile up to one target instant or sorcery card from that player's graveyard and put a page counter on it.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.PerOpponentGraveyardClauses("instant or sorcery card", effects.Or(effects.Instant(), effects.Sorcery())),
				Purpose: game.Purpose{Draws: 1},
				Effect:  tashaPlusOne,
			},
			{
				Label:  "−3: You may cast a spell from among cards in exile with page counters on them without paying its mana cost.",
				Cost:   effects.LoyaltyCost(-3),
				Effect: tashaMinusThree,
			},
		},
	})
}

// tashaCastASpellYouDontOwn is "whenever you cast a spell you don't own".
func tashaCastASpellYouDontOwn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.Owner != uuid.Nil && spell.Owner != source.Controller
}

// tashaPlusOne draws a card, then exiles each still-legal target with a
// page counter.
func tashaPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.ExileCardForEffect(t.ID); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if err := g.AddCounterForEffect(t.ID, tashaPageCounter, 1); err != nil {
			return err
		}
	}
	return nil
}

// tashaMinusThree offers one free cast from among the cards in exile with
// page counters.
func tashaMinusThree(g *game.Game, item *game.StackItem) error {
	var cards []game.Card
	var ids []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.FaceDown || c.IsLand() || c.Counters[tashaPageCounter] < 1 {
			continue
		}
		cards = append(cards, c)
		ids = append(ids, c.InstanceID)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      item.Controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   1,
		LapseOnPass: game.LapseStaysInExile,
		Source:      item.SourceCardID,
		SourceName:  "Tasha, the Witch Queen",
		Label:       "Tasha, the Witch Queen — cast a spell from among cards in exile with page counters without paying its mana cost",
	}, cards)
	effects.OfferResolutionCastOver(g, item.Controller, item.SourceCardID, "Tasha, the Witch Queen",
		game.MayCastKeywordFreeCast, effects.DeclineLeaveInExile, ids...)
	return nil
}
