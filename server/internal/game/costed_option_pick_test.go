package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// costed_option_pick_test.go — #2854: option_pick options that carry a
// mana cost. Winter's Chill's "its controller may pay {1} or {2}" is
// one choice with three outcomes; the options it cannot pay are not
// offered (CR 118.3), the chosen one is paid through the pay_unless
// payer, an option the board can no longer pay is refused with the
// prompt still open, and a keyed prompt is a restore point.

// costedOptions is the Winter's Chill shape: nothing, {1}, {2}.
func costedOptions() []ChoiceOption {
	return []ChoiceOption{
		{Label: "Pay nothing"},
		{Label: "Pay {1}", ManaCost: "{1}"},
		{Label: "Pay {2}", ManaCost: "{2}"},
	}
}

// queueCostedPick queues the three-way question for chooser and records
// the index Then is handed in *got (-2 until it runs).
func queueCostedPick(t *testing.T, g *Game, chooser uuid.UUID, got *int) *PendingChoice {
	t.Helper()
	*got = -2
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser:  chooser,
			Source:   uuid.New(),
			Question: "Pay {1} or {2}?",
			Options:  costedOptions(),
			Then: func(_ *Game, index int) error {
				*got = index
				return nil
			},
		})
	})
	if id == uuid.Nil {
		t.Fatal("nothing was queued")
	}
	_, c := g.findChoiceLocked(id)
	return c
}

func offeredLabels(c *PendingChoice) []string {
	var out []string
	for _, o := range c.PickOptions {
		out = append(out, o.Label)
	}
	return out
}

func untappedCount(g *Game, ids ...uuid.UUID) int {
	n := 0
	for _, id := range ids {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == id && !c.Tapped {
				n++
			}
		}
	}
	return n
}

func mountains(g *Game, controller uuid.UUID, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, pushBattlefieldForTest(g, controller, "Mountain", "Basic Land — Mountain", ""))
	}
	return out
}

func TestCostedOptionPickOffersEveryOptionThePayerCanAfford(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	lands := mountains(g, payer.ID, 2)
	var got int
	c := queueCostedPick(t, g, payer.ID, &got)
	if len(c.PickOptions) != 3 {
		t.Fatalf("offered %v, want all three", offeredLabels(c))
	}
	if c.PickOptions[2].ManaCost != "{2}" {
		t.Errorf("the {2} option lost its cost: %+v", c.PickOptions[2])
	}
	if err := g.ResolveOptionPick(c.ID, payer.ID, 2); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if got != 2 {
		t.Errorf("Then got %d, want 2", got)
	}
	if n := untappedCount(g, lands...); n != 0 {
		t.Errorf("%d Mountains still untapped after paying {2}, want 0", n)
	}
	if len(g.PendingChoices) != 0 {
		t.Error("the prompt is still open")
	}
}

