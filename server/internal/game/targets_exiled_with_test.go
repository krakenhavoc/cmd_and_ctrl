package game

import (
	"testing"

	"github.com/google/uuid"
)

// targets_exiled_with_test.go -- #3259. A clause over cards in exile
// (Zones {ZoneExile}) and TargetSpec.ExiledWithSource, CR 607.2a's
// "exiled with this". Both are judged in the one shared walk, so the
// legal set, the announce check (CR 601.2c) and the resolution re-check
// (CR 608.2b) agree.

func exileSpec(withSource bool) *TargetSpec {
	return &TargetSpec{
		Mode: "card_in_exile", Zones: []ZoneKind{ZoneExile},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		ExiledWithSource: withSource,
		Min:              1, Max: 1,
	}
}

// pushExiled puts a creature card into exile, linked to `by` (the zero
// ref for "exiled with nothing").
func pushExiled(g *Game, owner uuid.UUID, by PermissionCardRef, faceDown bool) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(Card{
		InstanceID: id, Name: "Exiled Bear", TypeLine: "Creature -- Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		ExiledWith: by, FaceDown: faceDown,
	})
	return id
}

func sourceRef(g *Game, id uuid.UUID) PermissionCardRef {
	var c Card
	g.ReadSnapshot(func() { c = *findBattlefieldCard(g, id) })
	return PermissionCardRef{ID: id, Epoch: c.ObjectEpoch}
}

func legalIn(g *Game, src TargetSource, spec *TargetSpec, id uuid.UUID) bool {
	var ok bool
	g.WithWriteLock(func() {
		ok = g.targetLegalLocked(src, spec, TargetRef{Kind: TargetCard, ID: id})
	})
	return ok
}

