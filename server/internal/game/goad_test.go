package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// goad_test.go pins #1598 (ADR 0045 Decision 53): a creature remembers
// every player who goaded it (CR 701.15c), each goad ends as its own
// goader's next turn begins (CR 701.15a), and the attack requirements
// read the whole set. The four seats are named from the creature's
// controller: `me` controls X, A and B goad it, and C is the one
// opponent who did not.

type goadTable struct {
	g                  *Game
	me, a, b           *Player
	c                  *Player
	x                  uuid.UUID
	meSeat, aSeat, bSt int
}

func newGoadTable(t *testing.T) goadTable {
	t.Helper()
	g := newFourPlayerActiveGame(t)
	s := g.Turn.ActiveSeat
	tb := goadTable{
		g:      g,
		me:     g.Seats[s],
		a:      g.Seats[(s+1)%4],
		b:      g.Seats[(s+2)%4],
		c:      g.Seats[(s+3)%4],
		meSeat: s,
		aSeat:  (s + 1) % 4,
		bSt:    (s + 2) % 4,
	}
	tb.x = pushKeywordCreature(t, g, tb.me, 2, 2)
	return tb
}

func goadersOf(t *testing.T, g *Game, id uuid.UUID) []uuid.UUID {
	t.Helper()
	c, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Goaders()
}

func sameGoaders(got []uuid.UUID, want ...uuid.UUID) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// weightsAt is how many of X's requirements an attack at each player
// obeys — the number CR 508.1d maximises.
func weightsAt(g *Game, x uuid.UUID, targets ...*Player) []int {
	out := make([]int, len(targets))
	g.WithWriteLock(func() {
		g.recomputeLayersLocked()
		c := findBattlefieldCard(g, x)
		reqs := g.attackRequirementsOfLocked(c)
		for i, p := range targets {
			out[i] = g.requirementWeightLocked(reqs, p.ID)
		}
	})
	return out
}

// TestGoadByTwoPlayersMustAttackTheThird — A goads X, then B goads X.
// The second goad does not replace the first: attacking A is refused
// (naming A's goad), attacking B is refused, attacking C is accepted,
// and X at home is refused at the pass.
func TestGoadByTwoPlayersMustAttackTheThird(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goad(t, g, tb.x, tb.a.ID)
	goad(t, g, tb.x, tb.b.ID)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.a.ID, tb.b.ID) {
		t.Fatalf("goaders = %v, want [A B]", got)
	}
	advanceIntoStep(t, g, StepDeclareAttackers)

	requirementErr(t, g.PassPriority())
	re := requirementErr(t, g.DeclareAttacker(tb.x, tb.a.ID))
	if re.Requirement.OtherThan != tb.a.ID {
		t.Errorf("at A: refusal names %+v, want A's other-than requirement", re.Requirement)
	}
	re = requirementErr(t, g.DeclareAttacker(tb.x, tb.b.ID))
	if re.Requirement.OtherThan != tb.b.ID {
		t.Errorf("at B: refusal names %+v, want B's other-than requirement", re.Requirement)
	}
	if got := re.Sentence(tb.b.ID); got != "Kw Creature is goaded by you and must attack a player other than you if able." {
		t.Errorf("sentence to B = %q", got)
	}
	if err := g.DeclareAttacker(tb.x, tb.c.ID); err != nil {
		t.Fatalf("X at C, the player who goaded it neither time: %v", err)
	}
	requirementErr(t, g.DeclareAttacker(tb.x, tb.a.ID))
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both goads obeyed: %v", err)
	}
}

