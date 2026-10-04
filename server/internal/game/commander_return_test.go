package game

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// commander_return_test.go — ADR 0115 PR 2: CR 903.9a's state-based
// action, shipped switched OFF. The first test is the one that matters
// for this PR: with the switch off, nothing a commander does is any
// different. The rest switch it on for the length of one test and pin
// the plumbing PR 3 will turn on.

// withCommanderReturnSBA switches ADR 0115's state-based action on for
// one test.
func withCommanderReturnSBA(t *testing.T) {
	t.Helper()
	commanderReturnSBA = true
	t.Cleanup(func() { commanderReturnSBA = false })
}

// commanderOnBattlefield puts a commander card owned by `owner` and
// controlled by `controller` on the battlefield.
func commanderOnBattlefield(g *Game, owner, controller uuid.UUID) uuid.UUID {
	c := NewCommander("Test Commander", owner)
	c.TypeLine = "Legendary Creature — Bear"
	c.Power, c.Toughness = 2, 2
	c.Controller = controller
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// commanderIntoGraveyard moves a battlefield commander straight into
// its owner's graveyard, the way the destroy path ends, without the
// CR 903.9 replacement in between.
func commanderIntoGraveyard(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatalf("commander %s is not on the battlefield", id)
	}
	owner := g.playerByIDLocked(c.Owner)
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, owner.Graveyard, id); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
}

func commanderReturnPrompts(g *Game) []*PendingChoice {
	var out []*PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceCommanderReturn {
			out = append(out, c)
		}
	}
	return out
}

func cardIn(z *Zone, id uuid.UUID) *Card {
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			return &z.Cards[i]
		}
	}
	return nil
}

