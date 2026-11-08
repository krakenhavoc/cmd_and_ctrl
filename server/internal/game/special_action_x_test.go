package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// special_action_x_test.go — #5249: a turn-face-up cost with an {X}
// (CR 107.3d). X is chosen as the special action is taken, paid as {X}
// generic mana, refused when it cannot be paid, and handed to the
// permanent's other abilities on EventTurnedFaceUp (CR 702.168e).

// disguiseXCost is Aurelia's Vindicator's shape on the morphGame card:
// "Disguise {X}{1}{U}".
const disguiseXCost = "{X}{1}{U}"

// disguiseXGame is morphGame with a disguise {X}{1}{U} declaration, the
// card cast face down and resolved. The cast spends {3} of the eight
// blue in the pool, so five are left: enough for X = 3 and no more.
func disguiseXGame(t *testing.T) (*Game, *Player, uuid.UUID) {
	t.Helper()
	offer := morphOffer(FaceDownDisguised, disguiseXCost, false)
	offer.Key = "disguise"
	g, me, id := morphGame(t, offer)
	castAndLandFaceDown(t, g, me, id, "disguise")
	if got := me.ManaPool.Total(); got != 5 {
		t.Fatalf("fixture: %d mana left after the face-down cast, want 5", got)
	}
	return g, me, id
}

// turnedFaceUpAmount is the Amount of the EventTurnedFaceUp for id, and
// whether there was one.
func turnedFaceUpAmount(g *Game, id uuid.UUID) (int, bool) {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == EventTurnedFaceUp && g.Events[i].CardID == id {
			return g.Events[i].Amount, true
		}
	}
	return 0, false
}

// TestTurnFaceUpPaysTheChosenX is the headline: X = 2 charges {2}{1}{U},
// turns the permanent up, and the event carries the 2 the player chose.
func TestTurnFaceUpPaysTheChosenX(t *testing.T) {
	g, me, id := disguiseXGame(t)
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("turn face up for X=2: %v", err)
	}
	if c, _, _ := cardAnywhere(g, id); c.FaceDown {
		t.Fatal("still face down after the special action")
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("%d mana left, want 1: X=2 costs {2}{1}{U}, four of the five", got)
	}
	if x, ok := turnedFaceUpAmount(g, id); !ok || x != 2 {
		t.Errorf("EventTurnedFaceUp Amount = %d (found %v), want the chosen 2 (CR 702.168e)", x, ok)
	}
	var label string
	for _, ev := range g.Events {
		if ev.Kind == EventSpecialAction && ev.CardID == id {
			label = ev.Label
		}
	}
	if label != "Turn face up {X}{1}{U} for X=2" {
		t.Errorf("log label = %q, want the cost and the X paid", label)
	}
}

// TestTurnFaceUpRefusesAnXItCannotPay: five mana pay X = 3 and not
// X = 4. A strict payment refuses before anything is spent, and the
// permanent stays face down.
func TestTurnFaceUpRefusesAnXItCannotPay(t *testing.T) {
	g, me, id := disguiseXGame(t)
	err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, XValue: 4})
	if !errors.Is(err, ErrInsufficientMana) {
		t.Fatalf("turn face up for X=4 with five mana = %v, want ErrInsufficientMana", err)
	}
	if c, _, _ := cardAnywhere(g, id); !c.FaceDownIsPermanent() {
		t.Fatal("an unpaid turn face up turned the permanent up")
	}
	if got := me.ManaPool.Total(); got != 5 {
		t.Errorf("the refused action spent mana: %d left, want 5", got)
	}
	if _, ok := turnedFaceUpAmount(g, id); ok {
		t.Error("a refused action emitted EventTurnedFaceUp")
	}
	// The largest X it can pay is still taken.
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("turn face up for X=3 with five mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left after X=3, want 0", got)
	}
}

// TestTurnFaceUpAtXZero: X = 0 is a legal choice (CR 107.3d puts no
// floor on it) and costs the rest of the cost alone.
func TestTurnFaceUpAtXZero(t *testing.T) {
	g, me, id := disguiseXGame(t)
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up for X=0: %v", err)
	}
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("%d mana left, want 3: X=0 costs {1}{U}", got)
	}
	if x, ok := turnedFaceUpAmount(g, id); !ok || x != 0 {
		t.Errorf("EventTurnedFaceUp Amount = %d (found %v), want 0", x, ok)
	}
}

// TestTurnFaceUpRefusesAnXTheCostDoesNotHave: an X for a cost with no
// {X} is a confused client, refused rather than ignored (as an
// activation's is), before anything is paid. So is a negative X, an X
// past the announcement bound, and an X on any other kind.
func TestTurnFaceUpRefusesAnXTheCostDoesNotHave(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	castAndLandFaceDown(t, g, me, id, "morph")
	for _, x := range []int{1, -1, MaxAnnouncedX + 1} {
		err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, XValue: x})
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("X=%d on morph {1}{U} = %v, want ErrInvalidParam", x, err)
		}
	}
	if c, _, _ := cardAnywhere(g, id); !c.FaceDownIsPermanent() {
		t.Fatal("a refused action turned the permanent up")
	}
	if err := g.PerformSpecialAction(me.ID, uuid.New(), SpecialActionForetell, SpecialActionParams{XValue: 1}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("X on a foretell = %v, want ErrInvalidParam", err)
	}
}

// TestTurnFaceUpForEffectCarriesNoX: an effect pays no cost, so nothing
// chose an X and the event's Amount is 0.
func TestTurnFaceUpForEffectCarriesNoX(t *testing.T) {
	g, me, id := disguiseXGame(t)
	var turned bool
	g.WithWriteLock(func() { turned = g.TurnFaceUpForEffect(uuid.Nil, me.ID, id) })
	if !turned {
		t.Fatal("the effect did not turn the permanent up")
	}
	if x, ok := turnedFaceUpAmount(g, id); !ok || x != 0 {
		t.Errorf("EventTurnedFaceUp Amount = %d (found %v), want 0", x, ok)
	}
}

// TestSpecialActionXCeilingIsTheLargestPayableX is the view's ceiling:
// five mana in the pool pay X = 3, and a cost with no {X} has none.
func TestSpecialActionXCeilingIsTheLargestPayableX(t *testing.T) {
	g, me, id := disguiseXGame(t)
	c, _, _ := cardAnywhere(g, id)
	sa := TurnFaceUpOffer(c)
	if sa == nil || !SpecialActionDemandsX(*sa) {
		t.Fatalf("offer %+v does not demand an X", sa)
	}
	var x int
	var ok bool
	g.WithWriteLock(func() { x, ok = g.SpecialActionXCeilingForEffect(me.ID, c, *sa) })
	if !ok || x != 3 {
		t.Errorf("ceiling = %d (ok %v), want 3", x, ok)
	}
	if SpecialActionDemandsX(SpecialAction{Kind: SpecialActionTurnFaceUp, Cost: morphCost}) {
		t.Error("{1}{U} demands an X")
	}
	if SpecialActionDemandsX(SpecialAction{Kind: SpecialActionSuspend, Cost: "{X}{R}"}) {
		t.Error("a suspend cost demands an X: only turn_face_up takes one")
	}
}
