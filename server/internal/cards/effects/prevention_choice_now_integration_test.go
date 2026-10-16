package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevention_choice_now_integration_test.go — #2040 against the two
// seams that landed beside it: CR 615.13's "damage was prevented"
// announcements (#2047, Selfless Squire), which wait in the same
// follow-up queue a question holds, and #2056's optional redirection
// (Blood of the Martyr), which asks its own question inside a damage
// event.

// squireTriggersLocked is squireTriggers for a caller that already holds
// the game's lock: a damage continuation.
func squireTriggersLocked(g *game.Game, squire uuid.UUID) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == squire {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered && it.SourceCardID == squire {
			n++
		}
	}
	return n
}

// openPrompts counts the prompts open on the table.
func openPrompts(g *game.Game) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil {
			n++
		}
	}
	return n
}

// settleOrderingTriggers passes priority until the stack is empty,
// answering each trigger-order prompt in the order it offers (Khârn's
// Sigil and the Squire's trigger are one controller's, CR 603.3b).
func settleOrderingTriggers(t *testing.T, g *game.Game) {
	t.Helper()
	for range 8 {
		passPriorityAroundTable(t, g)
		ch := triggerOrderPrompt(g)
		if ch == nil {
			return
		}
		if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("the triggers never settled")
}

// Khârn and Selfless Squire. One instance deals damage to Khârn and to
// you: Khârn asks, and the Squire's announcement, owed behind Khârn's
// question in the same queue, waits for the answer. A second instruction
// dealt while the question is open owes a second announcement and waits
// too. After the answer each announcement is emitted once, in the order
// it was owed — the first instance's before the second instruction goes
// on — and the Squire gets one trigger for each.
func TestKharnHoldsSelflessSquiresAnnouncementsUntilTheAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := kharnPush(g, me.ID, me.ID)
	squire := castSquire(t, g)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	life := me.Life
	seen := -1
	g.WithWriteLock(func() {
		if err := g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToCreatureForEffect(src, kharn, 1); err != nil {
				return err
			}
			return g.DealDamageToPlayerForEffect(src, me.ID, 2)
		}); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToPlayerThenForEffect(src, me.ID, 3, func(g *game.Game, _ int) error {
			seen = squireTriggersLocked(g, squire)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
		t.Fatal("Khârn asked nothing")
	}
	if n := squireTriggers(g, squire); n != 0 || seen != -1 {
		t.Fatalf("%d Squire triggers and the second instruction saw %d while the question is open; want 0 and not run", n, seen)
	}
	kharnAnswer(t, g, me.ID, third.ID)
	if seen != 2 {
		t.Fatalf("the second instruction went on with %d Squire triggers; want 2 — both owed announcements first", seen)
	}
	if n := squireTriggers(g, squire); n != 2 {
		t.Fatalf("%d Squire triggers after the answer, want 2: one per application", n)
	}
	settleOrderingTriggers(t, g)
	if got := kharnController(t, g, kharn); got != third.ID {
		t.Fatalf("Khârn's controller %s, want %s", got, third.ID)
	}
	if n := countersOn(g, squire, game.CounterPlusOne); n != 5 || me.Life != life {
		t.Fatalf("Squire counters %d (want 5), life %d (want %d)", n, me.Life, life)
	}
	if n := squireTriggers(g, squire); n != 0 {
		t.Fatalf("%d Squire triggers left over: an announcement was emitted twice", n)
	}
}

// A table waiting on Khârn's question with a Squire announcement held
// behind it is a restore point, and the restored game emits the
// announcement once, after the answer.
func TestKharnHeldSquireAnnouncementSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := kharnPush(g, me.ID, me.ID)
	squire := castSquire(t, g)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	g.WithWriteLock(func() {
		if err := g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToCreatureForEffect(src, kharn, 1); err != nil {
				return err
			}
			return g.DealDamageToPlayerForEffect(src, me.ID, 2)
		}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	r := restoreRoundTrip(t, g, true)
	rme := r.Seats[0]
	if n := squireTriggers(r, squire); n != 0 {
		t.Fatalf("restored: %d Squire triggers before the answer, want 0", n)
	}
	kharnAnswer(t, r, rme.ID, third.ID)
	if n := squireTriggers(r, squire); n != 1 {
		t.Fatalf("restored: %d Squire triggers after the answer, want 1", n)
	}
	settleOrderingTriggers(t, r)
	if n := countersOn(r, squire, game.CounterPlusOne); n != 2 {
		t.Fatalf("restored: Squire counters %d, want 2", n)
	}
	if got := kharnController(t, r, kharn); got != third.ID {
		t.Fatalf("restored: Khârn's controller %s, want %s", got, third.ID)
	}
}

// Channel Harm and Selfless Squire: the one prevention is both Channel
// Harm's question and the Squire's trigger. The trigger waits behind the
// question and is emitted once, before the rest of the instruction.
func TestChannelHarmAndSelflessSquire(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squire := pushCatalogPermanent(g, me.ID, "Selfless Squire", "Creature — Human Soldier", selflessSquireOracle, false)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 6)
	castCatalogSpell(t, g, "Channel Harm", "Instant", channelHarmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	seen := -1
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerThenForEffect(src, me.ID, 3, func(g *game.Game, _ int) error {
			seen = squireTriggersLocked(g, squire)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil || seen != -1 || squireTriggers(g, squire) != 0 {
		t.Fatalf("want Channel Harm's question with the trigger and the instruction waiting; question %v, seen %d, triggers %d",
			ask, seen, squireTriggers(g, squire))
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if seen != 1 || damageMarkedOn(g, bear) != 3 {
		t.Fatalf("after yes: the instruction saw %d Squire triggers (want 1), bear damage %d (want 3)", seen, damageMarkedOn(g, bear))
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, squire, game.CounterPlusOne); n != 3 {
		t.Fatalf("Squire counters %d, want 3", n)
	}
}

// Channel Harm's yes meets Blood of the Martyr (#2056): the damage
// Channel Harm deals is offered for redirection, a second question asked
// only after the first is answered. The prevented instruction still waits
// until that one is answered too.
func TestChannelHarmsDamageMeetsBloodOfTheMartyr(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 6)
	castCatalogSpell(t, g, "Channel Harm", "Instant", channelHarmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Blood of the Martyr", "Instant", rcBloodOfTheMartyr, nil)
	passPriorityAroundTable(t, g)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	life := me.Life
	ran := false
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerThenForEffect(src, me.ID, 3, func(*game.Game, int) error {
			ran = true
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil || openPrompts(g) != 1 || ran {
		t.Fatalf("want only Channel Harm's question open and the instruction waiting; question %v, %d open, ran %v", ask, openPrompts(g), ran)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	may := rcMayPrompt(t, g, me.ID)
	if openPrompts(g) != 1 || ran || !g.ResolutionPaused() {
		t.Fatalf("want only the redirection open, holding the instruction; %d open, ran %v, paused %v", openPrompts(g), ran, g.ResolutionPaused())
	}
	if err := g.ResolveOptionalReplacement(may.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Fatal("the instruction did not go on after both answers")
	}
	if me.Life != life-3 || damageMarkedOn(g, bear) != 0 {
		t.Fatalf("redirected: life %d (want %d), bear %d (want 0)", me.Life, life-3, damageMarkedOn(g, bear))
	}
}

// Khârn and Blood of the Martyr on one event (#2056): both apply to the
// damage to Khârn, so its controller orders them (CR 616.1). With the
// redirection first and declined, Khârn's prevention applies and its
// question is asked; the instruction waits for that answer. Never two
// questions at once.
func TestKharnAndBloodOfTheMartyrOnOneEvent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := kharnPush(g, me.ID, me.ID)
	castCatalogSpell(t, g, "Blood of the Martyr", "Instant", rcBloodOfTheMartyr, nil)
	passPriorityAroundTable(t, g)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	life := me.Life
	seen := uuid.Nil
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureThenForEffect(src, kharn, 2, func(g *game.Game, _ int) error {
			c, _ := g.LookupCardForEffect(kharn)
			seen = c.Controller
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	order := latestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, me.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 || openPrompts(g) != 1 {
		t.Fatalf("want one CR 616 ordering prompt for Khârn's controller, got %+v (%d open)", order, openPrompts(g))
	}
	// The redirection first. Which ID is which is the engine's gather
	// order; the prompt that follows says which went first.
	ids := order.ReplacementEffectIDs
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{ids[1], ids[0]}); err != nil {
		t.Fatal(err)
	}
	may := latestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID)
	if may == nil {
		t.Fatal("the redirection did not go first: swap the order in this test")
	}
	if openPrompts(g) != 1 || seen != uuid.Nil {
		t.Fatalf("%d prompts open, continuation saw %s: want only the redirection, instruction waiting", openPrompts(g), seen)
	}
	if err := g.ResolveOptionalReplacement(may.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil || openPrompts(g) != 1 {
		t.Fatalf("want Khârn's question alone after the decline; %d open", openPrompts(g))
	}
	if seen != uuid.Nil {
		t.Fatal("the rest of the instruction ran before Khârn's question was answered")
	}
	kharnAnswer(t, g, me.ID, third.ID)
	if seen != third.ID {
		t.Fatalf("the continuation saw controller %s, want %s", seen, third.ID)
	}
	if me.Life != life || damageMarkedOn(g, kharn) != 0 {
		t.Fatalf("declined and prevented: life %d (want %d), Khârn damage %d (want 0)", me.Life, life, damageMarkedOn(g, kharn))
	}
}
