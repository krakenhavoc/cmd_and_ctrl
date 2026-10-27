package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// option_pick_keyed_engine_test.go — #4361 slice 6: the engine's own
// option picks are keyed, so a table waiting on one is a restore point.
// Each test pauses on the prompt, writes a strict restore point, restores
// it through JSON and answers the RESTORED game.

// testNoopOptionPick is the keyed continuation for a test prompt whose
// answer runs nothing (the retired closure Then's `return nil`).
var testNoopOptionPick = RegisterOptionPickThen(testEffectKeyPrefix+"4361-noop-option",
	func(*Game, OptionPicked) error { return nil })

// strictRestoreMidPrompt writes g's restore point, refusing one that is
// not restorable, and restores it strictly through JSON.
func strictRestoreMidPrompt(t *testing.T, g *Game) *Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("the open prompt blocks the restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	return restored
}

// restoredChoice is the restored game's copy of the prompt `id`.
func restoredChoice(t *testing.T, g *Game, id uuid.UUID) *PendingChoice {
	t.Helper()
	_, c := g.findChoiceLocked(id)
	if c == nil {
		t.Fatalf("prompt %v did not survive the restore: %+v", id, g.PendingChoices)
	}
	return c
}

// seatOptionIndex is the offset of `seat`'s option on a seat prompt.
func seatOptionIndex(t *testing.T, c *PendingChoice, seat uuid.UUID) int {
	t.Helper()
	for i, o := range c.PickOptions {
		if o.Player == seat {
			return i
		}
	}
	t.Fatalf("seat %v is not offered: %+v", seat, c.PickOptions)
	return -1
}

// A Siege's "choose Khans or Dragons" (CR 614.12) answered after a
// restore stamps the word on the permanent.
func TestChooseOptionAsEntersSurvivesARestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, c := queueSiegeChoice(t, g, me)
	if c.optionPickResume != nil || c.PickThen == "" {
		t.Fatalf("the as-enters option prompt holds a closure: %+v", c)
	}
	restored := strictRestoreMidPrompt(t, g)
	answerSiege(t, restored, restoredChoice(t, restored, c.ID), me.ID, "Dragons")
	if got := restored.ChosenOptionOf(src); got != "Dragons" {
		t.Errorf("chosen option %q after the restore, want Dragons", got)
	}
}

// True-Name Nemesis's "choose a player" answered after a restore lands on
// the permanent.
func TestChoosePlayerAsEntersSurvivesARestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, them := g.Seats[0], g.Seats[2]
	src := pushSiegePermanent(g, me, "Test Nemesis")
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChoosePlayerAsEntersForEffect(me.ID, src, "choose a player", []uuid.UUID{g.Seats[1].ID, them.ID, g.Seats[3].ID})
	})
	if id == uuid.Nil {
		t.Fatal("nothing was queued")
	}
	restored := strictRestoreMidPrompt(t, g)
	c := restoredChoice(t, restored, id)
	if err := restored.ResolveOptionPick(id, me.ID, seatOptionIndex(t, c, them.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if got := restored.ChosenPlayerOf(src); got != them.ID {
		t.Errorf("chosen player %v after the restore, want %v", got, them.ID)
	}
}

// Sower of Discord's "choose two players": a restore before each answer,
// and the second prompt still leaves the first answer out.
func TestChooseTwoPlayersAsEntersSurvivesARestoreAtEachPrompt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	first, second := g.Seats[1], g.Seats[3]
	src := pushSiegePermanent(g, me, "Test Sower")
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseTwoPlayersAsEntersForEffect(me.ID, src, "choose two players", []uuid.UUID{first.ID, g.Seats[2].ID, second.ID})
	})
	if id == uuid.Nil {
		t.Fatal("nothing was queued")
	}
	g = strictRestoreMidPrompt(t, g)
	c := restoredChoice(t, g, id)
	if err := g.ResolveOptionPick(id, me.ID, seatOptionIndex(t, c, first.ID)); err != nil {
		t.Fatalf("first answer: %v", err)
	}
	next := choosePlayerPromptFor(g, me.ID)
	if next == nil {
		t.Fatal("the second prompt was not asked")
	}
	if next.optionPickResume != nil || next.PickThen == "" {
		t.Fatalf("the second prompt holds a closure: %+v", next)
	}
	for _, o := range next.PickOptions {
		if o.Player == first.ID {
			t.Errorf("the second prompt offers the first answer again")
		}
	}
	g = strictRestoreMidPrompt(t, g)
	c = restoredChoice(t, g, next.ID)
	if err := g.ResolveOptionPick(next.ID, me.ID, seatOptionIndex(t, c, second.ID)); err != nil {
		t.Fatalf("second answer: %v", err)
	}
	if a, b := g.ChosenPlayersOf(src); a != first.ID || b != second.ID {
		t.Errorf("chosen pair %v, %v after the restores, want %v, %v", a, b, first.ID, second.ID)
	}
}

