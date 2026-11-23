package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// targeting_bar_table_test.go pins #5718's two engine halves: the "this
// turn and next turn" duration, which lasts through exactly the turn it
// was made in and the turn after, whoever takes it (an extra turn
// included), and the table-wide bar on spells and activated abilities,
// which a triggered ability passes. Card-level cases are in
// cards/effects/peace_talks_test.go.

func thisAndNextLive(g *Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			if e.Duration.Kind == ThisTurnAndNextTurn {
				n++
			}
		}
	})
	return n
}

func registerThisAndNext(t *testing.T, g *Game) Duration {
	t.Helper()
	var d Duration
	g.WithWriteLock(func() {
		d = g.ThisTurnAndNextTurnDuration()
		if !g.PlayersAndPermanentsCantBeTargetedForEffect(uuid.New(), g.activePlayerIDLocked(), d, "this turn and next turn") {
			t.Fatal("the bar was not registered")
		}
	})
	if d.Kind != ThisTurnAndNextTurn || !d.Known() || d.AfterTurn != g.Turn.Seq {
		t.Fatalf("duration = %+v, want a known this-turn-and-next-turn window stamped %d", d, g.Turn.Seq)
	}
	return d
}

// Made on seat 0's turn, it lasts through seat 0's cleanup and through
// seat 1's turn, and is gone as seat 2's begins: two turns, not a round.
func TestThisTurnAndNextTurnLastsExactlyTwoTurns(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	registerThisAndNext(t, g)
	advanceToStepOfSeat(t, g, 0, StepCleanup)
	if thisAndNextLive(g) != 1 {
		t.Fatal("the effect ended at the cleanup of the turn it was made in")
	}
	advanceToStepOfSeat(t, g, 1, StepEnd)
	if thisAndNextLive(g) != 1 {
		t.Fatal("the effect ended before the next turn's cleanup")
	}
	advanceToStepOfSeat(t, g, 1, StepCleanup)
	if thisAndNextLive(g) != 0 {
		t.Fatal("the effect outlived the next turn's cleanup step")
	}
	advanceToStepOfSeat(t, g, 2, StepUpkeep)
	if thisAndNextLive(g) != 0 {
		t.Error("the effect came back on the third turn")
	}
}

// An extra turn taken right after this one is the next turn (CR 500.7;
// the card's ruling: it "could affect the same player twice"), so the
// effect is over at its cleanup and the following seat's turn is free.
func TestThisTurnAndNextTurnCountsAnExtraTurnAsTheNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	registerThisAndNext(t, g)
	g.WithWriteLock(func() { g.TakeExtraTurnsForEffect(me, uuid.New(), 1) })
	start := g.Turn.Seq
	for i := 0; i < 40 && g.Turn.Seq == start; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != 0 || !g.Turn.Extra {
		t.Fatalf("active seat %d extra %v, want seat 0's extra turn", g.Turn.ActiveSeat, g.Turn.Extra)
	}
	if thisAndNextLive(g) != 1 {
		t.Fatal("the effect ended before the extra turn")
	}
	advanceToStepOfSeat(t, g, 0, StepEnd)
	if thisAndNextLive(g) != 1 {
		t.Fatal("the effect ended before the extra turn's cleanup")
	}
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if thisAndNextLive(g) != 0 {
		t.Error("the effect outlived the extra turn, which was its next turn")
	}
}

