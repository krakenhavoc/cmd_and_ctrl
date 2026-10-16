package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevention_choice_now_test.go — #2040 (ADR 0108 amendment of
// 2026-10-16): a prevention effect's additional effect that asks a
// question is answered as the damage is prevented. The rest of the
// damage instruction, the next follow-up, a modal spell's next bullet and
// the state-based actions after combat damage all wait for the answer,
// and a table waiting on it is still a restore point.

const kharnOracle = "3246eff5-8739-469f-a401-e5de927dd6bb"

// kharnPush puts Khârn the Betrayer (5/1) onto the battlefield under
// `controller`, sick-free.
func kharnPush(g *game.Game, owner, controller uuid.UUID) uuid.UUID {
	return apaPush(g, owner, controller, game.Card{Name: "Khârn the Betrayer",
		TypeLine: "Legendary Creature — Astartes Berserker", OracleID: kharnOracle, Power: 5, Toughness: 1})
}

func kharnController(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatal("Khârn is not on the battlefield")
	}
	return c.Controller
}

// kharnAnswer has `chooser` answer the open "choose an opponent" prompt
// with `pick`, and fails when the prompt is not theirs.
func kharnAnswer(t *testing.T, g *game.Game, chooser, pick uuid.UUID) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, chooser)
	if c == nil {
		t.Fatalf("no option pick open for %s", chooser)
	}
	for i, o := range c.PickOptions {
		if o.Player == pick {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatalf("%s is not offered: %+v", pick, c.PickOptions)
}

// The control change happens before the rest of the damage instruction:
// a ...ThenForEffect continuation reads Khârn's controller after the
// answer, not before it (the ruling: "happens immediately, even if you
// are in the middle of resolving the spell or ability").
func TestKharnTheBetrayerChangesControlBeforeTheRestOfTheInstruction(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := kharnPush(g, me.ID, me.ID)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	var seen uuid.UUID
	dealt := -1
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureThenForEffect(src, kharn, 3, func(g *game.Game, n int) error {
			c, _ := g.LookupCardForEffect(kharn)
			seen, dealt = c.Controller, n
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if seen != uuid.Nil {
		t.Fatal("the rest of the instruction ran before the opponent was chosen")
	}
	if !g.ResolutionPaused() {
		t.Fatal("the question is not holding the boundary")
	}
	kharnAnswer(t, g, me.ID, third.ID)
	if seen != third.ID || dealt != 0 {
		t.Fatalf("the continuation saw controller %s and %d dealt; want %s and 0", seen, dealt, third.ID)
	}
	if got := kharnController(t, g, kharn); got != third.ID || damageMarkedOn(g, kharn) != 0 {
		t.Fatalf("controller %s, damage %d; want %s, 0", got, damageMarkedOn(g, kharn), third.ID)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("Sigil of Corruption: hand %d, want %d — the player who lost control draws two", me.Hand.Size(), hand+2)
	}
}

// Damaged twice in one instruction, Khârn changes hands twice, and the
// second question is asked of the player the first answer gave it to.
func TestKharnTheBetrayerSecondDamageAsksTheNewController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	kharn := kharnPush(g, me.ID, me.ID)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureThenForEffect(src, kharn, 1, func(g *game.Game, _ int) error {
			return g.DealDamageToCreatureForEffect(src, kharn, 1)
		}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	kharnAnswer(t, g, me.ID, third.ID)
	if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) != nil {
		t.Fatal("the second question went to the old controller")
	}
	kharnAnswer(t, g, third.ID, fourth.ID)
	if got := kharnController(t, g, kharn); got != fourth.ID {
		t.Fatalf("controller %s, want %s", got, fourth.ID)
	}
	myHand, thirdHand := me.Hand.Size(), third.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != myHand+2 || third.Hand.Size() != thirdHand+2 {
		t.Fatalf("each player who lost control draws two: me %d (want %d), third %d (want %d)",
			me.Hand.Size(), myHand+2, third.Hand.Size(), thirdHand+2)
	}
}

// Two damage instructions written as plain statements: the second runs
// while the first question is still open, so its follow-up waits in the
// queue and is asked, after the answer, of the new controller.
func TestKharnTheBetrayerHeldFollowUpAsksTheNewController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	kharn := kharnPush(g, me.ID, me.ID)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	g.WithWriteLock(func() {
		for range 2 {
			if err := g.DealDamageToCreatureForEffect(src, kharn, 1); err != nil {
				t.Fatal(err)
			}
		}
	})
	g.RunStateChecksForTest()
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d questions open, want 1: the second waits for the first", n)
	}
	kharnAnswer(t, g, me.ID, third.ID)
	kharnAnswer(t, g, third.ID, fourth.ID)
	if got := kharnController(t, g, kharn); got != fourth.ID {
		t.Fatalf("controller %s, want %s", got, fourth.ID)
	}
}

// Fiery Confluence, "1 damage to each creature" three times: Khârn
// changes hands three times, each question asked of whoever controls it
// then, and the next bullet waits for each answer (the ruling).
func TestKharnTheBetrayerAgainstFieryConfluence(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	kharn := kharnPush(g, opp.ID, opp.ID)
	ox := b12Creature(g, fourth.ID, "Their Ox", "Creature — Ox", 3, 3)
	castModal(t, g, "Fiery Confluence", "Sorcery", fieryConfluenceOracle, []int{0, 0, 0}, nil)
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, ox) != 1 {
		t.Fatalf("the ox has %d damage after the first question; want 1 — the next bullet waits", damageMarkedOn(g, ox))
	}
	kharnAnswer(t, g, opp.ID, me.ID)
	if damageMarkedOn(g, ox) != 2 {
		t.Fatalf("the ox has %d damage after the second question; want 2", damageMarkedOn(g, ox))
	}
	kharnAnswer(t, g, me.ID, third.ID)
	kharnAnswer(t, g, third.ID, fourth.ID)
	passPriorityAroundTable(t, g)
	if got := kharnController(t, g, kharn); got != fourth.ID || damageMarkedOn(g, kharn) != 0 {
		t.Fatalf("controller %s, damage %d; want %s, 0", got, damageMarkedOn(g, kharn), fourth.ID)
	}
	if g.Battlefield.Contains(ox) {
		t.Fatal("three bullets of 1 damage should have killed the ox")
	}
}

