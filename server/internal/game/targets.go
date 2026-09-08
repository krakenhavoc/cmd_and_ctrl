package game

import "github.com/google/uuid"

// targets.go — S20 sub-PR 1: structured targeting. A catalog card
// declares WHAT it can target as a TargetSpec (which zones, which
// players, and a predicate over candidates); the engine turns that
// into the legal-target set the client's picker shows, rejects
// illegal targets at announce (CR 601.2c), and re-checks the same
// predicate at resolution (CR 608.2b) so a Doom Blade whose target
// turned black in response is countered by game rules — not just
// one whose target left.
//
// Before S20 the only structure was Spec.TargetMode, a client hint
// ("creature") with no server-side check: any card ID the client
// sent was accepted. TargetMode survives as the derived hint the
// client uses for banner copy; TargetSpec is the truth.

// TargetSpec declares a spell's or ability's single target slot.
// Multi-target ("up to two target creatures", "divide damage") is
// S22 territory — Min / Max are carried now so the wire shape
// doesn't churn, but sub-PR 1 only ever sets 1 / 1.
type TargetSpec struct {
	// Mode is the client-facing hint derived from the spec: "any",
	// "player", "creature", "permanent", "stack_spell",
	// "card_in_graveyard". Drives banner copy and which surfaces
	// enter targeting mode; legality itself comes from
	// LegalTargets, not from Mode.
	Mode string

	// Label is the human-readable targeting clause, shown in the
	// picker banner: "target non-black creature".
	Label string

	// Players allows seated, non-eliminated players as targets,
	// filtered by PlayerOK when set.
	Players bool

	// Zones lists the card zones a card target may live in
	// (ZoneBattlefield, ZoneStack, ZoneGraveyard). Empty means no
	// card targets.
	Zones []ZoneKind

	// CardOK is the candidate predicate for card targets. Receives
	// the live game, the caster, the candidate, and the zone it was
	// found in. Nil means every card in Zones qualifies.
	//
	// Runs under g.mu (read or write). MUST NOT call public locking
	// mutators; read-only *ForEffect accessors are fine.
	CardOK func(g *Game, caster uuid.UUID, c Card, zone ZoneKind) bool

	// PlayerOK is the candidate predicate for player targets. Nil
	// means every seated, non-eliminated player qualifies. Same
	// locking rule as CardOK.
	PlayerOK func(g *Game, caster uuid.UUID, p *Player) bool

	// Min / Max bound the number of targets. Sub-PR 1: 1 / 1.
	Min, Max int
}

// CatalogTargetSpec is the catalog hook the effects package wires
// at init. Nil (no catalog) or a nil return (card has no structured
// targeting) means "fall back to the S13.1 free-form picker".
var CatalogTargetSpec func(oracleID string) *TargetSpec

// TargetSpecFor returns the structured targeting for a card, or nil.
func TargetSpecFor(oracleID string) *TargetSpec {
	if CatalogTargetSpec == nil || oracleID == "" {
		return nil
	}
	return CatalogTargetSpec(oracleID)
}

// LegalTargets is the set of targets a spec accepts right now, for
// the given caster. Players and cards are returned separately so
// the wire can carry them as two ID lists.
type LegalTargets struct {
	Players []uuid.UUID
	Cards   []uuid.UUID
}

// legalTargetsLocked computes the legal-target set for spec from
// the caster's point of view. Walks players, then the requested
// zones in a stable order (battlefield, stack, then each seat's
// graveyard in seat order) so the wire list is deterministic.
// Caller must hold g.mu.
func (g *Game) legalTargetsLocked(caster uuid.UUID, spec *TargetSpec) LegalTargets {
	var out LegalTargets
	if spec == nil {
		return out
	}
	if spec.Players {
		for _, p := range g.Seats {
			if p == nil || p.Eliminated {
				continue
			}
			if spec.PlayerOK != nil && !spec.PlayerOK(g, caster, p) {
				continue
			}
			out.Players = append(out.Players, p.ID)
		}
	}
	for _, zk := range spec.Zones {
		for _, z := range g.zonesOfKindLocked(zk) {
			for _, c := range z.Cards {
				if spec.CardOK != nil && !spec.CardOK(g, caster, c, zk) {
					continue
				}
				out.Cards = append(out.Cards, c.InstanceID)
			}
		}
	}
	return out
}

