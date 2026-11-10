package game

import "github.com/google/uuid"

// TargetRelation binds a clause to the single pick of an earlier clause
// in the same mode occurrence. It is catalog data, re-derived on restore.
// Each pick of the dependent clause is judged against that one anchor on
// its own; only a kind that ManyPicks admits may pick more than one
// object (#4862).
type TargetRelation struct {
	Slot int
	Kind TargetRelationKind
}

type TargetRelationKind string

const (
	TargetInControllersGraveyard TargetRelationKind = "controllers_graveyard"
	TargetAttachedTo             TargetRelationKind = "attached_to"
	// TargetControlledBySameController is "another target creature that
	// player controls", where that player controls the earlier target
	// (Breaking of the Fellowship), or IS the earlier target when it is a
	// player (Ravager of the Fells). An anchor that has left the
	// battlefield answers with its last-known controller (CR 608.2h).
	TargetControlledBySameController TargetRelationKind = "same_controller"
	// TargetSharesCardType is "that shares a card type with it", where it
	// is the earlier target permanent (Burglar's Plot's "two target nonland
	// permanents that share a card type"): both are on the battlefield and
	// have at least one card type in common after the layers (CR 205.2a).
	TargetSharesCardType TargetRelationKind = "shares_card_type"
	// TargetInTargetPlayersGraveyard is "from their graveyard", where
	// they are the earlier target PLAYER (Quandrix Command's "target
	// player shuffles up to three target cards from their graveyard into
	// their library", #4862). The anchor must be a player pick, and the
	// card must be in that player's graveyard. It is the one kind whose
	// dependent clause may pick several cards, each judged on its own: a
	// card that has left the graveyard is dropped while the rest stay
	// legal (CR 608.2b). An anchor that has become an illegal target still
	// names the graveyard, so its cards stay legal; the card's body then
	// does nothing with them, because an illegal player performs no action
	// (the Memory's Journey ruling).
	TargetInTargetPlayersGraveyard TargetRelationKind = "players_graveyard"
	// TargetSharesCreatureType is "that share a creature type", where the
	// other is the earlier target permanent (Secret Tunnel's "two target
	// creatures you control that share a creature type"): judged on the
	// creature types both have after the layers (SharesCreatureType, so a
	// changeling shares with anything that has a creature type, and land or
	// artifact subtypes never count). An anchor that has left the
	// battlefield answers with the creature types it had as it last existed
	// (CR 608.2h, the Secret Tunnel ruling), so the creature still here is
	// still a legal target as long as it shares one of them.
	TargetSharesCreatureType TargetRelationKind = "shares_creature_type"
	// TargetCardSharesCreatureType is TargetSharesCreatureType for two
	// CARDS in a graveyard (Unbury's "two target creature cards that share
	// a creature type"). The anchor is read wherever its card is now, so a
	// card that left the graveyard answers with the creature types it has
	// (a card's types outside the battlefield are its own, CR 205.3d), and
	// the other card is still returned as long as it shares one of them
	// (the Unbury ruling). The zone of the candidate is the clause's own.
	TargetCardSharesCreatureType TargetRelationKind = "card_shares_creature_type"
)

// ManyPicks reports whether a dependent clause of this kind may pick
// more than one object. Such a kind must judge every pick against the
// anchor alone, never against the clause's other picks, because the
// copy and retarget walks refill its picks one at a time
// (relatedTargetsCanFinishLocked).
func (k TargetRelationKind) ManyPicks() bool {
	return k == TargetInTargetPlayersGraveyard
}

// TargetRelationMatchesForEffect judges the relationship using the current
// objects, at announcement and again at resolution (CR 601.2c, 608.2b).
// A missing anchor supplies no information, except to the same-controller
// relationship, which reads the anchor's last-known controller (CR 608.2h);
// a different mode's pick cannot fill it. Caller holds g.mu.
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
	case TargetControlledBySameController:
		controller, ok := g.relationAnchorControllerLocked(anchor)
		return ok && z.Kind == ZoneBattlefield && c.Controller == controller
	case TargetInTargetPlayersGraveyard:
		return anchor.Kind == TargetPlayer && anchor.ID != uuid.Nil && z.Kind == ZoneGraveyard && z.Owner == anchor.ID
	case TargetSharesCardType:
		a := findBattlefieldCard(g, anchor.ID)
		return anchor.Kind == TargetCard && a != nil && z.Kind == ZoneBattlefield &&
			len(SharedCardTypes([]Card{*a, *c})) > 0
	case TargetCardSharesCreatureType:
		if anchor.Kind != TargetCard {
			return false
		}
		a := g.findCardByIDLocked(anchor.ID)
		return a != nil && SharesCreatureType(a, c)
	case TargetSharesCreatureType:
		if anchor.Kind != TargetCard || z.Kind != ZoneBattlefield {
			return false
		}
		if a := findBattlefieldCard(g, anchor.ID); a != nil {
			return SharesCreatureType(a, c)
		}
		info, ok := g.LastKnownPermanentForEffect(anchor.ID)
		if !ok {
			return false
		}
		last := info.Characteristic
		return SharesCreatureType(&Card{effective: &last}, c)
	}
	return false
}

// relationAnchorControllerLocked is "that player" for a relation whose
// anchor is a permanent: its controller now while it is on the
// battlefield, and otherwise the controller it last had there this turn
// (CR 608.2h). An anchor that is still on the battlefield but illegal
// (it gained hexproof, or changed controller) answers with its current
// controller, since it is still in the zone the effect expects. A card
// that never left the battlefield this turn and is not on it answers
// nothing.
//
// An anchor that is a PLAYER is "that player" itself: Ravager of the
// Fells' "target opponent or planeswalker and up to one target creature
// that player or that planeswalker's controller controls" (#4071).
func (g *Game) relationAnchorControllerLocked(anchor TargetRef) (uuid.UUID, bool) {
	if anchor.Kind == TargetPlayer {
		return anchor.ID, anchor.ID != uuid.Nil
	}
	if anchor.Kind != TargetCard {
		return uuid.Nil, false
	}
	if a := findBattlefieldCard(g, anchor.ID); a != nil {
		return a.Controller, true
	}
	if info, ok := g.LastKnownPermanentForEffect(anchor.ID); ok {
		return info.Controller, true
	}
	return uuid.Nil, false
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