// TestGoadEndsPerGoaderAtThatGoadersTurn — CR 701.15a per goader. Once
// A's next turn begins, A's goad is over and B's is not: an attack on A
// then obeys as much as one on C, and one on B still obeys less. When
// B's turn begins, B's ends too. A goad B makes on that turn lasts to
// B's NEXT turn, so at X's controller's next declaration X may attack
// A and may not attack B.
func TestGoadEndsPerGoaderAtThatGoadersTurn(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goad(t, g, tb.x, tb.a.ID)
	goad(t, g, tb.x, tb.b.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(tb.x, tb.c.ID); err != nil {
		t.Fatalf("X at C: %v", err)
	}

	advanceToStepOfSeat(t, g, tb.aSeat, StepUpkeep)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.b.ID) {
		t.Fatalf("after A's turn began, goaders = %v, want [B]", got)
	}
	w := weightsAt(g, tb.x, tb.a, tb.b, tb.c)
	if w[0] != w[2] || w[1] >= w[0] {
		t.Errorf("weights at A, B, C = %v: want A == C > B once only B's goad is left", w)
	}

	advanceToStepOfSeat(t, g, tb.bSt, StepUpkeep)
	if got := goadersOf(t, g, tb.x); len(got) != 0 {
		t.Fatalf("after B's turn began, goaders = %v, want none", got)
	}
	goad(t, g, tb.x, tb.b.ID)

	advanceToStepOfSeat(t, g, tb.meSeat, StepDeclareAttackers)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.b.ID) {
		t.Fatalf("B's goad from B's own turn ended early: goaders = %v", got)
	}
	requirementErr(t, g.DeclareAttacker(tb.x, tb.b.ID))
	if err := g.DeclareAttacker(tb.x, tb.a.ID); err != nil {
		t.Fatalf("X at A once A's goad is over: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
}

// TestGoadedByEveryOpponentMustAttackAnyOfThem — CR 701.15c's fallback:
// when every opponent has goaded X no player satisfies every
// "other than", so X must still attack, and any opponent is as good as
// any other.
func TestGoadedByEveryOpponentMustAttackAnyOfThem(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	for _, p := range []*Player{tb.a, tb.b, tb.c} {
		goad(t, g, tb.x, p.ID)
	}
	advanceIntoStep(t, g, StepDeclareAttackers)
	requirementErr(t, g.PassPriority())
	for _, p := range []*Player{tb.a, tb.b, tb.c} {
		if err := g.DeclareAttacker(tb.x, p.ID); err != nil {
			t.Fatalf("X at a goader when every opponent goaded it: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with X attacking: %v", err)
	}
}

// TestRegoadRefreshesTheGoadersEntry — the same player goading again is
// one goad, not two: the entry moves to the end (the latest goad), its
// end is restamped to that player's next turn, and it still ends then.
func TestRegoadRefreshesTheGoadersEntry(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goad(t, g, tb.x, tb.a.ID)
	goad(t, g, tb.x, tb.b.ID)
	// A stale stamp, as a goad made long ago would carry.
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(tb.x)
		c.Goads[0].ExpiresAtTurnsBegun = -1
	})
	goad(t, g, tb.x, tb.a.ID)

	c, _ := battlefieldCardByID(g, tb.x)
	if !sameGoaders(c.Goaders(), tb.b.ID, tb.a.ID) {
		t.Fatalf("goaders after A re-goads = %v, want [B A]", c.Goaders())
	}
	if c.LatestGoader() != tb.a.ID {
		t.Errorf("latest goader = %v, want A", c.LatestGoader())
	}
	if want := tb.a.TurnsBegun + 1; c.Goads[1].ExpiresAtTurnsBegun != want {
		t.Errorf("A's refreshed goad ends at %d, want %d (A's next turn)", c.Goads[1].ExpiresAtTurnsBegun, want)
	}

	// The refreshed goad survives a turn start that is not A's…
	g.WithWriteLock(func() { g.sweepExpiredGoadsLocked() })
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.b.ID, tb.a.ID) {
		t.Fatalf("a sweep on nobody's new turn changed the goaders to %v", got)
	}
	// …and ends when A's does.
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(tb.x, tb.c.ID); err != nil {
		t.Fatalf("X at C: %v", err)
	}
	advanceToStepOfSeat(t, g, tb.aSeat, StepUpkeep)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.b.ID) {
		t.Fatalf("after A's turn began, goaders = %v, want [B]", got)
	}
}

// TestGoadsSurviveCloneAndRestoreFrom — undo. The clone owns its goads:
// an edit to the live card's entries does not reach it, and restoring
// it brings every goader back.
func TestGoadsSurviveCloneAndRestoreFrom(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goad(t, g, tb.x, tb.a.ID)
	goad(t, g, tb.x, tb.b.ID)
	saved := g.Clone()

	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(tb.x)
		c.Goads[0].By = tb.c.ID
	})
	if got := goadersOf(t, saved, tb.x); !sameGoaders(got, tb.a.ID, tb.b.ID) {
		t.Fatalf("an edit to the live goads reached the clone: %v", got)
	}
	goad(t, g, tb.x, uuid.Nil)
	g.RestoreFrom(saved)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.a.ID, tb.b.ID) {
		t.Fatalf("goaders after undo = %v, want [A B]", got)
	}
}

