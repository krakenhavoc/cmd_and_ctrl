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

// TargetClause is ONE clause of a target statement: one predicate,
// chosen Min..Max times. It is an ALIAS for TargetSpec, not a second
// type — a TargetSpec *is* its first clause, and a statement with
// more than one clause hangs the rest off it in Rest (#764, ADR
// 0065 §1). That shape is what lets a cost-payment predicate
// (AbilityCost.SacrificeOther and friends), a mode option's clause
// and a two-slot spell all be the same struct walked by the same
// code, with one predicate struct in the tree.
//
// Read a clause list with ClauseCount / Clause (allocation-free) or
// Clauses (convenient). Never read Rest off a clause you got from
// one of those: the list is flat by construction and Register
// refuses a nested Rest at boot.
type TargetClause = TargetSpec

// TargetSpec declares a spell's or ability's target clause: one
// predicate, chosen Min..Max times. "Target creature" is 1 / 1;
// "two target creatures" 2 / 2; "up to three target cards" 0 / 3;
// "any number of targets" 1 / 0 (Max 0 = unbounded). Targets are
// distinct unless AllowSame (CR 115.3: one object can't be chosen
// more than once for a single instance of the word "target").
// Announce-time order is preserved on StackItem.Targets, so a
// positional clause ("2 damage to any target and 1 damage to any
// other target") reads Targets[0] / Targets[1].
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

	// Min / Max bound the number of targets. Max 0 means unbounded.
	//
	// On a SACRIFICE clause (AbilityCost.SacrificeOther,
	// ManaAbilityShape.SacrificeOther, AdditionalCost.Sacrifice) they
	// count the permanents sacrificed instead, and Min == Max == N:
	// "Sacrifice two artifacts" is 2 / 2 (#747). A sacrifice clause
	// is still a predicate, not targeting (CR 601.2h). effects.Register
	// refuses a sacrifice clause with Min != Max, a count below 1,
	// CountFromX, AllowSame or Players — variable counts have no seam
	// yet. See SacrificeCostCount.
	Min, Max int

	// CountFromX makes the clause's target count the announced X
	// rather than a printed constant — "Exile X target creatures you
	// control" (Waterbender's Restoration), where X is defined by the
	// waterbend cost paid at announce. The cast path replaces Min and
	// Max with the announced XValue before validating, and stores the
	// resolved copy on the stack item, so the resolution re-check and
	// every downstream reader see a concrete count.
	//
	// Without this the clause would have to ship as "any number of
	// target creatures", which is what it read as while the waterbend
	// cost was uncharged — and which was strictly stronger than the
	// printed card (issue #259).
	CountFromX bool

	// AllowSame permits the same player / card in more than one
	// slot. Off for every ordinary clause; reserved for effects
	// whose wording uses separate "target" words that may coincide.
	AllowSame bool

	// Distinct makes this clause's picks differ from every EARLIER
	// clause's picks in the same statement — "a SECOND target
	// permanent you control" (Resourceful Defense), "target creature
	// or planeswalker you don't control" after "target creature you
	// control". CR 601.2c lets one object fill two different
	// instances of the word "target" unless the card says otherwise,
	// so this is opt-in per clause rather than the default.
	//
	// AllowSame is the WITHIN-clause twin: it governs whether two
	// picks of the SAME clause may coincide. The two are
	// independent. Added by #764.
	Distinct bool

	// Rest holds clauses 2..n of a multi-clause statement, in
	// printed order. Empty — which is every clause the catalog
	// declared before #764, every cost-payment predicate and every
	// mode option that targets once — means "this statement is one
	// clause", and everything about the spec reads exactly as it did.
	//
	// The list is FLAT: an entry's own Rest is ignored, and
	// effects.Register refuses one at boot. Build it with
	// effects.Clauses(first, then…). Added by #764 (ADR 0065 §1).
	Rest []TargetClause
}

// ClauseCount is the number of clauses in this statement: 1 for the
// ordinary single-clause spec. Nil-safe (0).
func (s *TargetSpec) ClauseCount() int {
	if s == nil {
		return 0
	}
	return 1 + len(s.Rest)
}

// Clause returns clause i of the statement, or nil when i is out of
// range. Clause(0) is the spec itself — which is the whole point of
// the head-and-tail shape — so the returned pointer's Rest is the
// statement's tail and MUST NOT be read as though it were the
// clause's own. Read predicate fields only.
func (s *TargetSpec) Clause(i int) *TargetClause {
	if s == nil || i < 0 || i > len(s.Rest) {
		return nil
	}
	if i == 0 {
		return s
	}
	return &s.Rest[i-1]
}

