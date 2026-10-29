package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// move_reflexive_test.go — #4517: a CR 603.12 reflexive trigger whose
// creator is a replacement that applied to a move ("exile it instead.
// When you do, …"). The trigger is declared on the move and created only
// once the card has landed where the replacement sent it.

// moveReflexiveProbe is a registered reflexive body the tests name.
var moveReflexiveProbe = ReflexiveBody(testEffectKeyPrefix+"move-reflexive-probe",
	func(*Game, *StackItem, EffectParams) error { return nil }, nil)

const moveReflexiveLabel = "probe — when you do"

// exileInsteadWhenYouDo is "if a creature controlled by someone other
// than the source's controller would die, exile it instead. When you do,
// …", with the source a value the closure holds.
func exileInsteadWhenYouDo(src Card) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
			if ev.Kind != RepEventMove || ev.OldZone != ZoneBattlefield || ev.NewZone != ZoneGraveyard {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.IsCreature() && c.Controller != src.Controller
		},
		Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
			ev.NewZone = ZoneExile
			ev.NewZoneOwner = uuid.Nil
			s := src
			return g.QueueReflexiveTriggerAfterReplacedMoveForEffect(ev, &s,
				ReflexiveTrigger{Label: moveReflexiveLabel, Body: moveReflexiveProbe})
		},
		Label: "exile instead. When you do, …",
	}
}

// afterExile is a second replacement on the same card's move to exile:
// redirect sends it to its owner's hand instead, otherwise it cancels
// the move.
func afterExile(card uuid.UUID, redirect bool) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == card && ev.NewZone == ZoneExile
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			if redirect {
				ev.NewZone = ZoneHand
				return nil
			}
			ev.Cancel()
			return nil
		},
		Label: "and then something else",
	}
}

func probeTriggers(g *Game) []*StackItem {
	var out []*StackItem
	for _, it := range g.PendingTriggers {
		if it != nil && it.Label == moveReflexiveLabel {
			out = append(out, it)
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Label == moveReflexiveLabel {
			out = append(out, it)
		}
	}
	return out
}

func reflexiveMoveBoard(t *testing.T) (*Game, Card, uuid.UUID) {
	t.Helper()
	g := newActiveGameWithSeats(t, 3)
	src := Card{InstanceID: uuid.New(), Name: "Probe Hunter", Owner: g.Seats[0].ID, Controller: g.Seats[0].ID, ObjectEpoch: 3}
	bear := pushBear(g, g.Seats[1].ID)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(exileInsteadWhenYouDo(src)) })
	return g, src, bear
}

// The trigger is created once the card is in exile, under the
// replacement's controller and naming its source object, and waits on
// PendingTriggers for the next priority (CR 603.3).
func TestAReflexiveTriggerFromAReplacedMoveIsCreatedOnLanding(t *testing.T) {
	g, src, bear := reflexiveMoveBoard(t)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Exile.Contains(bear) {
		t.Fatal("the bear is not in exile")
	}
	got := probeTriggers(g)
	if len(got) != 1 {
		t.Fatalf("%d reflexive triggers, want 1", len(got))
	}
	it := got[0]
	if it.Controller != src.Controller || it.SourceCardID != src.InstanceID {
		t.Errorf("trigger controller %v source %v, want %v and %v", it.Controller, it.SourceCardID, src.Controller, src.InstanceID)
	}
	if it.SourceObject != (ObjectRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}) {
		t.Errorf("trigger source object %+v, want the replacement's source as it applied", it.SourceObject)
	}
}

// One trigger per replaced move (CR 603.12a): three creatures destroyed
// together.
func TestAReflexiveTriggerFromAReplacedMoveOncePerMove(t *testing.T) {
	g, _, bear := reflexiveMoveBoard(t)
	ids := []uuid.UUID{bear, pushBear(g, g.Seats[1].ID), pushBear(g, g.Seats[2].ID)}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentsKeyedForEffect(ids, KeyedTail{}); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if got := len(probeTriggers(g)); got != 3 {
		t.Fatalf("%d reflexive triggers, want 3", got)
	}
}

// "When you do" is about this replacement's action. A later replacement
// that sends the card somewhere else, or cancels the move, means it was
// never exiled this way: no trigger.
func TestAReflexiveTriggerFromAReplacedMoveNeedsTheCardToLandThere(t *testing.T) {
	for _, redirect := range []bool{true, false} {
		g, _, bear := reflexiveMoveBoard(t)
		g.WithWriteLock(func() {
			g.RegisterReplacementForTest(afterExile(bear, redirect))
			_ = g.DestroyPermanentForEffect(bear)
		})
		if g.Exile.Contains(bear) {
			t.Fatalf("redirect=%v: the bear reached exile", redirect)
		}
		if got := len(probeTriggers(g)); got != 0 {
			t.Errorf("redirect=%v: %d reflexive triggers, want none", redirect, got)
		}
	}
}

