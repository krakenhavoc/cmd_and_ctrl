package game

import "github.com/google/uuid"

// targeting_restriction.go — ADR 0109 §6 (#1885): "cards in graveyards
// can't be the targets of spells or abilities" (Ground Seal, Dennick,
// Silent Gravestone, Underworld Cerberus) and its narrower kin, Tomik's
// "lands on the battlefield and land cards in graveyards can't be the
// targets of spells or abilities your opponents control".
//
// canBeTargetedBy (keywords.go) answered for the object's OWN shroud,
// hexproof and protection, and returned true at once for anything off
// the battlefield, because those three are abilities of a permanent. A
// rule about a ZONE has no object to live on: it is a static of some
// other permanent, about every card in a zone. So it is its own slot,
// Spec.TargetingRestrictions, read LIVE off the battlefield on every
// targeting question, the way hexproof_bypass.go reads its waiver: no
// state is written, so undo, the snapshot and a restore point have
// nothing to carry, and the source leaving lifts the restriction on the
// next question.
//
// ONE READ, EVERY CALLER. The restriction is asked inside the two
// targeting choke points in targets.go — specMatchesLocked (the legal
// set the view, the legal-move enumerator and LegalTargetsForEffect
// read) and specMatchLocked (the CR 601.2c announce gate and the CR
// 608.2b resolution re-check) — and only when they are TARGETING. A cost
// payment, or a "choose" that is not "target" (CR 115.10a), passes
// targeting=false and is not refused: Ground Seal does not stop delve,
// or "exile a card from a graveyard" as a cost.
//
// CR 601.2c: a player announces "an appropriate object" for each target,
// obeying any effect that says something can't be chosen; CR 101.2 makes
// the "can't" beat any "target card in a graveyard". CR 608.2b: a target
// a restriction now refuses is illegal at resolution, so a Regrowth cast
// before Ground Seal entered fizzles if the Seal is still there when it
// resolves.

// TargetingQuery is everything a targeting restriction may look at.
// Passed by value, as CastQuery is: a restriction is asked inside the
// target walk, and a value copy keeps a card file from mutating the
// board while it decides.
type TargetingQuery struct {
	// Game is the game the target is being chosen in. Read-only.
	Game *Game

	// Card is the candidate target, as it stands in Zone: a permanent's
	// effective characteristics on the battlefield, a card's printed
	// ones in a graveyard.
	Card Card

	// Zone is where the candidate is.
	Zone ZoneKind

	// Controller is the controller of the spell or ability choosing the
	// target — the "your opponents control" of Tomik's clause.
	Controller uuid.UUID

	// Source is the permanent contributing the restriction; its
	// Controller is the "you" of the ABILITY.
	Source Card
}

// TargetingRestriction is one "<these> can't be the targets of spells or
// abilities" static contributed by a permanent on the battlefield. A
// struct of hooks, as CastRestriction and LandPlayRestriction are: a card
// file writes a literal (or calls a constructor in
// cards/effects/targeting_restriction.go), not a type.
type TargetingRestriction struct {
	// Label is the clause as printed ("Cards in graveyards can't be the
	// targets of spells or abilities."). The graveyard's banner shows it.
	Label string

	// Zones are the zones the clause is about. A candidate in any other
	// zone is never asked about, so a graveyard rule costs nothing when a
	// spell targets a creature. Register refuses an empty list.
	Zones []ZoneKind

	// Forbids decides whether this restriction refuses THIS candidate as
	// a target of a spell or ability THIS player controls. Nil forbids
	// nothing. Read-only, under g.mu.
	Forbids func(q TargetingQuery) bool

	// ActiveWhen is the CR 716 / 719 / 721 designation gate, as on
	// CastRestriction.
	ActiveWhen Designation
}

// covers reports whether the restriction is about zone z at all.
func (r TargetingRestriction) covers(z ZoneKind) bool {
	for _, k := range r.Zones {
		if k == z {
			return true
		}
	}
	return false
}

// CatalogTargetingRestrictions is the catalog hook the effects package
// wires at init, mirroring CatalogCastRestrictions. Keyed by
// CatalogAbilityKey: a Ground Seal that has lost its abilities
// (CR 613.1f) restricts nothing.
var CatalogTargetingRestrictions func(key string) []TargetingRestriction

// TargetingRestrictionsForCard returns the restrictions a permanent
// contributes right now: none under an ability-removing effect, and none
// for one whose designation gate is unsatisfied.
func TargetingRestrictionsForCard(c Card) []TargetingRestriction {
	if CatalogTargetingRestrictions == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogTargetingRestrictions(key), func(r TargetingRestriction) Designation {
		return r.ActiveWhen
	})
}

// boundTargetingRestriction pairs a restriction with the permanent that
// contributes it. The pointer is into g.Battlefield.Cards and is valid
// for one walk under the lock that built it.
type boundTargetingRestriction struct {
	r      TargetingRestriction
	source *Card
}

// targetingRestrictions is the live set, gathered once per target walk
// so a legal-target enumeration over a whole graveyard walks the
// battlefield once rather than once per card.
type targetingRestrictions []boundTargetingRestriction

// activeTargetingRestrictionsLocked gathers every targeting restriction
// on the battlefield. Nil on a board with none, which is nearly every
// board. Nil-receiver safe: the exported CanBeTargetedBy has no game.
//
// Caller must hold g.mu (read or write) with fresh layers.
func (g *Game) activeTargetingRestrictionsLocked() targetingRestrictions {
	if g == nil || g.Battlefield == nil || CatalogTargetingRestrictions == nil {
		return nil
	}
	var out targetingRestrictions
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		for _, r := range TargetingRestrictionsForCard(*src) {
			if r.Forbids == nil || len(r.Zones) == 0 {
				continue
			}
			out = append(out, boundTargetingRestriction{r: r, source: src})
		}
	}
	return out
}

// refuses reports whether some restriction in the set refuses `c`, in
// `zone`, as a target of a spell or ability src.Controller controls.
func (rs targetingRestrictions) refuses(g *Game, c *Card, zone ZoneKind, src TargetSource) bool {
	if len(rs) == 0 || c == nil {
		return false
	}
	for _, b := range rs {
		if !b.r.covers(zone) {
			continue
		}
		if b.r.Forbids(TargetingQuery{Game: g, Card: *c, Zone: zone, Controller: src.Controller, Source: *b.source}) {
			return true
		}
	}
	return false
}

// TargetingBan is one live restriction about a zone, for the zone's
// banner: the printed clause and the permanent that prints it.
type TargetingBan struct {
	Label      string
	Source     uuid.UUID
	SourceName string
}

// TargetingBansForZoneForEffect lists the live restrictions that are
// about zone z — the graveyard viewer's "Cards in graveyards can't be
// the targets of spells or abilities. — Ground Seal". It lists the
// clause, not a verdict: Tomik's applies only to land cards and only to
// his opponents' spells, and the banner says so in his own words.
//
// Caller must hold g.mu (read or write) with fresh layers.
func (g *Game) TargetingBansForZoneForEffect(z ZoneKind) []TargetingBan {
	var out []TargetingBan
	for _, b := range g.activeTargetingRestrictionsLocked() {
		if b.r.covers(z) {
			out = append(out, TargetingBan{Label: b.r.Label, Source: b.source.InstanceID, SourceName: b.source.Name})
		}
	}
	return out
}
