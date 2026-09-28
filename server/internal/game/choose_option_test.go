package game

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// choose_option_test.go — #1572, CR 614.12's anchor-word form: "As
// this enters, choose Khans or Dragons". The engine half — the prompt,
// the stored answer, its CR 400.7 lifecycle, its carry through undo and
// the snapshot, and the ADR 0071 gate that reads it. The Sieges have
// their own tests in the effects package.

// pushSiegePermanent puts a bare enchantment on the battlefield for the
// as-enters prompt to be about. No catalog hook is involved.
func pushSiegePermanent(g *Game, owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Enchantment"
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// queueSiegeChoice queues the as-enters prompt for a fresh permanent
// and returns the permanent and the prompt.
func queueSiegeChoice(t *testing.T, g *Game, owner *Player) (uuid.UUID, *PendingChoice) {
	t.Helper()
	src := pushSiegePermanent(g, owner, "Test Siege")
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseOptionAsEntersForEffect(owner.ID, src, "Test Siege — choose Khans or Dragons",
			[]string{"Khans", "Dragons"})
	})
	if id == uuid.Nil {
		t.Fatal("nothing was queued")
	}
	c := choosePlayerPromptFor(g, owner.ID)
	if c == nil || c.ID != id {
		t.Fatalf("the prompt is addressed to the controller: %+v", g.PendingChoices)
	}
	return src, c
}

// answerSiege picks `word` off the open prompt through the wire
// resolver.
func answerSiege(t *testing.T, g *Game, c *PendingChoice, chooser uuid.UUID, word string) {
	t.Helper()
	for i, opt := range c.PickOptions {
		if opt.Label == word {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("%q is not offered: %v", word, optionLabels(c))
}

// TestTheOptionPromptRidesTheExistingKind — no new PendingChoiceKind:
// the gate, the enumerator, the wire and the client modal answer it
// unchanged. The words are offered in printed order, as the card's own
// labels, and nothing is stored until the controller answers.
func TestTheOptionPromptRidesTheExistingKind(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[1]
	src, c := queueSiegeChoice(t, g, owner)

	if c.Kind != PendingChoiceOptionPick {
		t.Errorf("kind %s, want the existing option_pick", c.Kind)
	}
	if !ChoiceBlocksTable(c.Kind) {
		t.Error("an unanswered as-enters choice stops the table")
	}
	if c.Source != src {
		t.Errorf("the prompt names the entering permanent: %v", c.Source)
	}
	if got := optionLabels(c); strings.Join(got, ",") != "Khans,Dragons" {
		t.Errorf("options %v, want the printed words in printed order", got)
	}
	for _, opt := range c.PickOptions {
		if opt.Player != uuid.Nil || len(opt.Cards) > 0 {
			t.Errorf("a word option names no seat and no card: %+v", opt)
		}
	}
	if got := g.ChosenOptionOf(src); got != "" {
		t.Errorf("nothing is chosen before the answer: %q", got)
	}
}

// TestAnsweringTheOptionPromptStampsThePermanent — the answer lands on
// the permanent (not a stack item) and is announced in the event log.
func TestAnsweringTheOptionPromptStampsThePermanent(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[1]
	src, c := queueSiegeChoice(t, g, owner)

	answerSiege(t, g, c, owner.ID, "Dragons")
	if got := g.ChosenOptionOf(src); got != "Dragons" {
		t.Errorf("stored %q, want Dragons", got)
	}
	if left := choosePlayerPromptFor(g, owner.ID); left != nil {
		t.Errorf("the prompt closed: %+v", left)
	}
	var logged *Event
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == EventOptionChosen {
			logged = &g.Events[i]
			break
		}
	}
	if logged == nil {
		t.Fatal("the answer is recorded in the event log")
	}
	if logged.CardID != src || logged.Label != "Dragons" || logged.Actor != owner.ID {
		t.Errorf("the log entry says who chose which word for which permanent: %+v", logged)
	}
}

// TestTheOptionGateSwitchesOnlyTheChosenLine — ADR 0071's predicate
// for the new kind. An empty answer matches NEITHER word: a Siege
// whose controller has not answered has neither ability, never both.
func TestTheOptionGateSwitchesOnlyTheChosenLine(t *testing.T) {
	khans, dragons := ChosenOptionIs("Khans"), ChosenOptionIs("Dragons")
	if !khans.IsGate() {
		t.Fatal("a chosen-option gate gates")
	}
	cases := []struct {
		chosen       string
		wantK, wantD bool
	}{
		{"", false, false},
		{"Khans", true, false},
		{"Dragons", false, true},
		{"khans", false, false}, // exact: the prompt stamps the card's own spelling
	}
	for _, tc := range cases {
		c := Card{ChosenOption: tc.chosen}
		if got := khans.Active(c); got != tc.wantK {
			t.Errorf("chosen %q: Khans line active = %v, want %v", tc.chosen, got, tc.wantK)
		}
		if got := dragons.Active(c); got != tc.wantD {
			t.Errorf("chosen %q: Dragons line active = %v, want %v", tc.chosen, got, tc.wantD)
		}
	}
	if (Designation{Kind: DesignationChosenOption}).Active(Card{}) {
		t.Error("a gate naming no word is never satisfied, even by an empty answer")
	}
}

// TestTheChosenOptionIsClearedOnTheWayOut — CR 400.7 at the first
// site: a Siege that leaves the battlefield forgets its word, so a
// bounced and recast one chooses again and one in a graveyard has
// neither ability.
func TestTheChosenOptionIsClearedOnTheWayOut(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, c := queueSiegeChoice(t, g, me)
	answerSiege(t, g, c, me.ID, "Khans")

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Hand, src); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	for _, h := range me.Hand.Cards {
		if h.InstanceID == src && h.ChosenOption != "" {
			t.Errorf("in hand: ChosenOption = %q, want empty (CR 400.7)", h.ChosenOption)
		}
	}
	if got := g.ChosenOptionOf(src); got != "" {
		t.Errorf("ChosenOptionOf a permanent that has left: %q", got)
	}
}