// TestCommanderReturnIsInertWhileSwitchedOff is PR 2's promise: the
// plumbing ships and nothing changes. A destroyed commander still asks
// the CR 903.9 replacement before it moves, no card is marked, no
// commander_return prompt is ever queued, and the snapshot carries no
// new key.
func TestCommanderReturnIsInertWhileSwitchedOff(t *testing.T) {
	if commanderReturnSBA {
		t.Fatal("commanderReturnSBA is on: ADR 0115 PR 2 ships it off")
	}
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := commanderOnBattlefield(g, owner.ID, owner.ID)

	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceOptionalReplacement {
		t.Fatalf("pending choices = %+v, want today's one CR 903.9 replacement prompt", g.PendingChoices)
	}
	answerOnlyCommanderPrompt(t, g, owner.ID, false)
	c := cardIn(owner.Graveyard, id)
	if c == nil {
		t.Fatal("declining the replacement did not leave the commander in the graveyard")
	}
	if c.CommanderReturnDue {
		t.Error("a commander in the graveyard was marked with the switch off")
	}
	runChecks(g)
	if n := len(commanderReturnPrompts(g)); n != 0 {
		t.Errorf("%d commander_return prompts queued with the switch off", n)
	}

	// A plain MoveCard into exile does not mark it either.
	g.WithWriteLock(func() {
		if _, err := MoveCard(owner.Graveyard, g.Exile, id); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if cardIn(g.Exile, id).CommanderReturnDue {
		t.Error("a commander moved into exile was marked with the switch off")
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "commanderReturnDue") {
		t.Error("the snapshot names commanderReturnDue with the switch off")
	}
}

// TestCommanderReturnAsksTheOwnerOnce: the mark is set as the card
// lands, the check asks once and clears it, and a second pass does not
// ask again.
func TestCommanderReturnAsksTheOwnerOnce(t *testing.T) {
	withCommanderReturnSBA(t)
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := commanderOnBattlefield(g, owner.ID, owner.ID)
	commanderIntoGraveyard(t, g, id)
	if !cardIn(owner.Graveyard, id).CommanderReturnDue {
		t.Fatal("a commander put into a graveyard was not marked")
	}

	runChecks(g)
	prompts := commanderReturnPrompts(g)
	if len(prompts) != 1 {
		t.Fatalf("commander_return prompts = %d, want 1", len(prompts))
	}
	p := prompts[0]
	if p.Chooser != owner.ID || p.Source != id {
		t.Errorf("prompt chooser %s source %s, want the owner %s and the card %s", p.Chooser, p.Source, owner.ID, id)
	}
	if cardIn(owner.Graveyard, id).CommanderReturnDue {
		t.Error("the mark was not cleared once the check asked")
	}
	if !ChoiceBlocksTable(PendingChoiceCommanderReturn) {
		t.Error("commander_return does not block the table")
	}

	runChecks(g)
	if n := len(commanderReturnPrompts(g)); n != 1 {
		t.Errorf("after a second pass there are %d prompts, want still 1", n)
	}
}

// TestCommanderReturnYesMovesItHome: a "yes" moves the card from the
// graveyard to its owner's command zone and says where it came from.
func TestCommanderReturnYesMovesItHome(t *testing.T) {
	withCommanderReturnSBA(t)
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := commanderOnBattlefield(g, owner.ID, owner.ID)
	commanderIntoGraveyard(t, g, id)
	runChecks(g)
	prompt := commanderReturnPrompts(g)[0]
	before := len(g.Events)

	if err := g.ResolveCommanderReturn(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveCommanderReturn: %v", err)
	}
	if cardIn(owner.Command, id) == nil {
		t.Fatal("the commander is not in the command zone after yes")
	}
	if cardIn(owner.Graveyard, id) != nil {
		t.Error("the commander is still in the graveyard")
	}
	found := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventZoneMove && ev.CardID == id && ev.OldZone == ZoneGraveyard && ev.NewZone == ZoneCommand {
			found = true
		}
	}
	if !found {
		t.Error("no EventZoneMove graveyard → command for the yes")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts left after the answer: %+v", g.PendingChoices)
	}
}

// TestCommanderReturnNoIsNotAskedAgain: a "no" leaves it where it is
// and is not asked again on the next pass; moving it from the graveyard
// into exile asks again (ADR 0115 decision 1).
func TestCommanderReturnNoIsNotAskedAgain(t *testing.T) {
	withCommanderReturnSBA(t)
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := commanderOnBattlefield(g, owner.ID, owner.ID)
	commanderIntoGraveyard(t, g, id)
	runChecks(g)
	if err := g.ResolveCommanderReturn(commanderReturnPrompts(g)[0].ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveCommanderReturn: %v", err)
	}
	if cardIn(owner.Graveyard, id) == nil {
		t.Fatal("a no moved the commander")
	}
	runChecks(g)
	if n := len(commanderReturnPrompts(g)); n != 0 {
		t.Fatalf("a declined commander was asked again (%d prompts)", n)
	}

	g.WithWriteLock(func() {
		if _, err := MoveCard(owner.Graveyard, g.Exile, id); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	runChecks(g)
	prompts := commanderReturnPrompts(g)
	if len(prompts) != 1 {
		t.Fatalf("exiled from the graveyard: %d prompts, want 1", len(prompts))
	}
	if err := g.ResolveCommanderReturn(prompts[0].ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveCommanderReturn: %v", err)
	}
	if cardIn(owner.Command, id) == nil {
		t.Error("yes from exile did not reach the command zone")
	}
}

// TestCommanderReturnAsksTheOwnerNotTheController: a stolen commander
// that dies is its owner's question (CR 903.9a, "its owner").
func TestCommanderReturnAsksTheOwnerNotTheController(t *testing.T) {
	withCommanderReturnSBA(t)
	g := newActiveGame(t)
	owner, thief := g.Seats[1], g.Seats[0]
	id := commanderOnBattlefield(g, owner.ID, thief.ID)
	commanderIntoGraveyard(t, g, id)
	runChecks(g)
	prompts := commanderReturnPrompts(g)
	if len(prompts) != 1 || prompts[0].Chooser != owner.ID {
		t.Fatalf("prompts %+v, want one to the owner %s", prompts, owner.ID)
	}
	if err := g.ResolveCommanderReturn(prompts[0].ID, thief.ID, true); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("the controller answered: err = %v, want ErrNotTheChooser", err)
	}
}

// TestCommanderReturnSkipsTokensAndDepartedOwners: a token is never a
// commander (CR 903.3), and a commander whose owner has left the game
// is not asked about (CR 800.4a).
func TestCommanderReturnSkipsTokensAndDepartedOwners(t *testing.T) {
	withCommanderReturnSBA(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	tok := NewCommander("Test Commander", me.ID)
	tok.TypeLine = "Token Legendary Creature — Bear"
	g.Battlefield.PushTop(tok)
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, tok.InstanceID); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if c := cardIn(me.Graveyard, tok.InstanceID); c != nil && c.CommanderReturnDue {
		t.Error("a token was marked")
	}

	gone := g.Seats[1]
	id := commanderOnBattlefield(g, gone.ID, gone.ID)
	commanderIntoGraveyard(t, g, id)
	gone.Eliminated = true
	runChecks(g)
	if n := len(commanderReturnPrompts(g)); n != 0 {
		t.Errorf("%d prompts, want none", n)
	}
}

// TestCommanderReturnHoldsTheTriggers: while the question is open no
// waiting trigger goes on the stack (CR 704.3, ADR 0115 decision 3);
// the answer releases them.
func TestCommanderReturnHoldsTheTriggers(t *testing.T) {
	withCommanderReturnSBA(t)
	g := newActiveGame(t)
	owner := g.Seats[0]
	src := pushCreatureToBattlefield(t, g, owner)
	id := commanderOnBattlefield(g, owner.ID, owner.ID)
	commanderIntoGraveyard(t, g, id)
	trig := uuid.New()
	g.WithWriteLock(func() {
		g.PendingTriggers = append(g.PendingTriggers, &StackItem{
			ID: trig, Kind: StackItemTriggered, Controller: owner.ID, Owner: owner.ID,
			SourceCardID: src, Label: "waiting trigger",
		})
	})
	runChecks(g)
	if len(commanderReturnPrompts(g)) != 1 {
		t.Fatal("no commander_return prompt")
	}
	if len(g.PendingTriggers) != 1 || g.StackMeta[trig] != nil {
		t.Fatal("a waiting trigger went on the stack while the CR 903.9a question was open")
	}
	if err := g.ResolveCommanderReturn(commanderReturnPrompts(g)[0].ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveCommanderReturn: %v", err)
	}
	if len(g.PendingTriggers) != 0 || g.StackMeta[trig] == nil {
		t.Error("the answer did not release the waiting trigger onto the stack")
	}
}

// commanderDiesInCleanup parks a game in its cleanup step on a
// commander_return prompt: the commander lands in its graveyard during
// the end step and the cleanup step's CR 514.3a check asks.
func commanderDiesInCleanup(t *testing.T) (*Game, *Player, *PendingChoice) {
	t.Helper()
	withCommanderReturnSBA(t)
	g := newActiveGame(t)
	advanceTo(t, g, StepEnd)
	owner := g.Seats[g.Turn.ActiveSeat]
	id := commanderOnBattlefield(g, owner.ID, owner.ID)
	commanderIntoGraveyard(t, g, id)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into cleanup: %v", err)
	}
	if g.Turn.Step != StepCleanup {
		t.Fatalf("step = %v, want cleanup", g.Turn.Step)
	}
	prompts := commanderReturnPrompts(g)
	if len(prompts) != 1 {
		t.Fatalf("cleanup asked %d commander_return prompts, want 1", len(prompts))
	}
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("priority holder = %d while the question is open, want nobody", g.Turn.PriorityHolder)
	}
	return g, owner, prompts[0]
}

