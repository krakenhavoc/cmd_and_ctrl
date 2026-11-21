package testsp

import (
	"testing"

	"github.com/google/uuid"

	cardsch "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ch"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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

// Khârn and Selfless Squire. One instance deals damage to Khârn and to
// you: Khârn asks, and the Squire's announcement, owed behind Khârn's
// question in the same queue, waits for the answer. A second instruction
// dealt while the question is open owes a second announcement and waits
// too. After the answer each announcement is emitted once, in the order
// it was owed — the first instance's before the second instruction goes
// on — and the Squire gets one trigger for each.
func TestKharnHoldsSelflessSquiresAnnouncementsUntilTheAnswer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := effectstest.KharnPush(g, me.ID, me.ID)
	squire := effectstest.CastSquire(t, g)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToCreatureForEffect(src, kharn, 1); err != nil {
				return err
			}
			return g.DealDamageToPlayerForEffect(src, me.ID, 2)
		}); err != nil {
			t.Fatal(err)
		}
		effectstest.TestSquireSeen = -1
		if err := g.DealDamageToPlayerKeyedForEffect(src, me.ID, 3, game.KeyedTail{Ref: effectstest.TestSquireSeenTail, Carry: []uuid.UUID{squire}}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
		t.Fatal("Khârn asked nothing")
	}
	if n := effectstest.SquireTriggers(g, squire); n != 0 || effectstest.TestSquireSeen != -1 {
		t.Fatalf("%d Squire triggers and the second instruction saw %d while the question is open; want 0 and not run", n, effectstest.TestSquireSeen)
	}
	effectstest.KharnAnswer(t, g, me.ID, third.ID)
	if effectstest.TestSquireSeen != 2 {
		t.Fatalf("the second instruction went on with %d Squire triggers; want 2 — both owed announcements first", effectstest.TestSquireSeen)
	}
	if n := effectstest.SquireTriggers(g, squire); n != 2 {
		t.Fatalf("%d Squire triggers after the answer, want 2: one per application", n)
	}
	effectstest.SettleOrderingTriggers(t, g)
	if got := effectstest.KharnController(t, g, kharn); got != third.ID {
		t.Fatalf("Khârn's controller %s, want %s", got, third.ID)
	}
	if n := effectstest.CountersOn(g, squire, game.CounterPlusOne); n != 5 || me.Life != life {
		t.Fatalf("Squire counters %d (want 5), life %d (want %d)", n, me.Life, life)
	}
	if n := effectstest.SquireTriggers(g, squire); n != 0 {
		t.Fatalf("%d Squire triggers left over: an announcement was emitted twice", n)
	}
}

// A table waiting on Khârn's question with a Squire announcement held
// behind it is a restore point, and the restored game emits the
// announcement once, after the answer.
func TestKharnHeldSquireAnnouncementSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := effectstest.KharnPush(g, me.ID, me.ID)
	squire := effectstest.CastSquire(t, g)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
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
	r := effectstest.RestoreRoundTrip(t, g, true)
	rme := r.Seats[0]
	if n := effectstest.SquireTriggers(r, squire); n != 0 {
		t.Fatalf("restored: %d Squire triggers before the answer, want 0", n)
	}
	effectstest.KharnAnswer(t, r, rme.ID, third.ID)
	if n := effectstest.SquireTriggers(r, squire); n != 1 {
		t.Fatalf("restored: %d Squire triggers after the answer, want 1", n)
	}
	effectstest.SettleOrderingTriggers(t, r)
	if n := effectstest.CountersOn(r, squire, game.CounterPlusOne); n != 2 {
		t.Fatalf("restored: Squire counters %d, want 2", n)
	}
	if got := effectstest.KharnController(t, r, kharn); got != third.ID {
		t.Fatalf("restored: Khârn's controller %s, want %s", got, third.ID)
	}
}

