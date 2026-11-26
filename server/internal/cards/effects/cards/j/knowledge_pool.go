package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Knowledge Pool — Artifact {6} (Mirrodin Besieged, S82 OLD b467, #5576):
//
//	"Imprint — When this artifact enters, each player exiles the top three
//	 cards of their library.
//	 Whenever a player casts a spell from their hand, that player exiles it. If
//	 the player does, they may cast a spell from among other cards exiled with
//	 this artifact without paying its mana cost."
//
// Both exiles are LINKED to the Pool as an object (CR 607.2a,
// ExileLinkedKeyedForEffect), so a Pool that left and came back is a new
// object with nothing exiled with it (CR 400.7). The entry exiles each
// player's top three in turn order from the active player (CR 101.4).
//
// The cast trigger is Possibility Storm's: one trigger over every player's casts
// from their hand (the cast event's OldZone). The spell is exiled from the
// stack without being countered; once it is in exile it is linked to the Pool
// (RelinkExiledWithForEffect), so it is one of the "other cards" for the NEXT
// cast. "If the player does": a spell that already left the stack is not
// exiled, so nothing is offered. The caster, not the Pool's controller, is
// offered one free cast during the resolution (CR 608.2g,
// OfferResolutionCastForEffect) over every nonland card linked to the Pool
// except the spell just exiled; a card not cast stays in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb53bd72-dee8-46fa-b0f9-1eee3a99379a",
		Name:         "Knowledge Pool",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Knowledge Pool — imprint: each player exiles the top three cards of their library", knowledgePoolImprint),
			effects.On(game.EventCast, knowledgePoolCastFromHand,
				"Knowledge Pool — that player exiles it; if they do, they may cast a spell from among other cards exiled with this artifact without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Trigger == nil {
						return nil
					}
					return effects.ExileTargetSpell{
						StackID:  item.Trigger.Event.CardID,
						ThenTail: ctx.ItemTail(knowledgePoolExiled),
					}.Apply(ctx)
				}),
		},
	})
}

// knowledgePoolCastFromHand is "a player casts a spell from their hand".
func knowledgePoolCastFromHand(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.OldZone == game.ZoneHand && ev.Actor != uuid.Nil
}

// knowledgePoolImprint exiles the top three of every library, linked to the
// Pool, one player at a time in turn order.
func knowledgePoolImprint(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	link, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	for _, p := range effects.ApnapPlayers(g) {
		ids := g.LookAtTopOfLibraryForEffect(p, 3)
		if len(ids) == 0 {
			continue
		}
		if err := g.ExileLinkedKeyedForEffect(link, ids, game.KeyedTail{}); err != nil {
			return err
		}
	}
	return nil
}

// knowledgePoolExiled runs once the spell has left the stack: link it to the
// Pool, then offer the caster a free cast of one of the other linked cards.
// An on-disk key: never renamed, never reused.
var knowledgePoolExiled = effects.CardTail("knowledge-pool/exiled-offer-cast", func(ctx *effects.Context, o game.TailOutcome) error {
	item := ctx.Item
	if item == nil || item.Trigger == nil {
		return nil
	}
	spellID, caster := item.Trigger.Event.CardID, item.Trigger.Event.Actor
	exiled := false
	for _, id := range o.IDs {
		if id == spellID {
			exiled = true
		}
	}
	if !exiled {
		return nil
	}
	g := ctx.Game
	link, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	g.RelinkExiledWithForEffect(spellID, game.PermissionCardRef(link))
	var others []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(link)) {
		if c.InstanceID != spellID {
			others = append(others, c.InstanceID)
		}
	}
	if len(others) == 0 {
		return nil
	}
	perm := game.CastPermission{
		Player:      caster,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   1,
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  "Knowledge Pool",
		Label:       "Knowledge Pool — cast it without paying its mana cost",
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      caster,
		Source:       ctx.Source(),
		Reason:       "Knowledge Pool",
		Cards:        others,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Don't cast a spell",
	})
	return nil
})