// TestGoadsRoundTripTheSnapshot — every goader and each goad's end come
// back from a restore point, and the restored game still enforces them.
// The legacy key carries the latest goader for an older binary.
func TestGoadsRoundTripTheSnapshot(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goad(t, g, tb.x, tb.a.ID)
	goad(t, g, tb.x, tb.b.ID)
	before, _ := battlefieldCardByID(g, tb.x)

	snap, restored := roundTrip(t, g)
	for _, cs := range snap.Battlefield.Cards {
		if cs.InstanceID == tb.x && cs.GoadedBy != tb.b.ID {
			t.Errorf("legacy goadedBy = %v, want the latest goader B", cs.GoadedBy)
		}
	}
	after, ok := battlefieldCardByID(restored, tb.x)
	if !ok {
		t.Fatal("X lost in the round trip")
	}
	if len(after.Goads) != 2 || after.Goads[0] != before.Goads[0] || after.Goads[1] != before.Goads[1] {
		t.Fatalf("goads after round trip = %+v, want %+v", after.Goads, before.Goads)
	}
	advanceIntoStep(t, restored, StepDeclareAttackers)
	requirementErr(t, restored.DeclareAttacker(tb.x, tb.a.ID))
	requirementErr(t, restored.DeclareAttacker(tb.x, tb.b.ID))
	if err := restored.DeclareAttacker(tb.x, tb.c.ID); err != nil {
		t.Fatalf("restored X at C: %v", err)
	}
}

// TestLegacySingleGoaderSnapshotRestores — a restore point written
// before #1598 has `goadedBy` and no `goads`. It comes back as a
// one-goader set stamped to end at that goader's next turn, and ends
// there.
func TestLegacySingleGoaderSnapshotRestores(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goad(t, g, tb.x, tb.a.ID)
	snap := g.CaptureSnapshot()
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	// Strip the new key the way a pre-#1598 binary never wrote it.
	var tree map[string]any
	if err := json.Unmarshal(raw, &tree); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range tree["battlefield"].(map[string]any)["cards"].([]any) {
		card := c.(map[string]any)
		if _, ok := card["goads"]; ok {
			delete(card, "goads")
			found = card["goadedBy"] == tb.a.ID.String()
		}
	}
	if !found {
		t.Fatal("the capture had no goads key beside goadedBy = A to strip")
	}
	raw, err = json.Marshal(tree)
	if err != nil {
		t.Fatal(err)
	}
	var legacy GameSnapshot
	if err := json.Unmarshal(raw, &legacy); err != nil {
		t.Fatal(err)
	}
	restored, err := legacy.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	c, _ := battlefieldCardByID(restored, tb.x)
	if len(c.Goads) != 1 || c.Goads[0].By != tb.a.ID {
		t.Fatalf("legacy goad restored as %+v, want one goad by A", c.Goads)
	}
	if want := restored.Seats[tb.aSeat].TurnsBegun + 1; c.Goads[0].ExpiresAtTurnsBegun != want {
		t.Errorf("legacy goad ends at %d, want %d (A's next turn)", c.Goads[0].ExpiresAtTurnsBegun, want)
	}
	advanceIntoStep(t, restored, StepDeclareAttackers)
	requirementErr(t, restored.DeclareAttacker(tb.x, tb.a.ID))
	if err := restored.DeclareAttacker(tb.x, tb.b.ID); err != nil {
		t.Fatalf("restored X at B: %v", err)
	}
	advanceToStepOfSeat(t, restored, tb.aSeat, StepUpkeep)
	if got := goadersOf(t, restored, tb.x); len(got) != 0 {
		t.Fatalf("legacy goad outlived A's next turn: %v", got)
	}
}

// TestSetGoadedNilClearsEveryGoad — the sandbox's "clear goad".
func TestSetGoadedNilClearsEveryGoad(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goad(t, g, tb.x, tb.a.ID)
	goad(t, g, tb.x, tb.b.ID)
	goad(t, g, tb.x, uuid.Nil)
	if got := goadersOf(t, g, tb.x); len(got) != 0 {
		t.Fatalf("goaders after clear = %v", got)
	}
	if err := g.SetGoaded(uuid.New(), tb.a.ID); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("goading a missing card = %v, want ErrCardNotFound", err)
	}
}