// Channel Harm and Selfless Squire: the one prevention is both Channel
// Harm's question and the Squire's trigger. The trigger waits behind the
// question and is emitted once, before the rest of the instruction.
func TestChannelHarmAndSelflessSquire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squire := effectstest.PushCatalogPermanent(g, me.ID, "Selfless Squire", "Creature — Human Soldier", effectstest.SelflessSquireOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 6)
	effectstest.CastCatalogSpell(t, g, "Channel Harm", "Instant", cardsch.ChannelHarmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	g.WithWriteLock(func() {
		effectstest.TestSquireSeen = -1
		if err := g.DealDamageToPlayerKeyedForEffect(src, me.ID, 3, game.KeyedTail{Ref: effectstest.TestSquireSeenTail, Carry: []uuid.UUID{squire}}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	ask := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil || effectstest.TestSquireSeen != -1 || effectstest.SquireTriggers(g, squire) != 0 {
		t.Fatalf("want Channel Harm's question with the trigger and the instruction waiting; question %v, seen %d, triggers %d",
			ask, effectstest.TestSquireSeen, effectstest.SquireTriggers(g, squire))
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if effectstest.TestSquireSeen != 1 || effectstest.DamageMarkedOn(g, bear) != 3 {
		t.Fatalf("after yes: the instruction saw %d Squire triggers (want 1), bear damage %d (want 3)", effectstest.TestSquireSeen, effectstest.DamageMarkedOn(g, bear))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, squire, game.CounterPlusOne); n != 3 {
		t.Fatalf("Squire counters %d, want 3", n)
	}
}

// Channel Harm's yes meets Blood of the Martyr (#2056): the damage
// Channel Harm deals is offered for redirection, a second question asked
// only after the first is answered. The prevented instruction still waits
// until that one is answered too.
func TestChannelHarmsDamageMeetsBloodOfTheMartyr(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 6)
	effectstest.CastCatalogSpell(t, g, "Channel Harm", "Instant", cardsch.ChannelHarmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Blood of the Martyr", "Instant", effectstest.RcBloodOfTheMartyr, nil)
	effectstest.PassPriorityAroundTable(t, g)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	life := me.Life
	g.WithWriteLock(func() {
		effectstest.TestDamageTailRan = false
		if err := g.DealDamageToPlayerKeyedForEffect(src, me.ID, 3, game.KeyedTail{Ref: effectstest.TestDamageRanTail}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	ask := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil || openPrompts(g) != 1 || effectstest.TestDamageTailRan {
		t.Fatalf("want only Channel Harm's question open and the instruction waiting; question %v, %d open, ran %v", ask, openPrompts(g), effectstest.TestDamageTailRan)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	may := effectstest.RcMayPrompt(t, g, me.ID)
	if openPrompts(g) != 1 || effectstest.TestDamageTailRan || !g.ResolutionPaused() {
		t.Fatalf("want only the redirection open, holding the instruction; %d open, ran %v, paused %v", openPrompts(g), effectstest.TestDamageTailRan, g.ResolutionPaused())
	}
	if err := g.ResolveOptionalReplacement(may.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if !effectstest.TestDamageTailRan {
		t.Fatal("the instruction did not go on after both answers")
	}
	if me.Life != life-3 || effectstest.DamageMarkedOn(g, bear) != 0 {
		t.Fatalf("redirected: life %d (want %d), bear %d (want 0)", me.Life, life-3, effectstest.DamageMarkedOn(g, bear))
	}
}

// Khârn and Blood of the Martyr on one event (#2056): both apply to the
// damage to Khârn, so its controller orders them (CR 616.1). With the
// redirection first and declined, Khârn's prevention applies and its
// question is asked; the instruction waits for that answer. Never two
// questions at once.
func TestKharnAndBloodOfTheMartyrOnOneEvent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := effectstest.KharnPush(g, me.ID, me.ID)
	effectstest.CastCatalogSpell(t, g, "Blood of the Martyr", "Instant", effectstest.RcBloodOfTheMartyr, nil)
	effectstest.PassPriorityAroundTable(t, g)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	life := me.Life
	resetKharnSeen()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureKeyedForEffect(src, kharn, 2, game.KeyedTail{Ref: effectstest.TestKharnSeenTail}); err != nil {
			t.Fatal(err)
		}
	})
	order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, me.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 || openPrompts(g) != 1 {
		t.Fatalf("want one CR 616 ordering prompt for Khârn's controller, got %+v (%d open)", order, openPrompts(g))
	}
	// The redirection first. Which ID is which is the engine's gather
	// order; the prompt that follows says which went first.
	ids := order.ReplacementEffectIDs
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{ids[1], ids[0]}); err != nil {
		t.Fatal(err)
	}
	may := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID)
	if may == nil {
		t.Fatal("the redirection did not go first: swap the order in this test")
	}
	if openPrompts(g) != 1 || effectstest.TestKharnSeen != uuid.Nil {
		t.Fatalf("%d prompts open, continuation saw %s: want only the redirection, instruction waiting", openPrompts(g), effectstest.TestKharnSeen)
	}
	if err := g.ResolveOptionalReplacement(may.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil || openPrompts(g) != 1 {
		t.Fatalf("want Khârn's question alone after the decline; %d open", openPrompts(g))
	}
	if effectstest.TestKharnSeen != uuid.Nil {
		t.Fatal("the rest of the instruction ran before Khârn's question was answered")
	}
	effectstest.KharnAnswer(t, g, me.ID, third.ID)
	if effectstest.TestKharnSeen != third.ID {
		t.Fatalf("the continuation saw controller %s, want %s", effectstest.TestKharnSeen, third.ID)
	}
	if me.Life != life || effectstest.DamageMarkedOn(g, kharn) != 0 {
		t.Fatalf("declined and prevented: life %d (want %d), Khârn damage %d (want 0)", me.Life, life, effectstest.DamageMarkedOn(g, kharn))
	}
}