// LegalTargetsForEffect is the *ForEffect-surface wrapper around
// legalTargetsLocked for callers already under g.mu (catalog
// HasLegalTarget predicates, trigger target pickers).
func (g *Game) LegalTargetsForEffect(caster uuid.UUID, spec *TargetSpec) LegalTargets {
	return g.legalTargetsLocked(caster, spec)
}

// LegalTargetsFor is the locking entry point used by the protocol
// projection: which targets could `caster` pick for the card with
// this oracle ID right now. Nil / empty when the card has no
// structured targeting.
func (g *Game) LegalTargetsFor(caster uuid.UUID, oracleID string) *LegalTargets {
	spec := TargetSpecFor(oracleID)
	if spec == nil {
		return nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	lt := g.legalTargetsLocked(caster, spec)
	return &lt
}

// zonesOfKindLocked returns every zone of the given kind: the one
// shared battlefield / stack, or every seat's graveyard. Caller must
// hold g.mu.
func (g *Game) zonesOfKindLocked(kind ZoneKind) []*Zone {
	switch kind {
	case ZoneBattlefield:
		if g.Battlefield != nil {
			return []*Zone{g.Battlefield}
		}
	case ZoneStack:
		if g.Stack != nil {
			return []*Zone{g.Stack}
		}
	case ZoneExile:
		if g.Exile != nil {
			return []*Zone{g.Exile}
		}
	case ZoneGraveyard:
		out := make([]*Zone, 0, len(g.Seats))
		for _, p := range g.Seats {
			if p != nil && p.Graveyard != nil {
				out = append(out, p.Graveyard)
			}
		}
		return out
	}
	return nil
}

// targetLegalLocked reports whether one announce-time target ref is
// currently legal under spec for caster: the referenced player /
// card exists in an allowed zone AND passes the predicate. Used at
// announce (CR 601.2c) and again at resolution (CR 608.2b). Self /
// none refs are always legal. Caller must hold g.mu.
func (g *Game) targetLegalLocked(caster uuid.UUID, spec *TargetSpec, ref TargetRef) bool {
	switch ref.Kind {
	case TargetSelf, TargetNone:
		return true
	case TargetPlayer:
		if !spec.Players {
			return false
		}
		p := g.playerByIDLocked(ref.ID)
		if p == nil || p.Eliminated {
			return false
		}
		return spec.PlayerOK == nil || spec.PlayerOK(g, caster, p)
	case TargetCard:
		z := g.findCardZoneLocked(ref.ID)
		if z == nil {
			return false
		}
		allowed := false
		for _, zk := range spec.Zones {
			if zk == z.Kind {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
		for _, c := range z.Cards {
			if c.InstanceID == ref.ID {
				return spec.CardOK == nil || spec.CardOK(g, caster, c, z.Kind)
			}
		}
	}
	return false
}

// validateTargetsLocked is the announce-time gate (CR 601.2c): the
// number of targeted slots must fall within Min..Max and each must
// be legal. Returns ErrInvalidParam for a count violation and
// ErrIllegalTarget for a bad pick. Caller must hold g.mu.
func (g *Game) validateTargetsLocked(caster uuid.UUID, spec *TargetSpec, targets []TargetRef) error {
	n := 0
	for _, t := range targets {
		if t.Kind == TargetSelf || t.Kind == TargetNone {
			continue
		}
		n++
		if !g.targetLegalLocked(caster, spec, t) {
			return ErrIllegalTarget
		}
	}
	if n < spec.Min || (spec.Max > 0 && n > spec.Max) {
		return ErrInvalidParam
	}
	return nil
}