// In combat the question is answered after all the step's damage and
// before the state-based actions (CR 510.2, 704.3): the creature Khârn
// dealt lethal damage to is still on the battlefield while the question
// is open.
func TestKharnTheBetrayerCombatQuestionComesBeforeStateBasedActions(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := kharnPush(g, me.ID, me.ID)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 2, 2)
	pr7aBlockedCombat(t, g, opp.ID, kharn, blocker)
	advanceTo(t, g, game.StepCombatDamage)
	if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
		t.Fatal("no question after combat damage")
	}
	if !g.Battlefield.Contains(blocker) {
		t.Fatal("the state-based actions ran before the question was answered")
	}
	kharnAnswer(t, g, me.ID, third.ID)
	if g.Battlefield.Contains(blocker) {
		t.Fatal("the blocker survived lethal damage after the answer")
	}
	if got := kharnController(t, g, kharn); got != third.ID {
		t.Fatalf("controller %s, want %s", got, third.ID)
	}
}

// A table waiting on Khârn's question is a restore point, and the
// restored game answers it and finishes the spell.
func TestKharnTheBetrayerQuestionSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := kharnPush(g, opp.ID, opp.ID)
	ox := b12Creature(g, third.ID, "Their Ox", "Creature — Ox", 4, 4)
	castModal(t, g, "Fiery Confluence", "Sorcery", fieryConfluenceOracle, []int{0, 0, 0}, nil)
	passPriorityAroundTable(t, g)
	r := restoreRoundTrip(t, g, true)
	kharnAnswer(t, r, opp.ID, me.ID)
	if got := kharnController(t, r, kharn); got != me.ID {
		t.Fatalf("restored: controller %s, want %s", got, me.ID)
	}
	r2 := restoreRoundTrip(t, r, true)
	kharnAnswer(t, r2, me.ID, third.ID)
	kharnAnswer(t, r2, third.ID, opp.ID)
	passPriorityAroundTable(t, r2)
	if got := kharnController(t, r2, kharn); got != opp.ID || damageMarkedOn(r2, ox) != 3 {
		t.Fatalf("restored: controller %s, ox damage %d; want %s, 3", got, damageMarkedOn(r2, ox), opp.ID)
	}
}

// "You lose control" includes leaving the battlefield (the ruling).
func TestKharnTheBetrayerSigilWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kharn := kharnPush(g, me.ID, me.ID)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(kharn); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

// Channel Harm: damage from a source you don't control is prevented, and
// you are asked at once whether Channel Harm deals that much to its
// target; declining deals nothing, and your own source's damage is not
// prevented.
func TestChannelHarmAsksAsTheDamageIsPrevented(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 6)
	castCatalogSpell(t, g, "Channel Harm", "Instant", "4b3d1cfb-9f1d-4e4e-a15a-2e248ff5bdfc",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	life := me.Life
	after := -1
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerThenForEffect(src, me.ID, 3, func(g *game.Game, _ int) error {
			after = findBattlefieldCardForTest(g, bear).DamageMarked
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil || me.Life != life || after != -1 {
		t.Fatalf("want the damage prevented, a question and the instruction waiting; life %d (want %d), question %v, continuation saw %d",
			me.Life, life, ask, after)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if damageMarkedOn(g, bear) != 3 || after != 3 {
		t.Fatalf("bear damage %d, seen by the rest of the instruction %d; want 3 and 3", damageMarkedOn(g, bear), after)
	}
	pr7Hit(t, g, src, me.ID, 2)
	ask = latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil {
		t.Fatal("no question for the second instance")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if damageMarkedOn(g, bear) != 3 || me.Life != life {
		t.Fatalf("declined: bear %d (want 3), life %d (want %d)", damageMarkedOn(g, bear), me.Life, life)
	}
	mine := pr7Creature(g, me.ID, "Mine", 1, "W")
	pr7Hit(t, g, mine, me.ID, 1)
	if me.Life != life-1 {
		t.Fatalf("my own source: life %d, want %d", me.Life, life-1)
	}
}

// Immortal Coil's choice of cards is made at once, and a table waiting on
// it is a restore point.
func TestImmortalCoilChoiceSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Immortal Coil", "Artifact", pr8ImmortalCoil, false)
	var gy []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four"} {
		gy = append(gy, pushGraveyardCardForTest(me, n))
	}
	src := pr7Creature(g, opp.ID, "Src", 2, "R")
	pr7Hit(t, g, src, me.ID, 2)
	r := restoreRoundTrip(t, g, true)
	rme := r.Seats[0]
	pick := latestChoiceOfKindFor(r, game.PendingChoiceChooseCards, rme.ID)
	if pick == nil || pick.ChooseMin != 2 {
		t.Fatalf("restored: want the choose-two prompt, got %+v", pick)
	}
	if err := r.ResolveChooseCards(pick.ID, rme.ID, []uuid.UUID{gy[0], gy[2]}); err != nil {
		t.Fatal(err)
	}
	if rme.Graveyard.Size() != 2 || !rme.Graveyard.Contains(gy[1]) || !rme.Graveyard.Contains(gy[3]) {
		t.Fatalf("restored graveyard after the choice: %d cards, want Two and Four", rme.Graveyard.Size())
	}
}
