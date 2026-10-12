package game

import "github.com/google/uuid"

// TargetRelation binds a clause to the single pick of an earlier clause
// in the same mode occurrence. It is catalog data, re-derived on restore.
type TargetRelation struct {
	Slot int
	Kind TargetRelationKind
}

type TargetRelationKind string

const (
	TargetInControllersGraveyard TargetRelationKind = "controllers_graveyard"
	TargetAttachedTo             TargetRelationKind = "attached_to"
)

// TargetRelationMatchesForEffect judges the relationship using the current
// objects, at announcement and again at resolution (CR 601.2c, 608.2b).
// A missing anchor supplies no information; a different mode's pick cannot
// fill it. Caller holds g.mu.
func (g *Game) TargetRelationMatchesForEffect(clause *TargetClause, ref TargetRef, targets []TargetRef) bool {
	if clause == nil || clause.Relation == nil {
		return true
	}
	r := clause.Relation
	var anchor TargetRef
	n := 0
	for _, t := range targets {
		if t.Mode == ref.Mode && t.Slot == r.Slot && (t.Kind == TargetCard || t.Kind == TargetPlayer) {
			anchor = t
			n++
		}
	}
	if n != 1 || ref.Kind != TargetCard {
		return false
	}
	z := g.findCardZoneLocked(ref.ID)
	if z == nil {
		return false
	}
	c := g.findCardByIDLocked(ref.ID)
	if c == nil {
		return false
	}
	switch r.Kind {
	case TargetInControllersGraveyard:
		a := findBattlefieldCard(g, anchor.ID)
		return anchor.Kind == TargetCard && a != nil && z.Kind == ZoneGraveyard && z.Owner == a.Controller
	case TargetAttachedTo:
		return anchor.Kind == TargetCard && findBattlefieldCard(g, anchor.ID) != nil && z.Kind == ZoneBattlefield && c.AttachedTo.ID == anchor.ID
	}
	return false
}

// LegalTargetsForClauseForEffect narrows the ordinary candidate set by the
// picks already made. It is shared by trigger prompts and the bot enumerator.
func (g *Game) LegalTargetsForClauseForEffect(src TargetSource, step AnnouncedClause, targets []TargetRef) LegalTargets {
	lt := g.legalTargetsLocked(src, &step.Clause)
	lt = withoutPicked(lt, targets, step.Clause.Distinct)
	if step.Clause.Relation == nil {
		return lt
	}
	out := LegalTargets{}
	for _, id := range lt.Players {
		ref := TargetRef{Kind: TargetPlayer, ID: id, Mode: step.Mode, Slot: step.Slot}
		if g.TargetRelationMatchesForEffect(&step.Clause, ref, targets) {
			out.Players = append(out.Players, id)
		}
	}
	for _, id := range lt.Cards {
		ref := TargetRef{Kind: TargetCard, ID: id, Mode: step.Mode, Slot: step.Slot}
		if g.TargetRelationMatchesForEffect(&step.Clause, ref, targets) {
			out.Cards = append(out.Cards, id)
		}
	}
	return out
}

// LegalTargetsForAnnouncedClauseForEffect also omits an earlier pick
// that cannot fill a required dependent clause. This prevents a picker
// from offering the first half of an impossible announcement.
func (g *Game) LegalTargetsForAnnouncedClauseForEffect(src TargetSource, steps []AnnouncedClause, index int, targets []TargetRef) LegalTargets {
	step := steps[index]
	lt := g.legalTargetsLocked(src, &step.Clause)
	hasDependent := false
	for _, later := range steps[index+1:] {
		if r := later.Clause.Relation; r != nil && later.Mode == step.Mode && r.Slot == step.Slot && later.Clause.Min > 0 {
			hasDependent = true
			break
		}
	}
	if step.Clause.Relation != nil && targets != nil {
		lt = g.LegalTargetsForClauseForEffect(src, step, targets)
	}
	if !hasDependent {
		return lt
	}
	fits := func(id uuid.UUID, kind TargetRefKind) bool {
		picks := append(append([]TargetRef(nil), targets...), TargetRef{Kind: kind, ID: id, Mode: step.Mode, Slot: step.Slot})
		for _, later := range steps[index+1:] {
			if r := later.Clause.Relation; r != nil && later.Mode == step.Mode && r.Slot == step.Slot && later.Clause.Min > 0 {
				candidates := g.LegalTargetsForClauseForEffect(src, later, picks)
				if g.fillableCountLocked(&later.Clause, candidates) < later.Clause.Min {
					return false
				}
			}
		}
		return true
	}
	out := LegalTargets{}
	for _, id := range lt.Players {
		if fits(id, TargetPlayer) {
			out.Players = append(out.Players, id)
		}
	}
	for _, id := range lt.Cards {
		if fits(id, TargetCard) {
			out.Cards = append(out.Cards, id)
		}
	}
	return out
}

// TargetRelationPairsForEffect projects the relationship without teaching
// clients Magic rules: anchor id -> compatible candidate ids. Public zones
// only; the target clause still decides which objects may be offered.
func (g *Game) TargetRelationPairsForEffect(src TargetSource, spec *TargetSpec, slot int) map[uuid.UUID][]uuid.UUID {
	c := spec.Clause(slot)
	if c == nil || c.Relation == nil {
		return nil
	}
	a := spec.Clause(c.Relation.Slot)
	if a == nil {
		return nil
	}
	anchors := g.legalTargetsLocked(src, a)
	out := make(map[uuid.UUID][]uuid.UUID)
	add := func(id uuid.UUID, kind TargetRefKind) {
		picks := []TargetRef{{Kind: kind, ID: id, Slot: c.Relation.Slot}}
		lt := g.LegalTargetsForClauseForEffect(src, AnnouncedClause{Slot: slot, Clause: *c}, picks)
		out[id] = append(append([]uuid.UUID{}, lt.Players...), lt.Cards...)
	}
	for _, id := range anchors.Players {
		add(id, TargetPlayer)
	}
	for _, id := range anchors.Cards {
		add(id, TargetCard)
	}
	return out
}