// Without the flag, every creature card in exile qualifies, whoever
// exiled it: a plain "target creature card in exile".
func TestCardInExileClauseOffersExiledCards(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	crystal := pushScopedTestCreature(g, me, 0, 0)
	linked := pushExiled(g, opp, sourceRef(g, crystal), false)
	loose := pushExiled(g, opp, PermissionCardRef{}, false)
	// A non-creature in exile is filtered by CardOK as in any zone.
	g.Exile.PushTop(Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land -- Island", Owner: opp, Controller: opp})

	got := legalIDs(g, sourceOf(g, me, crystal), exileSpec(false))
	if len(got) != 2 || !got[linked] || !got[loose] {
		t.Errorf("exile clause offered %v, want the two creature cards", got)
	}
}

// "Exiled with this" narrows to the source's own link, as the object it
// is: another source's exile, an unlinked card and a stale epoch are out.
func TestExiledWithSourceOffersOnlyTheSourcesOwnExile(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	crystal := pushScopedTestCreature(g, me, 0, 0)
	other := pushScopedTestCreature(g, me, 0, 0)
	mine := pushExiled(g, opp, sourceRef(g, crystal), false)
	theirs := pushExiled(g, opp, sourceRef(g, other), false)
	unlinked := pushExiled(g, opp, PermissionCardRef{}, false)
	ref := sourceRef(g, crystal)
	stale := pushExiled(g, opp, PermissionCardRef{ID: ref.ID, Epoch: ref.Epoch + 1}, false)

	spec := exileSpec(true)
	src := sourceOf(g, me, crystal)
	got := legalIDs(g, src, spec)
	if len(got) != 1 || !got[mine] {
		t.Errorf("exiled-with-this offered %v, want only %s", got, mine)
	}
	for name, id := range map[string]uuid.UUID{"another source's": theirs, "unlinked": unlinked, "stale-epoch": stale} {
		if legalIn(g, src, spec, id) {
			t.Errorf("announce accepted the %s exile", name)
		}
	}
	if !legalIn(g, src, spec, mine) {
		t.Error("announce refused the source's own exile")
	}
}

// A walk with no source object admits nothing, never everything: the
// direction a simplification may err (#259).
func TestExiledWithSourceNeedsASource(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	crystal := pushScopedTestCreature(g, me, 0, 0)
	linked := pushExiled(g, opp, sourceRef(g, crystal), false)
	if got := legalIDs(g, SourceChooser(me), exileSpec(true)); len(got) != 0 {
		t.Errorf("no source admits %v, want nothing", got)
	}
	if legalIn(g, SourceChooser(me), exileSpec(true), linked) {
		t.Error("announce accepted a link with no source to compare it to")
	}
}

// The resolution re-check (CR 608.2b): a card that left exile in
// response is not a legal target, and one the source's exile no longer
// names (it moved and came back) is a new object with no link.
func TestExiledWithSourceIsReCheckedAtResolution(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	crystal := pushScopedTestCreature(g, me, 0, 0)
	linked := pushExiled(g, opp, sourceRef(g, crystal), false)
	spec := exileSpec(true)
	src := sourceOf(g, me, crystal)
	if !legalIn(g, src, spec, linked) {
		t.Fatal("setup: the linked card should be a legal target")
	}
	// It leaves exile in response.
	g.WithWriteLock(func() {
		if _, err := g.Exile.Remove(linked); err != nil {
			t.Fatalf("Remove: %v", err)
		}
	})
	if legalIn(g, src, spec, linked) {
		t.Error("a card that left exile is still a legal target")
	}
}

// A source that left the battlefield is read from its last-known
// record (CR 608.2h): the cards it exiled stay "exiled with" it for an
// ability already on the stack, while a new object under the same
// instance is not it (CR 400.7).
func TestExiledWithSourceReadsTheDepartedSource(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	crystal := pushScopedTestCreature(g, me, 0, 0)
	linked := pushExiled(g, opp, sourceRef(g, crystal), false)

	var rec PermanentInfo
	g.WithWriteLock(func() {
		g.rememberDepartingPermanentLocked(crystal)
		c := findBattlefieldCard(g, crystal)
		rec, _ = g.lastKnownPermanentLocked(ObjectRef{ID: crystal, Epoch: c.ObjectEpoch})
	})
	src := SourceSnapshot(me, &rec.Characteristic)
	src.Departed, src.DepartedID = &rec, crystal
	if got := legalIDs(g, src, exileSpec(true)); len(got) != 1 || !got[linked] {
		t.Errorf("departed source offered %v, want only %s", got, linked)
	}
	// A different epoch of the same card is a different object.
	later := src
	rec2 := rec
	rec2.Epoch++
	later.Departed = &rec2
	if got := legalIDs(g, later, exileSpec(true)); len(got) != 0 {
		t.Errorf("a later object of the same card offered %v, want nothing", got)
	}
}

// A face-down card in exile the chooser has not seen is never offered:
// the legal set crosses the wire and the bot's move list, and a
// predicate over a card nobody may look at would leak its type.
func TestExileClauseNeverOffersAHiddenCard(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	crystal := pushScopedTestCreature(g, me, 0, 0)
	hidden := pushExiled(g, opp, PermissionCardRef{}, true)
	seen := pushExiled(g, opp, PermissionCardRef{}, true)
	g.WithWriteLock(func() {
		for i := range g.Exile.Cards {
			if g.Exile.Cards[i].InstanceID == seen {
				g.Exile.Cards[i].AddKnower(me)
			}
		}
	})
	spec := exileSpec(false)
	src := sourceOf(g, me, crystal)
	got := legalIDs(g, src, spec)
	if got[hidden] {
		t.Error("an unseen face-down exiled card was offered")
	}
	if !got[seen] {
		t.Error("a face-down card its chooser has seen should be offerable")
	}
	if legalIn(g, src, spec, hidden) {
		t.Error("announce accepted an unseen face-down exiled card")
	}
}

// The spec is data: the walk gives the same answer on a Clone, which is
// what undo and a restored table rely on.
func TestExiledWithSourceSurvivesClone(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	crystal := pushScopedTestCreature(g, me, 0, 0)
	linked := pushExiled(g, opp, sourceRef(g, crystal), false)
	pushExiled(g, opp, PermissionCardRef{}, false)

	c := g.Clone()
	got := legalIDs(c, sourceOf(c, me, crystal), exileSpec(true))
	if len(got) != 1 || !got[linked] {
		t.Errorf("clone offered %v, want only %s", got, linked)
	}
}