func TestCostedOptionPickHidesTheOptionThePayerCannotAfford(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	lands := mountains(g, payer.ID, 1)
	var got int
	c := queueCostedPick(t, g, payer.ID, &got)
	if len(c.PickOptions) != 2 || c.PickOptions[1].ManaCost != "{1}" {
		t.Fatalf("offered %v, want nothing and {1}", offeredLabels(c))
	}
	if err := g.ResolveOptionPick(c.ID, payer.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	// The card printed {1} at index 1; the offered list is shorter but
	// Then still hears the printed index.
	if got != 1 {
		t.Errorf("Then got %d, want 1 (the printed {1} option)", got)
	}
	if n := untappedCount(g, lands...); n != 0 {
		t.Error("the Mountain was not tapped for {1}")
	}
}

func TestCostedOptionPickWithNoManaOffersOnlyPayingNothing(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	var got int
	c := queueCostedPick(t, g, payer.ID, &got)
	if len(c.PickOptions) != 1 || c.PickOptions[0].ManaCost != "" {
		t.Fatalf("offered %v, want only paying nothing", offeredLabels(c))
	}
	if err := g.ResolveOptionPick(c.ID, payer.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if got != 0 {
		t.Errorf("Then got %d, want 0", got)
	}
}

func TestCostedOptionPickRefusesAnOptionNoLongerAffordable(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	lands := mountains(g, payer.ID, 2)
	var got int
	c := queueCostedPick(t, g, payer.ID, &got)
	if len(c.PickOptions) != 3 {
		t.Fatalf("offered %v, want all three", offeredLabels(c))
	}
	// The board moves under the open question: one Mountain is tapped.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == lands[0] {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	err := g.ResolveOptionPick(c.ID, payer.ID, 2)
	if !errors.Is(err, ErrInsufficientMana) {
		t.Fatalf("answering {2} with one land: err = %v, want ErrInsufficientMana", err)
	}
	if got != -2 {
		t.Errorf("Then ran (%d) on a refused answer", got)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatal("the refused prompt was dequeued")
	}
	if untappedCount(g, lands[1]) != 1 {
		t.Error("the refused answer tapped the remaining Mountain")
	}
	// The cheaper option is still payable and is taken.
	if err := g.ResolveOptionPick(c.ID, payer.ID, 1); err != nil {
		t.Fatalf("answering {1}: %v", err)
	}
	if got != 1 {
		t.Errorf("Then got %d, want 1", got)
	}
}

func TestCostedOptionPickRefusesACostOnTheFirstOption(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	mountains(g, payer.ID, 2)
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser: payer.ID,
			Source:  uuid.New(),
			Options: []ChoiceOption{{Label: "Pay {1}", ManaCost: "{1}"}, {Label: "Pay nothing"}},
			Then:    func(*Game, int) error { return nil },
		})
	})
	if id != uuid.Nil {
		t.Fatal("a prompt whose always-legal option costs mana was queued")
	}
}

// The keyed form: no closure on the prompt, so a game waiting on it is a
// restore point, and the restored game answers it through the running
// binary's continuation.
var (
	costedKeyedSeen []OptionPicked
	costedKeyed     = RegisterOptionPickThen(testEffectKeyPrefix+"costed-option", func(_ *Game, r OptionPicked) error {
		costedKeyedSeen = append(costedKeyedSeen, r)
		return nil
	})
)

func TestKeyedCostedOptionPickSurvivesARestore(t *testing.T) {
	costedKeyedSeen = nil
	g := newActiveGame(t)
	payer := g.Seats[1]
	lands := mountains(g, payer.ID, 1)
	carried := uuid.New()
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser:  payer.ID,
			Source:   uuid.New(),
			Question: "Pay {1} or {2}?",
			Options:  costedOptions(),
			ThenKey:  costedKeyed,
			Carry:    []uuid.UUID{carried},
		})
	})
	if id == uuid.Nil {
		t.Fatal("nothing was queued")
	}
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a keyed option pick blocks the restore point: %+v", snap.Continuations)
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
	_, c := restored.findChoiceLocked(id)
	if c == nil {
		t.Fatal("the prompt did not survive the restore")
	}
	if len(c.PickOptions) != 2 || c.PickOptions[1].ManaCost != "{1}" {
		t.Fatalf("restored options %+v, want nothing and {1}", c.PickOptions)
	}
	if err := restored.ResolveOptionPick(id, payer.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick on the restored game: %v", err)
	}
	if len(costedKeyedSeen) != 1 {
		t.Fatalf("continuation ran %d times, want 1", len(costedKeyedSeen))
	}
	r := costedKeyedSeen[0]
	if r.Option == nil || r.Option.ManaCost != "{1}" || len(r.Carry) != 1 || r.Carry[0] != carried {
		t.Errorf("continuation got %+v, want the {1} option and the carried ID", r)
	}
	if untappedCount(restored, lands...) != 0 {
		t.Error("the restored game did not tap the Mountain for {1}")
	}
	// The live game is untouched by the restored one's answer.
	if len(g.PendingChoices) != 1 {
		t.Error("answering the restored game dequeued the live one")
	}
}

func TestKeyedOptionPickRefusesAnUnknownKeyAtRestore(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	g.WithWriteLock(func() {
		g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser: payer.ID,
			Source:  uuid.New(),
			Options: costedOptions(),
			ThenKey: costedKeyed,
		})
	})
	snap := g.CaptureSnapshot()
	for i := range snap.PendingChoices {
		snap.PendingChoices[i].PickThen = "test/never-registered"
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown option continuation: err = %v, want ErrUnknownEffectKey", err)
	}
}
