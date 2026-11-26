package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// targets_chosen_contest_test.go — #6228: game.EventTargetsChosen and
// the top-card contest that may change the targets (Psychic Battle's
// shape), driven through a test card built from the two helpers.

const contestOracle = "test-6228-targets-chosen-contest"

func registerContestCard(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     contestOracle,
		Name:         "Contest Engine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAPlayerChoosesTargets("Contest Engine",
				"Whenever a player chooses one or more targets, each player reveals the top card of their library. The player who reveals the card with the greatest mana value may change the target or targets.",
				effects.TopCardContestMayRetarget),
		},
	})
}

// topCardWithManaValue puts a card of the given mana cost on top of p's
// library.
func topCardWithManaValue(p *game.Player, cost string) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: "Top " + cost, TypeLine: "Sorcery", ManaCost: cost, Owner: p.ID, Controller: p.ID}
	p.Library.PushTop(c)
	return c.InstanceID
}

func contestTable(t *testing.T, costs [4]string) *game.Game {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	registerContestCard(t)
	effectstest.B12Push(g, g.Seats[2].ID, "Contest Engine", "Enchantment", contestOracle, 0, 0)
	for i, cost := range costs {
		topCardWithManaValue(g.Seats[i], cost)
	}
	return g
}

// passUntilRetargetOrEmpty passes priority until `chooser` is asked to
// change targets or the stack is empty.
func passUntilRetargetOrEmpty(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 32; i++ {
		if p := effectstest.LatestRetarget(g, chooser); p != nil {
			return p
		}
		if effectstest.StackFullyEmpty(g) {
			return nil
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return effectstest.LatestRetarget(g, chooser)
}

func countTargetsChosen(g *game.Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTargetsChosen {
			n++
		}
	}
	return n
}

// Seat 0 Bolts seat 2. Seat 1 reveals the greatest mana value and
// redirects the Bolt to seat 3; that change does not trigger the card
// again.
func TestTargetsChosenContestWinnerMayChangeTheTargets(t *testing.T) {
	g := contestTable(t, [4]string{"{1}", "{5}", "{2}", "{3}"})
	life2, life3 := g.Seats[2].Life, g.Seats[3].Life
	bolt := effectstest.BoltFrom(t, g, g.Seats[0], g.Seats[2].ID)
	if countTargetsChosen(g) != 1 {
		t.Fatalf("EventTargetsChosen emitted %d times for one cast, want 1", countTargetsChosen(g))
	}
	prompt := passUntilRetargetOrEmpty(t, g, g.Seats[1].ID)
	if prompt == nil {
		t.Fatal("the player who revealed the greatest mana value was not offered the change")
	}
	if prompt.PickTargetMin != 0 {
		t.Error("the change is a may, but the prompt cannot be declined")
	}
	effectstest.AnswerRetargetWith(t, g, g.Seats[1].ID, effectstest.RtPlayerRef(g.Seats[3].ID))
	item := g.StackMeta[bolt]
	if item == nil || item.Targets[0].ID != g.Seats[3].ID {
		t.Fatalf("the Bolt was not redirected: %+v", item)
	}
	if got := len(g.StackMeta); got != 1 {
		t.Fatalf("stack holds %d items after the change, want only the Bolt (its own change must not retrigger it)", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[3].Life != life3-3 || g.Seats[2].Life != life2 {
		t.Errorf("life: seat 2 %d→%d, seat 3 %d→%d; want the Bolt on seat 3", life2, g.Seats[2].Life, life3, g.Seats[3].Life)
	}
}

// A tie for greatest leaves the targets alone.
func TestTargetsChosenContestTieChangesNothing(t *testing.T) {
	g := contestTable(t, [4]string{"{1}", "{5}", "{2}", "{4}{1}"})
	life2 := g.Seats[2].Life
	effectstest.BoltFrom(t, g, g.Seats[0], g.Seats[2].ID)
	for _, p := range g.Seats {
		if passUntilRetargetOrEmpty(t, g, p.ID) != nil {
			t.Fatalf("seat %d was offered a change on a tie", p.Seat)
		}
	}
	if g.Seats[2].Life != life2-3 {
		t.Errorf("seat 2 life %d→%d, want the Bolt to resolve unchanged", life2, g.Seats[2].Life)
	}
}

// A change made by something else (here a plain retarget offer with no
// named source) is a choosing, so it triggers the card.
func TestTargetsChosenFiresOnAnotherEffectsChange(t *testing.T) {
	g := contestTable(t, [4]string{"{1}", "{5}", "{2}", "{3}"})
	bolt := effectstest.BoltFrom(t, g, g.Seats[0], g.Seats[2].ID)
	before := countTargetsChosen(g)
	g.WithWriteLock(func() {
		if err := g.RetargetStackItemForEffect(bolt, g.Seats[0].ID, game.RetargetChangeOne, effectstest.RtPlayerRef(g.Seats[1].ID)); err != nil {
			t.Fatalf("RetargetStackItemForEffect: %v", err)
		}
	})
	if countTargetsChosen(g) != before+1 {
		t.Errorf("a change of targets emitted %d EventTargetsChosen, want 1", countTargetsChosen(g)-before)
	}
}

// Nobody watching: the kind is not emitted, so a table without such a
// card keeps the event log it always had.
func TestTargetsChosenIsNotEmittedWithoutAWatcher(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.BoltFrom(t, g, g.Seats[0], g.Seats[2].ID)
	if n := countTargetsChosen(g); n != 0 {
		t.Errorf("EventTargetsChosen emitted %d times with no watcher", n)
	}
}
