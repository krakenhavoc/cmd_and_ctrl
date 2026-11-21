package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swindler's Scheme — Enchantment {2}{U} (Streets of New Capenna
// Commander, #5491):
//
//	"Whenever an opponent casts a spell from their hand, you may reveal the
//	 top card of your library. If it shares a card type with that spell,
//	 counter that spell and that opponent may cast the revealed card
//	 without paying its mana cost."
//
// The trigger is an optional one (CR 603.5): it goes on the stack for a
// spell cast from an opponent's hand (the cast event records the zone) and
// asks as it resolves. The reveal shows the card to every seat. The spell
// is named by the trigger's event, not targeted, so one that has left the
// stack is simply not countered, and its card types are read from wherever
// it is now. When the two share a card type (CR 205.2a) the spell is
// countered and the opponent is offered the revealed card, still in the
// Scheme controller's library, as a free cast during this resolution
// (CR 608.2g); a land cannot be cast, and a card not cast stays on top.
//
// No simplification.
func init() {
	const label = "Swindler's Scheme — reveal the top card of your library; if it shares a card type with that spell, counter it and that opponent may cast the revealed card free"
	effects.Register(effects.Spec{
		OracleID:     "e3ff6134-10dd-4565-afcb-0f41a2933d19",
		Name:         "Swindler's Scheme",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
					ev.OldZone == game.ZoneHand
			}, label, s82SncSwindlersSchemeResolve), "Swindler's Scheme — reveal the top card of your library?"),
		},
	})
}

// s82SncSwindlersSchemeThen is the offer's continuation: nothing is left to
// do once the opponent has answered, because the card not cast stays on top
// of its library. An on-disk key: never renamed, never reused.
var s82SncSwindlersSchemeThen = game.RegisterChooseCardsThen("resolution-cast/swindlers-scheme-free-cast",
	func(*game.Game, game.CardsChosen) error { return nil })

func s82SncSwindlersSchemeResolve(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	you, opp, spellID := item.Controller, ev.Actor, ev.CardID
	revealed := g.RevealTopOfLibraryForEffect(you, item.SourceCardID, 1, "Swindler's Scheme — the top card of the library")
	if len(revealed) == 0 {
		return nil
	}
	top, ok := g.LookupCardForEffect(revealed[0])
	spell, spellOK := g.LookupCardForEffect(spellID)
	if !ok || !spellOK || !effects.SharesACardType(top, spell) {
		return nil
	}
	if g.StackItemForEffect(spellID) != nil {
		if err := g.CounterTargetForEffect(spellID); err != nil {
			return err
		}
	}
	perm := g.FreeCastFromLibraryPermission(opp, item.SourceCardID, "Swindler's Scheme")
	// The revealed card was not exiled or moved: if it is not cast it stays
	// where it is, on top of its owner's library, so the lapse moves nothing.
	perm.LapseOnPass = game.LapseStaysInHand
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      opp,
		Source:       item.SourceCardID,
		Reason:       "Swindler's Scheme",
		Cards:        revealed,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it on top of the library",
		Then:         s82SncSwindlersSchemeThen,
		Carry:        revealed,
	})
	return nil
}
