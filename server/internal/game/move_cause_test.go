package game

import (
	"testing"

	"github.com/google/uuid"
)

// move_cause_test.go — #1320: a zone-change event says what moved the
// card. The effect case is inferred from the resolving stack item; the
// cost, special-action, rule and manual cases are named by the route,
// and naming them is what keeps a stale resolving slot from claiming a
// move it did not make.

// zoneMoveFor returns the last EventZoneMove for cardID.
func zoneMoveFor(t *testing.T, g *Game, cardID uuid.UUID) Event {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if ev := g.Events[i]; ev.Kind == EventZoneMove && ev.CardID == cardID {
			return ev
		}
	}
	t.Fatalf("no EventZoneMove for %s", cardID)
	return Event{}
}

// resolvingAbilityOf parks an ability item controlled by `controller`
// in the resolving slot, the state every effect runs in.
func resolvingAbilityOf(g *Game, controller uuid.UUID) *StackItem {
	item := &StackItem{
		ID: uuid.New(), Kind: StackItemTriggered,
		Controller: controller, Owner: controller, SourceCardID: uuid.New(),
	}
	g.beginResolvingLocked(item)
	return item
}

// TestExileByAResolvingAbilityRecordsItsController is the Ranar read:
// an ability I control exiles an opponent's permanent, and the event
// names me and the item — not the permanent's controller, and not
// nobody.
func TestExileByAResolvingAbilityRecordsItsController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushBear(g, opp.ID)
	g.mu.Lock()
	item := resolvingAbilityOf(g, me.ID)
	err := g.ExileCardForEffect(victim)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	ev := zoneMoveFor(t, g, victim)
	if ev.Cause != MoveCauseEffect || ev.CauseController != me.ID || ev.CauseItem != item.ID {
		t.Errorf("cause = %q / %s / %s, want effect / me / the item", ev.Cause, ev.CauseController, ev.CauseItem)
	}
	if !ExiledBySpellOrAbilityOf(ev, me.ID) {
		t.Error("ExiledBySpellOrAbilityOf(me) = false")
	}
	if ExiledBySpellOrAbilityOf(ev, opp.ID) {
		t.Error("the permanent's controller did not exile it")
	}
	// The LTB beside it carries the same cause, for a leaves-the-
	// battlefield reader.
	var ltb Event
	for _, e := range g.Events {
		if e.Kind == EventLTB && e.CardID == victim {
			ltb = e
		}
	}
	if ltb.Cause != MoveCauseEffect || ltb.CauseController != me.ID {
		t.Errorf("LTB cause = %q / %s, want effect / me", ltb.Cause, ltb.CauseController)
	}
}

// TestExileWithNothingResolvingRecordsNoCause: with no item resolving
// there is nothing to infer, and the event says so rather than guessing.
func TestExileWithNothingResolvingRecordsNoCause(t *testing.T) {
	g := newActiveGame(t)
	victim := pushBear(g, g.Seats[1].ID)
	g.mu.Lock()
	err := g.ExileCardForEffect(victim)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	if ev := zoneMoveFor(t, g, victim); ev.Cause != "" || ev.CauseController != uuid.Nil {
		t.Errorf("cause = %q / %s, want none", ev.Cause, ev.CauseController)
	}
}

// TestCostExileIsNotTheLastResolutionsDoing is the stale-slot case the
// explicit kinds exist for. Game.resolving outlives its resolution until
// play moves on, so a cost paid after a spell resolved — in the same
// priority window — would be attributed to that spell if the route
// inferred. The cost route names itself.
func TestCostExileIsNotTheLastResolutionsDoing(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushBear(g, me.ID)
	g.mu.Lock()
	resolvingAbilityOf(g, opp.ID) // an opponent's item resolved a moment ago
	err := g.payExileSelfCostLocked(me.ID, mine, true)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("payExileSelfCostLocked: %v", err)
	}
	ev := zoneMoveFor(t, g, mine)
	if ev.Cause != MoveCauseCost || ev.CauseController != me.ID {
		t.Errorf("cause = %q / %s, want cost / me", ev.Cause, ev.CauseController)
	}
	if ExiledBySpellOrAbilityOf(ev, opp.ID) {
		t.Error("the stale slot claimed a cost exile for the opponent's resolved item")
	}
}

// TestSandboxMoveIsManual: a card dragged into exile by hand is nobody's
// spell or ability.
func TestSandboxMoveIsManual(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mine := pushBear(g, me.ID)
	g.mu.Lock()
	resolvingAbilityOf(g, me.ID)
	g.mu.Unlock()
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneExile}, mine); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if ev := zoneMoveFor(t, g, mine); ev.Cause != MoveCauseManual {
		t.Errorf("cause = %q, want manual", ev.Cause)
	}
}

// TestForetellIsASpecialAction: foretell's hand-to-exile move names
// itself — Ranar's "cards put into exile from your hand" clause is
// cause-agnostic, but the event must not claim an unrelated item.
func TestForetellIsASpecialAction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(Card{InstanceID: id, Name: "Foretold", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	g.mu.Lock()
	resolvingAbilityOf(g, g.Seats[1].ID)
	err := g.foretellLocked(me, id, SpecialAction{CastCost: "{1}"})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("foretellLocked: %v", err)
	}
	ev := zoneMoveFor(t, g, id)
	if ev.Cause != MoveCauseSpecialAction || ev.CauseController != me.ID {
		t.Errorf("cause = %q / %s, want special_action / me", ev.Cause, ev.CauseController)
	}
}

// TestPausedExileKeepsItsCause: a commander exiled by an effect pauses
// on CR 903.9. By the time its owner answers, the resolving slot may
// say anything; the cause was captured onto the route when the move
// was asked for.
func TestPausedExileKeepsItsCause(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmd := seatCommander(t, g.Battlefield, opp)
	g.mu.Lock()
	resolvingAbilityOf(g, me.ID)
	err := g.ExileCardForEffect(cmd)
	g.resolving = nil // play has moved on by the time the owner answers
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, opp)
	if err := g.ResolveOptionalReplacement(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	ev := zoneMoveFor(t, g, cmd)
	if !ExiledBySpellOrAbilityOf(ev, me.ID) {
		t.Errorf("cause after the pause = %q / %s, want effect / me", ev.Cause, ev.CauseController)
	}
}