// Clauses is the allocating form of the same walk, for callers that
// want a range loop (the protocol projection, the bot enumerator).
// Each entry is a VALUE copy with Rest cleared, so nothing downstream
// can mistake the statement's tail for a clause's own.
func (s *TargetSpec) Clauses() []TargetClause {
	n := s.ClauseCount()
	if n == 0 {
		return nil
	}
	out := make([]TargetClause, 0, n)
	for i := 0; i < n; i++ {
		c := *s.Clause(i)
		c.Rest = nil
		out = append(out, c)
	}
	return out
}

// Then appends clauses to this statement and returns it, so a
// catalog constructor chain reads in printed order. Mutates and
// returns the receiver, as WithCount does.
func (s *TargetSpec) Then(more ...*TargetSpec) *TargetSpec {
	for _, m := range more {
		if m == nil {
			continue
		}
		c := *m
		c.Rest = nil
		s.Rest = append(s.Rest, c)
		// A nested statement flattens rather than nesting — the list
		// is flat by construction everywhere it is read.
		s.Rest = append(s.Rest, m.Rest...)
	}
	return s
}

// WithCount returns the spec with its Min / Max replaced — the
// catalog's way to turn a single-target constructor into "two
// target creatures" (2, 2) or "up to three" (0, 3). Mutates and
// returns the receiver for chaining.
func (s *TargetSpec) WithCount(min, max int) *TargetSpec {
	s.Min, s.Max = min, max
	return s
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

// countRealTargets counts the target slots a caster actually filled,
// ignoring the TargetSelf / TargetNone placeholders validateTargetsLocked
// also skips.
func countRealTargets(targets []TargetRef) int {
	n := 0
	for _, t := range targets {
		if t.Kind == TargetSelf || t.Kind == TargetNone {
			continue
		}
		n++
	}
	return n
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
//
// This is the TARGETING enumeration: the protection-style keyword
// gate (CanBeTargetedBy) is applied, so a hexproof creature an
// opponent controls never reaches the picker. Cost payments and
// other "choose" effects that are not targeting want
// specCandidatesLocked instead. Caller must hold g.mu.
func (g *Game) legalTargetsLocked(caster uuid.UUID, spec *TargetSpec) LegalTargets {
	return g.specMatchesLocked(caster, spec, true)
}

// specCandidatesLocked is legalTargetsLocked without the
// protection-style keyword gate: the set of cards that merely MATCH
// the spec's predicate and zones. Cost payments use it, because
// paying a cost is not targeting (CR 601.2f/h) — convoking a
// shrouded creature, or sacrificing one to an additional cost, is
// legal and always has been. sacrifice.go's
// sacrificeCandidatesLocked makes the same distinction for the
// effect-driven sacrifice prompt.
//
// Caller must hold g.mu.
func (g *Game) specCandidatesLocked(chooser uuid.UUID, spec *TargetSpec) LegalTargets {
	return g.specMatchesLocked(chooser, spec, false)
}

// specMatchesLocked is the shared walk behind both enumerations.
// `targeting` switches the CR 702 keyword gate on.
func (g *Game) specMatchesLocked(caster uuid.UUID, spec *TargetSpec, targeting bool) LegalTargets {
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
			for i := range z.Cards {
				c := z.Cards[i]
				if targeting && !CanBeTargetedBy(&c, zk, caster) {
					continue
				}
				if spec.CardOK != nil && !spec.CardOK(g, caster, c, zk) {
					continue
				}
				out.Cards = append(out.Cards, c.InstanceID)
			}
		}
	}
	return out
}

// TargetStillLegalForEffect is the per-slot CR 608.2b check for an
// item mid-resolution: the ref still exists in an allowed zone and
// still passes the item's spec (the one it was announced under).
// Items without a spec fall back to the existence check. Effects
// with several targets call this per slot and skip the illegal ones
// — "if only some targets are illegal, the spell does as much as it
// can" — while the all-illegal fizzle runs before OnResolve fires.
func (g *Game) TargetStillLegalForEffect(item *StackItem, ref TargetRef) bool {
	switch ref.Kind {
	case TargetSelf, TargetNone:
		return true
	case TargetPlayer, TargetCard:
		// #764: the clause the ref was announced under, not the
		// item's first clause — a two-slot spell's second pick is
		// re-checked against its OWN predicate here.
		if clause := g.clauseForRefLocked(item, ref); clause != nil {
			return g.targetLegalLocked(item.Controller, clause, ref)
		}
		return targetStillExistsLocked(g, ref)
	}
	return false
}

