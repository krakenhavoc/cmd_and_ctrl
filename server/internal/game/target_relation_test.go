package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

func relationSpec() *TargetSpec {
	return (&TargetSpec{Zones: []ZoneKind{ZoneBattlefield}, Min: 1, Max: 1}).Then(&TargetSpec{
		Zones: []ZoneKind{ZoneGraveyard}, Min: 1, Max: 1,
		Relation: &TargetRelation{Slot: 0, Kind: TargetInControllersGraveyard},
	})
}

func TestRelatedRetargetCommitsBothTargetsAndSurvivesUndo(t *testing.T) {
	g, spec, a, b := relationFixture(t)
	id := uuid.New()
	g.StackMeta = make(map[uuid.UUID]*StackItem)
	g.StackMeta[id] = &StackItem{ID: id, Controller: g.Seats[0].ID, Targets: a, targetSpec: spec}
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{ItemID: id, Chooser: g.Seats[0].ID, Policy: RetargetChooseNew}); err != nil {
			t.Fatal(err)
		}
	})
	choice := g.PendingChoices[0]
	if err := g.ResolveRetarget(choice.ID, choice.Chooser, b[:1]); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(g.StackMeta[id].Targets, a) {
		t.Fatal("retarget committed an incomplete pair")
	}
	choice = g.PendingChoices[0]
	if choice.PickTargetMin != 1 || !slices.Contains(choice.PickTargetCards, b[1].ID) {
		t.Fatalf("second prompt: %+v", choice)
	}
	snapshot := g.Clone()
	for run := 0; run < 2; run++ {
		if run > 0 {
			g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
		}
		if err := g.ResolveRetarget(choice.ID, choice.Chooser, b[1:]); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(g.StackMeta[id].Targets, b) {
			t.Fatal("the completed pair was not committed")
		}
	}
}

func relationFixture(t *testing.T) (*Game, *TargetSpec, []TargetRef, []TargetRef) {
	t.Helper()
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	x, y := ebLand(a.ID, "a"), ebLand(b.ID, "b")
	u, v := ebLand(a.ID, "grave a"), ebLand(b.ID, "grave b")
	g.Battlefield.PushTop(x)
	g.Battlefield.PushTop(y)
	a.Graveyard.PushTop(u)
	b.Graveyard.PushTop(v)
	return g, relationSpec(), []TargetRef{{Kind: TargetCard, ID: x.InstanceID}, {Kind: TargetCard, ID: u.InstanceID, Slot: 1}}, []TargetRef{{Kind: TargetCard, ID: y.InstanceID}, {Kind: TargetCard, ID: v.InstanceID, Slot: 1}}
}

func TestTargetRelationAnnouncementResolutionAndRetarget(t *testing.T) {
	g, spec, a, b := relationFixture(t)
	src := SourceChooser(g.Seats[0].ID)
	steps := AnnouncedClauses(spec, nil, nil)
	g.WithWriteLock(func() {
		if err := g.validateAnnouncedTargetsLocked(src, steps, a); err != nil {
			t.Fatal(err)
		}
		wrong := []TargetRef{a[0], b[1]}
		if err := g.validateAnnouncedTargetsLocked(src, steps, wrong); !errors.Is(err, ErrIllegalTarget) {
			t.Fatalf("mismatched pair: %v", err)
		}
		if err := g.retargetCheckLocked(src, steps, a, wrong, RetargetChooseNew); !errors.Is(err, ErrIllegalTarget) {
			t.Fatalf("retarget second only: %v", err)
		}
		if err := g.retargetCheckLocked(src, steps, a, b, RetargetChooseNew); err != nil {
			t.Fatalf("change both: %v", err)
		}
		if !g.copyTargetsCanFinishLocked(src, steps, a, []TargetRef{b[0], a[1]}, 1) {
			t.Fatal("copy picker could not start a legal change of both targets")
		}
		item := &StackItem{Controller: src.Controller, Targets: a, targetSpec: spec}
		if !g.TargetStillLegalForEffect(item, a[1]) {
			t.Fatal("legal graveyard target refused")
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == a[0].ID {
				g.Battlefield.Cards[i].Controller = g.Seats[1].ID
			}
		}
		if g.TargetStillLegalForEffect(item, a[1]) {
			t.Fatal("graveyard stayed legal after the anchor changed controller")
		}
	})
}

func TestTargetRelationCandidatesUseEarlierPicksAndExcludeImpossibleAnchors(t *testing.T) {
	g, spec, a, b := relationFixture(t)
	steps := AnnouncedClauses(spec, nil, nil)
	g.WithWriteLock(func() {
		lt := g.LegalTargetsForAnnouncedClauseForEffect(SourceChooser(g.Seats[0].ID), steps, 1, a[:1])
		if !slices.Contains(lt.Cards, a[1].ID) || slices.Contains(lt.Cards, b[1].ID) {
			t.Fatalf("second picks: %v", lt)
		}
		_, _ = g.Seats[1].Graveyard.Remove(b[1].ID)
		lt = g.LegalTargetsForAnnouncedClauseForEffect(SourceChooser(g.Seats[0].ID), steps, 0, nil)
		if slices.Contains(lt.Cards, b[0].ID) || !slices.Contains(lt.Cards, a[0].ID) {
			t.Fatalf("first picks: %v", lt)
		}
		_, _ = g.Seats[0].Graveyard.Remove(a[1].ID)
		if !g.anyClauseUnfillableLocked(SourceChooser(g.Seats[0].ID), steps) {
			t.Fatal("an impossible pair still had a legal offer")
		}
	})
}

func TestTargetRelationIsScopedToTheModeOccurrence(t *testing.T) {
	g, spec, a, b := relationFixture(t)
	g.WithWriteLock(func() {
		ref := a[1]
		ref.Mode = 1
		if g.TargetRelationMatchesForEffect(spec.Clause(1), ref, []TargetRef{a[0], b[0]}) {
			t.Fatal("another occurrence supplied the anchor")
		}
	})
}

func TestRelatedRetargetKeepsAnAlreadyIllegalDependent(t *testing.T) {
	g, spec, a, b := relationFixture(t)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == a[0].ID {
				g.Battlefield.Cards[i].Controller = g.Seats[1].ID
			}
		}
		next := []TargetRef{b[0], a[1]}
		if err := g.retargetCheckLocked(SourceChooser(g.Seats[0].ID), AnnouncedClauses(spec, nil, nil), a, next, RetargetChooseNew); err != nil {
			t.Fatalf("already-illegal unchanged dependent: %v", err)
		}
	})
}