// Curse of Leeches's "attach it to a player": after a restore the Aura is
// still exempt from CR 704.5m while the question is open (the key is the
// marker), and the answer attaches it.
func TestAttachSourceToPlayerSurvivesARestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, third := g.Seats[0], g.Seats[2]
	aura := pushAttachTestCard(g, me.ID, "Test Curse", "Enchantment — Aura Curse")
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.QueueAttachSourceToPlayerForEffect(aura, "attach it") })
	if id == uuid.Nil {
		t.Fatal("nothing was queued")
	}
	restored := strictRestoreMidPrompt(t, g)
	restored.WithWriteLock(func() {
		if !restored.attachPromptOpenForLocked(aura) {
			t.Error("the restored prompt is not recognised as the Aura's attach question")
		}
		restored.runStateChecksLocked()
	})
	if _, ok := battlefieldCardByID(restored, aura); !ok {
		t.Fatal("CR 704.5m took the Aura while its attach question was open")
	}
	c := restoredChoice(t, restored, id)
	if err := restored.ResolveOptionPick(id, me.ID, seatOptionIndex(t, c, third.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	got, _ := battlefieldCardByID(restored, aura)
	if got.AttachedTo.Kind != TargetPlayer || got.AttachedTo.ID != third.ID {
		t.Errorf("attached to %+v after the restore, want %v", got.AttachedTo, third.ID)
	}
}

// CR 616.1's choice between rebound and buyback, answered after a
// restore: the card leaves the stack the chosen way.
func TestResolutionExitChoiceSurvivesARestore(t *testing.T) {
	for _, tc := range []struct {
		name   string
		index  int
		inHand bool
	}{
		{"rebound", 0, false},
		{"buyback", 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newRestorableGame(t)
			stubOptionalCosts(t, testBuybackOracle, []AdditionalCost{
				{Optional: true, Key: BuybackKey, ManaCost: "{0}", Label: "Buyback {0}"},
			})
			me := g.Seats[g.Turn.ActiveSeat]
			id, err := castOptional(t, g, "Bought Back", "Instant", testBuybackOracle, []int{0})
			if err != nil {
				t.Fatalf("cast with buyback: %v", err)
			}
			grantRebound(t, g, id)
			resolveTop(t, g)
			choice := resolutionChoiceOutstanding(g)
			if choice == nil {
				t.Fatal("rebound and buyback both applied and nobody was asked")
			}
			if choice.optionPickResume != nil || choice.PickThen == "" {
				t.Fatalf("the resolution exit prompt holds a closure: %+v", choice)
			}
			g = strictRestoreMidPrompt(t, g)
			me = g.playerByIDLocked(me.ID)
			if !g.Stack.Contains(id) {
				t.Fatal("the card is not waiting on the stack after the restore")
			}
			if err := g.ResolveOptionPick(choice.ID, me.ID, tc.index); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			if got := me.Hand.Contains(id); got != tc.inHand {
				t.Errorf("in hand = %v, want %v", got, tc.inHand)
			}
			if got := g.Exile.Contains(id); got == tc.inHand {
				t.Errorf("in exile = %v, want %v", got, !tc.inHand)
			}
			wantTriggers := 1
			if tc.inHand {
				wantTriggers = 0
			}
			if n := len(reboundTriggers(g)); n != wantTriggers {
				t.Errorf("rebound triggers = %d, want %d", n, wantTriggers)
			}
			if me.Graveyard.Contains(id) || g.Stack.Contains(id) {
				t.Error("the card is also in the graveyard or still on the stack")
			}
		})
	}
}

// A resolution exit prompt dropped unanswered (its chooser left) takes
// the first exit, as the closure form did.
func TestADroppedResolutionExitChoiceTakesTheFirstExit(t *testing.T) {
	g := newRestorableGame(t)
	stubOptionalCosts(t, testBuybackOracle, []AdditionalCost{
		{Optional: true, Key: BuybackKey, ManaCost: "{0}", Label: "Buyback {0}"},
	})
	id, err := castOptional(t, g, "Bought Back", "Instant", testBuybackOracle, []int{0})
	if err != nil {
		t.Fatalf("cast with buyback: %v", err)
	}
	grantRebound(t, g, id)
	resolveTop(t, g)
	choice := resolutionChoiceOutstanding(g)
	if choice == nil {
		t.Fatal("nobody was asked")
	}
	g.WithWriteLock(func() {
		idx, c := g.findChoiceLocked(choice.ID)
		g.dequeueChoiceLocked(idx)
		g.defaultDroppedChoiceLocked(c)
	})
	if !g.Exile.Contains(id) || len(reboundTriggers(g)) != 1 {
		t.Errorf("a dropped exit choice did not take the first exit (rebound): exile %v, triggers %d",
			g.Exile.Contains(id), len(reboundTriggers(g)))
	}
}