// TestCommanderReturnYesInCleanupGivesPriority: CR 514.3a — a "yes" is
// a state-based action performed, so the active player gets priority
// in the cleanup step.
func TestCommanderReturnYesInCleanupGivesPriority(t *testing.T) {
	g, owner, prompt := commanderDiesInCleanup(t)
	active := g.Turn.ActiveSeat
	if err := g.ResolveCommanderReturn(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveCommanderReturn: %v", err)
	}
	if g.Turn.Step != StepCleanup || g.Turn.PriorityHolder != active {
		t.Errorf("after yes: step %v, priority %d; want cleanup with priority to %d", g.Turn.Step, g.Turn.PriorityHolder, active)
	}
}

// TestCommanderReturnNoInCleanupEndsTheTurn: a "no" performs nothing,
// so nobody gets priority and the turn ends (CR 514.3).
func TestCommanderReturnNoInCleanupEndsTheTurn(t *testing.T) {
	g, owner, prompt := commanderDiesInCleanup(t)
	active := g.Turn.ActiveSeat
	if err := g.ResolveCommanderReturn(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveCommanderReturn: %v", err)
	}
	if g.Turn.ActiveSeat == active && g.Turn.Step == StepCleanup {
		t.Errorf("after no: still in seat %d's cleanup with priority %d; want the turn ended", active, g.Turn.PriorityHolder)
	}
}

