package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4259 (CR 616.1, CR 406.3): a face-down exile that stops on an ordering
// prompt reports nothing until the prompt is answered, so a clause that reads
// "that card" has to run from the answer. Necropotence is that clause.

// orderExileOf registers two mandatory replacements over `card`'s move to
// exile, so CR 616.1 asks the affected player which goes first and the move
// waits for the answer.
func orderExileOf(g *game.Game, card uuid.UUID) {
	g.WithWriteLock(func() {
		for _, label := range []string{"first order", "second order"} {
			g.RegisterReplacementForTest(game.ReplacementEffect{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
					return ev.Kind == game.RepEventMove && ev.CardID == card && ev.NewZone == game.ZoneExile
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   label,
			})
		}
	})
}

// answerExileOrder answers the one open ordering prompt.
func answerExileOrder(t *testing.T, g *game.Game) {
	t.Helper()
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("pending choices %+v, want the one ordering prompt", g.PendingChoices)
	}
	prompt := g.PendingChoices[0]
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// necroPaused activates Necropotence over a top card whose exile asks for an
// ordering, and resolves the ability up to the open prompt.
func necroPaused(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	necro := seedReplacementPermanent(g, necropotenceOracle, "Necropotence", me.ID)
	needle := pushLibraryTopForTest(me, "Needle")
	orderExileOf(g, needle)
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open after the ability resolved, want the exile's ordering prompt", len(g.PendingChoices))
	}
	if !me.Library.Contains(needle) {
		t.Fatal("the card left the library while the prompt was open")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("%d delayed triggers queued before the card was exiled", len(g.DelayedTriggers))
	}
	return g, me, needle
}

func deliverAtOwnEndStep(t *testing.T, g *game.Game, me *game.Player, needle uuid.UUID) {
	t.Helper()
	walkSteps(t, g, 40, func() bool {
		return g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepEnd
	})
	passPriorityAroundTable(t, g)
	count := 0
	for _, c := range me.Hand.Cards {
		if c.InstanceID == needle {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("the exiled card is in the hand %d times after the end step, want once", count)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers left after delivery", len(g.DelayedTriggers))
	}
}

// The headline: the prompt is on the exile, and the card still comes back at
// the end step. Before #4259 no delayed trigger was scheduled and the card
// stayed exiled face down for good.
func TestNecropotenceDeliversTheCardAfterAnOrderingPromptOnTheExile(t *testing.T) {
	g, me, needle := necroPaused(t)

	answerExileOrder(t, g)

	exiled, ok := findInZone(g.Exile, needle)
	if !ok {
		t.Fatal("the card is not exiled once the prompt is answered")
	}
	if !exiled.FaceDown {
		t.Error("Necropotence's exile must be face down")
	}
	for i, p := range g.Seats {
		if exiled.IsKnownTo(p.ID) {
			t.Errorf("seat %d can read a face-down exiled card", i)
		}
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after the answer, want 1", len(g.DelayedTriggers))
	}
	deliverAtOwnEndStep(t, g, me, needle)
}

// An undo into the open prompt and the same answer again schedule one
// delivery, not two, and the card arrives once.
func TestNecropotenceOrderingPromptReplaysAfterAnUndo(t *testing.T) {
	g, _, needle := necroPaused(t)
	promptOpen := g.Clone()

	answerExileOrder(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after the first answer, want 1", len(g.DelayedTriggers))
	}

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("the rewind left %d delayed triggers", len(g.DelayedTriggers))
	}
	answerExileOrder(t, g)

	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after the replayed answer, want 1", len(g.DelayedTriggers))
	}
	if _, ok := findInZone(g.Exile, needle); !ok {
		t.Fatal("the card is not exiled after the replayed answer")
	}
	// An undo restores the game's fields in place, so the seat is read again.
	deliverAtOwnEndStep(t, g, g.Seats[0], needle)
}

// A restore point taken once the answer is in holds the scheduled delivery:
// the restored table delivers the card at the end step. (While the prompt is
// open the table is not a restore point, the same as any prompt that holds a
// continuation.)
func TestNecropotenceDeliveryScheduledFromAPromptSurvivesARestore(t *testing.T) {
	g, _, needle := necroPaused(t)
	if g.CaptureSnapshot().Restorable() {
		t.Fatal("a table with an ordering prompt open is a restore point: the census no longer counts it")
	}
	answerExileOrder(t, g)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("the table after the answer is not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(restored.DelayedTriggers) != 1 {
		t.Fatalf("restored table has %d delayed triggers, want 1", len(restored.DelayedTriggers))
	}
	deliverAtOwnEndStep(t, restored, restored.Seats[0], needle)
}

