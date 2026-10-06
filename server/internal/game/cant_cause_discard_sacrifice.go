package game

import "github.com/google/uuid"

// cant_cause_discard_sacrifice.go — "spells and abilities your opponents
// control can't cause you to discard cards / sacrifice permanents"
// (#2178): Tamiyo, Collector of Tales (both), Sigarda, Host of Herons
// and Tajuru Preserver (sacrifice only).
//
// Read live off the battlefield every time an EFFECT asks a player to
// discard or sacrifice, keyed by CatalogAbilityKey, and never stored
// (the LegendRuleExemption pattern): two sources compose, one leaving
// cannot revoke the other, and a permanent that lost its abilities
// protects nobody.
//
// WHO CAUSES IT. The effect's cause is the controller of the resolving
// stack item (Game.resolving, the same slot MoveCause reads). The clause
// protects against OPPONENTS' spells and abilities only, so the player's
// own Fleshbag Marauder, or a delayed "sacrifice it" they control, still
// works.
//
// WHAT IS NOT GATED, on purpose. A COST is never caused by an opponent
// (CR 118.3, 601.2h): a sacrifice or discard paid to cast or activate
// goes through the cost paths (DiscardCauseCost, sacrificeAnsweredLocked)
// and never asks this. The cleanup-step discard (CR 514.1) and a state-
// based action (the Saga sacrifice, CR 704.5s) are rules, not spells or
// abilities, and sit on their own paths for the same reason. Game.resolving
// outlives its resolution until the next one starts, so the gate is
// asked only from the explicit effect entry points rather than from the
// shared sacrifice body, where it would catch those rule-driven moves.

// OpponentEffectProtection is one printed "spells and abilities your
// opponents control can't cause you to …" static. Catalog data.
type OpponentEffectProtection struct {
	// Label is the clause as printed.
	Label string
	// Discard: opponents' effects can't make the controller discard.
	Discard bool
	// Sacrifice: opponents' effects can't make the controller sacrifice.
	Sacrifice bool
}

// CatalogOpponentEffectProtections returns the protections a permanent
// with the given catalog key has. carddef.go sets it from CardDef.
var CatalogOpponentEffectProtections func(key string) []OpponentEffectProtection

// effectCauseControllerLocked is the controller of the spell or ability
// currently resolving, or uuid.Nil when none is. Caller must hold g.mu.
func (g *Game) effectCauseControllerLocked() uuid.UUID {
	if g.resolving == nil || g.resolving.item == nil {
		return uuid.Nil
	}
	return g.resolving.item.Controller
}

// opponentEffectBlockedLocked reports whether the resolving effect is an
// opponent's and `victim` controls a permanent that stops it. Caller must
// hold g.mu (read or write). Reads only.
func (g *Game) opponentEffectBlockedLocked(victim uuid.UUID, discard bool) bool {
	return g.opponentEffectBlockedByLocked(g.effectCauseControllerLocked(), victim, discard)
}

// opponentEffectBlockedByLocked is opponentEffectBlockedLocked with the
// cause named rather than read off the resolving item. A discard made
// when a prompt is answered, after the resolution that asked for it
// (the revealed-hand pick's variants, ADR 0116's 2026-10-05 amendment,
// #2115), is caused by the controller of the effect that asked, and the
// prompt records that player as its chooser. Caller must hold g.mu
// (read or write). Reads only.
func (g *Game) opponentEffectBlockedByLocked(cause, victim uuid.UUID, discard bool) bool {
	if cause == uuid.Nil || cause == victim || victim == uuid.Nil {
		return false
	}
	if CatalogOpponentEffectProtections == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if src.Controller != victim {
			continue
		}
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, p := range CatalogOpponentEffectProtections(key) {
			if (discard && p.Discard) || (!discard && p.Sacrifice) {
				return true
			}
		}
	}
	return false
}

// effectDiscardBlockedLocked: an opponent's resolving spell or ability
// can't make `victim` discard. Caller must hold g.mu.
func (g *Game) effectDiscardBlockedLocked(victim uuid.UUID) bool {
	return g.opponentEffectBlockedLocked(victim, true)
}

// effectSacrificeBlockedLocked: an opponent's resolving spell or ability
// can't make `victim` sacrifice. Caller must hold g.mu.
func (g *Game) effectSacrificeBlockedLocked(victim uuid.UUID) bool {
	return g.opponentEffectBlockedLocked(victim, false)
}

// unprotectedFromSacrificeLocked drops every permanent whose controller
// is protected from the resolving effect. Caller must hold g.mu.
func (g *Game) unprotectedFromSacrificeLocked(ids []uuid.UUID) []uuid.UUID {
	if g.effectCauseControllerLocked() == uuid.Nil {
		return ids
	}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if g.effectSacrificeBlockedLocked(g.controllerOfBattlefieldCardLocked(id)) {
			continue
		}
		out = append(out, id)
	}
	return out
}