// TestCommanderReturnPromptIsARestorePoint: the prompt is plain data,
// so a table waiting on it captures as a restore point, restores with
// the question open, and the question can be answered.
func TestCommanderReturnPromptIsARestorePoint(t *testing.T) {
	withCommanderReturnSBA(t)
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := commanderOnBattlefield(g, owner.ID, owner.ID)
	commanderIntoGraveyard(t, g, id)
	runChecks(g)
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a table waiting on commander_return is not a restore point: %v", snap.Continuations.Labels)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r, err := back.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	prompts := commanderReturnPrompts(r)
	if len(prompts) != 1 || prompts[0].Source != id {
		t.Fatalf("restored prompts %+v, want the one about %s", prompts, id)
	}
	if err := r.ResolveCommanderReturn(prompts[0].ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveCommanderReturn after restore: %v", err)
	}
	if cardIn(r.Seats[0].Command, id) == nil {
		t.Error("the restored yes did not reach the command zone")
	}
}

// TestCommanderReturnDueSurvivesASnapshot: the mark is carried, so a
// table captured between the move and the check still asks.
func TestCommanderReturnDueSurvivesASnapshot(t *testing.T) {
	withCommanderReturnSBA(t)
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := commanderOnBattlefield(g, owner.ID, owner.ID)
	commanderIntoGraveyard(t, g, id)
	r, err := g.CaptureSnapshot().RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if c := cardIn(r.Seats[0].Graveyard, id); c == nil || !c.CommanderReturnDue {
		t.Fatal("the mark did not survive the snapshot")
	}
	runChecks(r)
	if n := len(commanderReturnPrompts(r)); n != 1 {
		t.Errorf("restored table asked %d times, want 1", n)
	}
}

// TestRestoreRefusesAnUnknownChoiceKind: a pending-choice kind this
// binary does not know was written by a newer one, and the file is
// refused rather than restored into a question nobody can answer
// (ADR 0115 §8).
func TestRestoreRefusesAnUnknownChoiceKind(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(PendingChoice{
			Kind: PendingChoiceKind("from_a_newer_binary"), Chooser: g.Seats[0].ID, Count: 1,
		})
	})
	if _, err := g.CaptureSnapshot().Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("Restore = %v, want ErrUnknownEffectKey", err)
	}
	// Every kind this binary declares restores.
	for _, kind := range ClassifiedChoiceKinds() {
		if !KnownChoiceKind(kind) {
			t.Errorf("%q is classified but not known", kind)
		}
	}
}