// TestTheChosenOptionIsClearedByTheNewObjectReset — the second CR
// 400.7 site: an entry that mints a new object wipes it, so a blinked
// Siege chooses again though it never passed through zone.go's exit.
func TestTheChosenOptionIsClearedByTheNewObjectReset(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, c := queueSiegeChoice(t, g, me)
	answerSiege(t, g, c, me.ID, "Khans")

	var newID uuid.UUID
	g.WithWriteLock(func() { newID = g.resetAsNewObjectLocked(src) })
	if newID == uuid.Nil {
		t.Fatal("the reset ran")
	}
	if got := g.ChosenOptionOf(newID); got != "" {
		t.Errorf("the new object still chose %q", got)
	}
}

// TestAnAnswerForAPermanentThatLeftLandsNowhere — the prompt is
// asynchronous. A Siege that left before its controller answered takes
// no answer with it, and a card elsewhere is never stamped.
func TestAnAnswerForAPermanentThatLeftLandsNowhere(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, _ := queueSiegeChoice(t, g, me)
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, src); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if live := choosePlayerPromptFor(g, me.ID); live != nil {
		answerSiege(t, g, live, me.ID, "Khans")
	}
	for _, gy := range me.Graveyard.Cards {
		if gy.ChosenOption != "" {
			t.Errorf("a card in the graveyard was stamped %q", gy.ChosenOption)
		}
	}
}

// TestTheChosenOptionSurvivesAnUndo — carried by clone. A restore that
// lost it would bring the Siege back with neither ability, silently.
func TestTheChosenOptionSurvivesAnUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, c := queueSiegeChoice(t, g, me)
	answerSiege(t, g, c, me.ID, "Dragons")

	snap := g.Clone()
	g.WithWriteLock(func() {
		g.Battlefield.Cards[findCardOnBattlefield(g, src)].ChosenOption = ""
	})
	g.RestoreFrom(snap)
	if got := g.ChosenOptionOf(src); got != "Dragons" {
		t.Errorf("after the restore the Siege chose %q, want Dragons", got)
	}
}

// TestTheChosenOptionSurvivesASnapshot — the other carry, asserted on
// the RESTORED game rather than a re-captured snapshot (which would be
// blind to a field missing from both projections).
func TestTheChosenOptionSurvivesASnapshot(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0]
	src, c := queueSiegeChoice(t, g, me)
	answerSiege(t, g, c, me.ID, "Dragons")

	_, restored := roundTrip(t, g)
	var got string
	restored.ReadSnapshot(func() { got = restored.ChosenOptionOf(src) })
	if got != "Dragons" {
		t.Errorf("restored ChosenOption = %q, want Dragons — a restored Siege would have neither ability", got)
	}
}

// TestTheChosenOptionIsNotCopiable — CR 707.2. The copiable values are
// the printed characteristics and the answer is not one of them, so a
// copy of a Siege makes its own choice as IT enters. The projection has
// no slot for it, which is the assertion.
func TestTheChosenOptionIsNotCopiable(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, c := queueSiegeChoice(t, g, me)
	answerSiege(t, g, c, me.ID, "Dragons")

	var vals PrintedValues
	g.ReadSnapshot(func() {
		live, _ := g.LookupCardForEffect(src)
		vals = CopiableValuesOf(live)
	})
	if strings.Contains(fmt.Sprintf("%+v", vals), "Dragons") {
		t.Error("the chosen option reached the copiable values — a copy of a Siege chooses its own")
	}
}

// TestADroppedOptionPromptStoresNothing — #1006: a prompt withdrawn
// unanswered (its chooser left) runs the continuation with
// NoChoiceIndex, which must store nothing rather than pick a word on
// the departed player's behalf.
func TestADroppedOptionPromptStoresNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[1]
	src, _ := queueSiegeChoice(t, g, owner)

	// Stamp a live permanent the continuation could wrongly write to,
	// then run the drop path directly: Concede would also remove the
	// Siege, which would hide a bad default behind CR 800.4a.
	var dropped *PendingChoice
	g.WithWriteLock(func() {
		for i, pc := range g.PendingChoices {
			if pc != nil && pc.Source == src {
				dropped = pc
				g.dequeueChoiceLocked(i)
				break
			}
		}
		g.defaultDroppedChoiceLocked(dropped)
	})
	if dropped == nil {
		t.Fatal("no prompt to drop")
	}
	if got := g.ChosenOptionOf(src); got != "" {
		t.Errorf("a dropped prompt chose %q on the chooser's behalf", got)
	}
}
