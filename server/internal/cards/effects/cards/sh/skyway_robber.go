package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyway Robber — Creature — Bird Rogue {3}{U}, 3/3 (New Capenna
// Commander, #5880):
//
//	"Flying
//	 Escape—{3}{U}, Exile five other cards from your graveyard. (You may
//	 cast this card from your graveyard for its escape cost.)
//	 This creature escapes with "Whenever this creature deals combat
//	 damage to a player, you may cast an artifact, instant, or sorcery
//	 spell from among cards exiled with this creature without paying its
//	 mana cost.""
//
// CR 702.138d: "[This permanent] escapes with [ability]" means "If this
// permanent escaped, it has [ability]." That is a printed static ability,
// so it is the combat-damage trigger gated on the permanent having escaped
// (CR 702.138b, Card.Escaped, the record Phlage and Uro read): a Robber
// cast from hand, reanimated or blinked has no trigger, a Clone of an
// escaped Robber did not escape and has none, and anything that removes or
// overwrites the printed text takes the trigger with it, as it would the
// static ability.
//
// "Cards exiled with this creature" are the five its escape cost exiled
// (CR 607.2q): escape's exile is linked to the permanent spell and moved
// to the permanent as it enters (#5880, relinkCostExilesLocked), so a
// Robber that died and escaped again reads only its new five (CR 400.7).
// Every card still linked is on offer, including ones passed up on an
// earlier hit. One cast ("a spell"), offered during the resolution (CR
// 608.2g) for no mana, filtered to an artifact, instant or sorcery spell
// (PermissionFilter.ArtifactInstantOrSorceryOnly); declining or casting
// leaves the rest in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "58aac732-3f03-4551-a684-6cb0e3b31c09",
		Name:             "Skyway Robber",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{3}{U}", 5)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.AllOf(effects.ThisDealtCombatDamageToAPlayer, skywayRobberEscaped),
				skywayRobberLabel, skywayRobberCast),
		},
	})
}

const (
	skywayRobberName  = "Skyway Robber"
	skywayRobberLabel = skywayRobberName + " — you may cast an artifact, instant, or sorcery spell from among cards exiled with it without paying its mana cost"
)

// skywayRobberEscaped is CR 702.138d's condition: the permanent escaped.
func skywayRobberEscaped(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Escaped()
}

func skywayRobberCast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	you := ctx.Controller()
	if you == uuid.Nil {
		return nil
	}
	perm := game.CastPermission{
		Player:      you,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Filter:      game.PermissionFilter{ArtifactInstantOrSorceryOnly: true},
		Source:      ctx.Source(),
		SourceName:  skywayRobberName,
		Label:       skywayRobberLabel,
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var linked []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		linked = append(linked, c.InstanceID)
	}
	candidates := effects.FreeCastCandidates(g, &perm, linked)
	if len(candidates) == 0 {
		return nil
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       ctx.Source(),
		Reason:       skywayRobberName,
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: effects.DeclineLeaveInExile,
	}) {
		return nil
	}
	var cards []game.Card
	for _, id := range candidates {
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	g.GrantCastPermissionToCardsForEffect(perm, cards)
	return nil
}