// The bar refuses a spell's and an activated ability's walk at a player
// and at a permanent, and a triggered ability's passes. It is gone with
// its duration.
func TestTableTargetingBarRefusesSpellsAndActivatedAbilitiesOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := pushScopedTestCreature(g, them.ID, 2, 2)
	registerThisAndNext(t, g)
	var card Card
	g.ReadSnapshot(func() {
		c, _ := g.battlefieldCardLocked(bear)
		card = *c
	})
	source := Card{InstanceID: uuid.New(), Name: "Probe", Controller: me.ID, Owner: me.ID}
	cases := []struct {
		name string
		src  TargetSource
		want bool
	}{
		{"a spell", SourceCast(me.ID, &source), false},
		{"an activated ability", SourceObject(me.ID, &source), false},
		{"a triggered ability", SourceObject(me.ID, &source).WithTrigger(TriggerContext{}), true},
	}
	for _, tc := range cases {
		g.ReadSnapshot(func() {
			if got := g.canPlayerBeTargetedByLocked(them, tc.src); got != tc.want {
				t.Errorf("%s targeting a player: %v, want %v", tc.name, got, tc.want)
			}
			if got := g.canPlayerBeTargetedByLocked(me, tc.src); got != tc.want {
				t.Errorf("%s targeting its own controller: %v, want %v", tc.name, got, tc.want)
			}
			if got := g.canBeTargetedByLocked(&card, ZoneBattlefield, tc.src); got != tc.want {
				t.Errorf("%s targeting a permanent: %v, want %v", tc.name, got, tc.want)
			}
			if !g.canBeTargetedByLocked(&card, ZoneGraveyard, tc.src) {
				t.Errorf("%s: a card outside the battlefield is not a permanent and is not barred", tc.name)
			}
		})
	}
	advanceToStepOfSeat(t, g, 2, StepUpkeep)
	g.ReadSnapshot(func() {
		if !g.canPlayerBeTargetedByLocked(them, SourceCast(me.ID, &source)) {
			t.Error("the bar outlived its duration")
		}
	})
}

// A triggered item on the stack is re-checked as a triggered ability's
// walk (CR 608.2b), and an activated one as an activation's.
func TestStackItemSourceKnowsATriggeredAbility(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	g.ReadSnapshot(func() {
		if !g.stackItemSourceLocked(&StackItem{Kind: StackItemTriggered, Controller: me}).IsTriggered() {
			t.Error("a triggered item's source is not a triggered ability's")
		}
		for _, k := range []StackItemKind{StackItemActivated, StackItemSpell} {
			if g.stackItemSourceLocked(&StackItem{Kind: k, Controller: me}).IsTriggered() {
				t.Errorf("a %s item's source reads as a triggered ability's", k)
			}
		}
	})
}

// A restore point carries the record and its stamp, and the restored
// game ends it at the same cleanup.
func TestTableTargetingBarSurvivesARestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	registerThisAndNext(t, g)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if thisAndNextLive(r) != 1 {
		t.Fatal("the restore dropped the record")
	}
	r.ReadSnapshot(func() {
		if r.canPlayerBeTargetedByLocked(r.Seats[2], SourceChooser(r.Seats[0].ID)) {
			t.Error("the restored record bars nothing")
		}
	})
	advanceToStepOfSeat(t, r, 2, StepUpkeep)
	if thisAndNextLive(r) != 0 {
		t.Error("the restored record outlived the next turn")
	}
}

// The kind keeps its persisted int, names itself, and refuses a stamp
// on another kind or a record in the wrong shape.
func TestThisTurnAndNextTurnKeepsItsPersistedIntAndShape(t *testing.T) {
	if !ThisTurnAndNextTurn.Known() || ThisTurnAndNextTurn != UntilYourNextUpkeep+1 || (ThisTurnAndNextTurn + 1).Known() {
		t.Error("ThisTurnAndNextTurn is not the last known duration kind, right after UntilYourNextUpkeep")
	}
	if int(ThisTurnAndNextTurn) != 13 {
		t.Errorf("ThisTurnAndNextTurn persists as %d, want 13", ThisTurnAndNextTurn)
	}
	if got := ThisTurnAndNextTurn.String(); got != "this turn and next turn" {
		t.Errorf("String() = %q", got)
	}
	if (Duration{Kind: ThisTurnAndNextTurn, AfterTurn: -1}).Known() {
		t.Error("a this-turn-and-next-turn duration with a negative stamp is known")
	}
	if !(Duration{Kind: ThisTurnAndNextTurn, AfterTurn: 3}).Known() {
		t.Error("a sound this-turn-and-next-turn duration is unknown")
	}
	mod := Mod{Kind: ModCantBeTargetedBySpellsOrActivated}
	if p := tableTargetingBarModProblem(ScopeGame, mod); p != "" {
		t.Errorf("a sound record has a problem: %s", p)
	}
	if tableTargetingBarModProblem(ScopeCreatures, mod) == "" {
		t.Error("the bar on a creature scope is not refused")
	}
	if tableTargetingBarModProblem(ScopeGame, Mod{Kind: ModCantBeTargetedBySpellsOrActivated, Player: uuid.New()}) == "" {
		t.Error("the bar naming a player is not refused")
	}
}