// Its own creature's death is not replaced, so nothing is declared.
func TestAReflexiveTriggerFromAReplacedMoveOnlyWhenItApplies(t *testing.T) {
	g, src, _ := reflexiveMoveBoard(t)
	mine := pushBear(g, src.Controller)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	if got := len(probeTriggers(g)); got != 0 {
		t.Fatalf("%d reflexive triggers for an unreplaced death", got)
	}
}

func TestQueueReflexiveTriggerAfterReplacedMoveRefusesMalformed(t *testing.T) {
	g := NewGame()
	src := &Card{InstanceID: uuid.New()}
	ok := ReflexiveTrigger{Label: "x", Body: moveReflexiveProbe}
	for name, tc := range map[string]struct {
		ev  *ReplacementEvent
		src *Card
		rt  ReflexiveTrigger
	}{
		"nil event":    {nil, src, ok},
		"not a move":   {&ReplacementEvent{Kind: RepEventDraw}, src, ok},
		"no source":    {&ReplacementEvent{Kind: RepEventMove}, nil, ok},
		"no label":     {&ReplacementEvent{Kind: RepEventMove}, src, ReflexiveTrigger{Body: moveReflexiveProbe}},
		"no body":      {&ReplacementEvent{Kind: RepEventMove}, src, ReflexiveTrigger{Label: "x"}},
		"optional":     {&ReplacementEvent{Kind: RepEventMove}, src, ReflexiveTrigger{Label: "x", Body: moveReflexiveProbe, Optional: &TriggerOptionalPrompt{}}},
		"with payload": {&ReplacementEvent{Kind: RepEventMove}, src, ReflexiveTrigger{Label: "x", Body: moveReflexiveProbe, Payload: []TargetRef{{}}}},
	} {
		var err error
		g.WithWriteLock(func() { err = g.QueueReflexiveTriggerAfterReplacedMoveForEffect(tc.ev, tc.src, tc.rt) })
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: err = %v, want ErrInvalidParam", name, err)
		}
	}
}

// A move paused on a CR 616 question with a reflexive trigger declared
// is a restore point under its own key, and the restored answer creates
// the trigger exactly when the live one would: the card lands in the
// declared zone (the hand) on a "no", and not on a "yes" (the command
// zone).
func TestAPausedMoveWithAReflexiveTriggerIsARestorePoint(t *testing.T) {
	for _, apply := range []bool{true, false} {
		g, cmdID := bouncedCommanderTable(t)
		owner := g.Seats[0]
		g.WithWriteLock(func() {
			f := g.PendingChoices[0].replacementResume
			f.ev.reflexiveAfterMove = []reflexiveMoveInstruction{{
				Source: cmdID, SourceObject: ObjectRef{ID: cmdID}, Controller: owner.ID,
				Zone: ZoneHand, Label: moveReflexiveLabel, Body: moveReflexiveProbe.Key(),
			}}
		})
		snap := throughJSON(t, g.CaptureSnapshot())
		if !snap.Restorable() {
			t.Fatalf("not a restore point: %+v", snap.Continuations)
		}
		if got := snap.PendingChoices[0].PickThen; got != ReplacementResumeMoveReflexive {
			t.Fatalf("frame written under %q, want %q", got, ReplacementResumeMoveReflexive)
		}
		restored, err := snap.RestoreStrict()
		if err != nil {
			t.Fatalf("RestoreStrict: %v", err)
		}
		rowner := restored.Seats[0]
		prompt := expectCommanderPrompt(t, restored, rowner)
		if err := restored.ResolveOptionalReplacement(prompt.ID, rowner.ID, apply); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
		want := 1
		if apply {
			want = 0
			assertOnlyIn(t, cmdID, rowner.Command, rowner.Hand, restored.Battlefield)
		} else {
			assertOnlyIn(t, cmdID, rowner.Hand, rowner.Command, restored.Battlefield)
		}
		got := probeTriggers(restored)
		if len(got) != want {
			t.Fatalf("apply=%v: %d reflexive triggers, want %d", apply, len(got), want)
		}
		if want == 1 && got[0].Controller != rowner.ID {
			t.Errorf("the restored trigger is controlled by %v, want %v", got[0].Controller, rowner.ID)
		}
	}
}