// An empty library still pays the life and schedules nothing.
func TestNecropotenceOnAnEmptyLibrarySchedulesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	necro := seedReplacementPermanent(g, necropotenceOracle, "Necropotence", me.ID)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want the 1 paid", life, me.Life)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers for an exile of nothing", len(g.DelayedTriggers))
	}
}

// The primitive's Then is the clause that reads the cards. It sees all of
// them, top first, from the answer to a prompt on the first.
func TestExileTopFaceDownThenSeesTheWholeBatchAfterAPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var top []uuid.UUID
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			top = append(top, me.Library.Cards[len(me.Library.Cards)-1-i].InstanceID)
		}
	})
	orderExileOf(g, top[0])

	var seen []uuid.UUID
	calls := 0
	g.WithWriteLock(func() {
		err := ExileTopFaceDown{Player: me.ID, N: 3, Then: func(_ *Context, exiled []uuid.UUID) error {
			seen = exiled
			calls++
			return nil
		}}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
		if err != nil {
			t.Errorf("Apply: %v", err)
		}
	})
	if calls != 0 {
		t.Fatalf("Then ran %d times before the prompt was answered", calls)
	}
	answerExileOrder(t, g)
	if calls != 1 {
		t.Fatalf("Then ran %d times, want once", calls)
	}
	if len(seen) != 3 || seen[0] != top[0] || seen[1] != top[1] || seen[2] != top[2] {
		t.Fatalf("Then saw %v, want the three cards %v top first", seen, top)
	}
}

// Then runs with an empty list when there was nothing to exile.
func TestExileTopFaceDownThenRunsWithNothingOnAnEmptyLibraryOrZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, n := range []int{0, 2} {
		g.WithWriteLock(func() { me.Library.Cards = nil })
		calls := 0
		g.WithWriteLock(func() {
			err := ExileTopFaceDown{Player: me.ID, N: n, Then: func(_ *Context, exiled []uuid.UUID) error {
				calls++
				if len(exiled) != 0 {
					t.Errorf("n=%d: Then saw %d cards from an empty library", n, len(exiled))
				}
				return nil
			}}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
			if err != nil {
				t.Errorf("n=%d: Apply: %v", n, err)
			}
		})
		if calls != 1 {
			t.Errorf("n=%d: Then ran %d times, want once", n, calls)
		}
	}
}

// Doomsday Excruciator exiles all but the bottom six cards of each library.
// A prompt on one of those cards, in the middle of the batch, must not leave
// the cards after it on the library.
func TestDoomsdayExcruciatorFinishesAfterAPromptMidBatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	var batch []uuid.UUID
	g.WithWriteLock(func() {
		n := len(me.Library.Cards) - 6
		for i := 0; i < n; i++ {
			batch = append(batch, me.Library.Cards[len(me.Library.Cards)-1-i].InstanceID)
		}
	})
	if len(batch) < 4 {
		t.Fatalf("library gives a batch of %d, want at least 4", len(batch))
	}
	mid := len(batch) / 2
	orderExileOf(g, batch[mid])

	castFromHandForTest(t, g, me, "Doomsday Excruciator", "Creature — Demon", "", s58DoomsdayExcruciator, game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open, want the one mid-batch ordering prompt", len(g.PendingChoices))
	}
	if got := me.Library.Size(); got != 6+len(batch)-mid {
		t.Fatalf("library %d while the prompt is open, want %d (the cards before it left)", got, 6+len(batch)-mid)
	}

	answerExileOrder(t, g)

	for i, id := range batch {
		c, ok := findInZone(g.Exile, id)
		if !ok {
			t.Errorf("card %d of the batch is not in exile", i)
			continue
		}
		if !c.FaceDown {
			t.Errorf("card %d of the batch was exiled face up", i)
		}
	}
	for i, p := range g.Seats {
		if p.Library.Size() != 6 {
			t.Errorf("seat %d keeps %d cards, want 6", i, p.Library.Size())
		}
	}
}
