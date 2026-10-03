package game

import (
	"testing"

	"github.com/google/uuid"
)

// targeting_restriction_test.go — ADR 0109 §6 (#1885): a static about a
// ZONE ("cards in graveyards can't be the targets of spells or
// abilities") refuses a candidate at both targeting choke points — the
// legal-target walk (the view, the enumerator) and the announce /
// resolution check (CR 601.2c, 608.2b) — and never refuses a cost or a
// "choose" that is not targeting. The real cards are tested in
// internal/cards/effects, internal/legal and internal/protocol.

const graveyardTargetBanOracle = "test-graveyard-target-ban"

// stubTargetingRestrictions wires CatalogTargetingRestrictions for one
// oracle ID and restores the previous hook at the end of the test.
func stubTargetingRestrictions(t *testing.T, oracleID string, rules []TargetingRestriction) {
	t.Helper()
	prev := CatalogTargetingRestrictions
	CatalogTargetingRestrictions = func(id string) []TargetingRestriction {
		if id == oracleID {
			return rules
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { CatalogTargetingRestrictions = prev })
}

// seedTargetBan puts a permanent under `controller` that carries `rules`.
func seedTargetBan(t *testing.T, g *Game, controller uuid.UUID, rules []TargetingRestriction) uuid.UUID {
	t.Helper()
	stubTargetingRestrictions(t, graveyardTargetBanOracle, rules)
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Test Seal", TypeLine: "Enchantment",
		OracleID: graveyardTargetBanOracle, Owner: controller, Controller: controller,
	})
	return id
}

// targetVerdicts is every targeting question asked about one card: in
// the legal-target walk, at the announce / resolution check, and (as a
// cost) in the candidate walk and the non-targeting check.
type targetVerdicts struct {
	legal, announce, costWalk, costCheck bool
}

func verdictsFor(g *Game, by uuid.UUID, spec *TargetSpec, id uuid.UUID) targetVerdicts {
	var v targetVerdicts
	src := TargetSource{Controller: by}
	ref := TargetRef{Kind: TargetCard, ID: id}
	g.ReadSnapshot(func() {
		v.legal = containsID(g.legalTargetsLocked(src, spec).Cards, id)
		v.announce = g.specMatchLocked(src, spec, ref, true)
		v.costWalk = containsID(g.specCandidatesLocked(by, spec).Cards, id)
		v.costCheck = g.specMatchLocked(src, spec, ref, false)
	})
	return v
}

// "Cards in graveyards can't be the targets of spells or abilities":
// refused at both targeting choke points, for every player, and never as
// a cost; a creature on the battlefield is untouched; and the source
// leaving lifts it on the next question.
func TestGraveyardTargetBanRefusesEveryTargetingQuestion(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := NewCard("Dead Bear", opp.ID)
	dead.TypeLine = "Creature — Bear"
	opp.Graveyard.PushTop(dead)
	live := NewCard("Live Bear", opp.ID)
	live.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(live)
	inGraveyard := &TargetSpec{Zones: []ZoneKind{ZoneGraveyard}}
	onBattlefield := &TargetSpec{Zones: []ZoneKind{ZoneBattlefield}}

	if v := verdictsFor(g, me.ID, inGraveyard, dead.InstanceID); !v.legal || !v.announce {
		t.Fatalf("setup: the graveyard card is not a legal target before the ban: %+v", v)
	}
	seal := seedTargetBan(t, g, me.ID, []TargetingRestriction{{
		Label:   "Cards in graveyards can't be the targets of spells or abilities.",
		Zones:   []ZoneKind{ZoneGraveyard},
		Forbids: func(TargetingQuery) bool { return true },
	}})

	for _, by := range []*Player{me, opp} {
		v := verdictsFor(g, by.ID, inGraveyard, dead.InstanceID)
		if v.legal || v.announce {
			t.Errorf("%s may target a graveyard card under the ban: %+v", by.Name, v)
		}
		if !v.costWalk || !v.costCheck {
			t.Errorf("%s's cost may not use a graveyard card under the ban (paying a cost is not targeting): %+v", by.Name, v)
		}
	}
	if v := verdictsFor(g, me.ID, onBattlefield, live.InstanceID); !v.legal || !v.announce {
		t.Errorf("a creature on the battlefield was refused by a graveyard ban: %+v", v)
	}

	g.WithWriteLock(func() { _ = g.ExileCardForEffect(seal) })
	if v := verdictsFor(g, me.ID, inGraveyard, dead.InstanceID); !v.legal || !v.announce {
		t.Errorf("the ban outlived its source: %+v", v)
	}
}

// The query carries the controller of the spell or ability choosing the
// target and the restricting permanent, so "your opponents control" is
// expressible, and Zones keeps a restriction off the zones it is not
// about.
func TestTargetingRestrictionReadsTheQuery(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := NewCard("Dead Forest", me.ID)
	land.TypeLine = "Basic Land — Forest"
	me.Graveyard.PushTop(land)
	var seen TargetingQuery
	seedTargetBan(t, g, me.ID, []TargetingRestriction{{
		Label: "Land cards in graveyards can't be the targets of spells or abilities your opponents control.",
		Zones: []ZoneKind{ZoneGraveyard},
		Forbids: func(q TargetingQuery) bool {
			seen = q
			return q.Controller != q.Source.Controller && q.Card.IsLand()
		},
	}})
	spec := &TargetSpec{Zones: []ZoneKind{ZoneGraveyard}}
	if v := verdictsFor(g, opp.ID, spec, land.InstanceID); v.legal || v.announce {
		t.Errorf("an opponent may target the land card: %+v", v)
	}
	if seen.Zone != ZoneGraveyard || seen.Card.InstanceID != land.InstanceID || seen.Source.Name != "Test Seal" || seen.Controller != opp.ID {
		t.Errorf("the query = %+v, want the land, its zone, the opponent and the source", seen)
	}
	if v := verdictsFor(g, me.ID, spec, land.InstanceID); !v.legal || !v.announce {
		t.Errorf("the restriction's controller may not target their own land card: %+v", v)
	}
	var bans []TargetingBan
	g.ReadSnapshot(func() { bans = g.TargetingBansForZoneForEffect(ZoneGraveyard) })
	if len(bans) != 1 || bans[0].SourceName != "Test Seal" {
		t.Errorf("the graveyard's bans = %+v, want the Seal's", bans)
	}
	g.ReadSnapshot(func() { bans = g.TargetingBansForZoneForEffect(ZoneBattlefield) })
	if len(bans) != 0 {
		t.Errorf("the battlefield carries a graveyard ban: %+v", bans)
	}
}