// LegalTargetsForEffect is the *ForEffect-surface wrapper around
// legalTargetsLocked for callers already under g.mu (catalog
// HasLegalTarget predicates, trigger target pickers).
func (g *Game) LegalTargetsForEffect(caster uuid.UUID, spec *TargetSpec) LegalTargets {
	return g.legalTargetsLocked(caster, spec)
}

// SpecCandidatesForEffect is the NON-targeting sibling of
// LegalTargetsForEffect: which permanents match this spec for the
// purpose of PAYING A COST. Convoke's tap list, an additional
// sacrifice cost's option list, and an activated ability's
// "sacrifice another creature" clause all reuse TargetSpec as a
// predicate, and none of them targets — so a shrouded creature is
// still convokable and a hexproof one is still sacrificeable.
// Callers already under g.mu (the protocol projection, the bot's
// move enumerator).
//
// Choosing between the two is the one judgement call this file
// asks of a caller: if the printed text says "target", use
// LegalTargetsForEffect; if it says "sacrifice", "tap", "choose"
// or "exile", use this.
func (g *Game) SpecCandidatesForEffect(chooser uuid.UUID, spec *TargetSpec) LegalTargets {
	return g.specCandidatesLocked(chooser, spec)
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
	case ZoneHand:
		// S28: "exile a blue card from your hand" (Force of Will,
		// Solitude). Every seat's hand, narrowed by the spec's own
		// predicate — which for every clause that reaches here says
		// "YOUR hand", so the constructor in the effects package
		// (CardInYourHand) bakes the ownership check in rather than
		// leaving it to a card file to remember.
		//
		// Hidden-zone caution: nothing in the catalog TARGETS a card
		// in hand, and nothing should — a hand is hidden information
		// and a legal-target list over it would leak an opponent's
		// hand size and contents to the picker. The one consumer is
		// the NON-targeting candidate scan (SpecCandidatesForEffect),
		// computed per viewer for their own hand.
		out := make([]*Zone, 0, len(g.Seats))
		for _, p := range g.Seats {
			if p != nil && p.Hand != nil {
				out = append(out, p.Hand)
			}
		}
		return out
	}
	return nil
}

// targetLegalLocked reports whether one announce-time target ref is
// currently legal under spec for caster: the referenced player /
// card exists in an allowed zone, is not shielded from this caster
// by a protection-style keyword, AND passes the predicate. Used at
// announce (CR 601.2c) and again at resolution (CR 608.2b) — the
// same function, which is why a creature that GAINS hexproof in
// response to a spell already on the stack makes that spell fizzle
// without anything else being taught the rule. Self / none refs are
// always legal. Caller must hold g.mu.
func (g *Game) targetLegalLocked(caster uuid.UUID, spec *TargetSpec, ref TargetRef) bool {
	return g.specMatchLocked(caster, spec, ref, true)
}

// specMatchLocked is targetLegalLocked's body with the CR 702
// keyword gate switchable. Cost-payment call sites (convoke's tap
// list, an activated ability's "sacrifice another" clause) pass
// targeting=false: they reuse TargetSpec purely as a predicate over
// permanents, and paying a cost never targets.
func (g *Game) specMatchLocked(caster uuid.UUID, spec *TargetSpec, ref TargetRef, targeting bool) bool {
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
		for i := range z.Cards {
			c := z.Cards[i]
			if c.InstanceID != ref.ID {
				continue
			}
			if targeting && !CanBeTargetedBy(&c, z.Kind, caster) {
				return false
			}
			return spec.CardOK == nil || spec.CardOK(g, caster, c, z.Kind)
		}
	}
	return false
}

// validateTargetsLocked is the announce-time gate (CR 601.2c) for a
// NON-MODAL clause list: each pick must fall in its own clause's
// Min..Max, be legal under that clause's predicate, and (unless
// AllowSame) not repeat within it. Returns ErrInvalidParam for a
// count / duplicate / bad-slot violation and ErrIllegalTarget for an
// illegal pick.
//
// It is a thin wrapper over validateAnnouncedTargetsLocked (#764),
// which is the general form the modal paths use; a single-clause spec
// is that walk with one step. Caller must hold g.mu.
func (g *Game) validateTargetsLocked(caster uuid.UUID, spec *TargetSpec, targets []TargetRef) error {
	return g.validateAnnouncedTargetsLocked(caster, AnnouncedClauses(spec, nil, nil), targets)
}
