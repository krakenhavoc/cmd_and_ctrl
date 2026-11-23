package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resetKharnSeen clears what testKharnSeenTail heard.
func resetKharnSeen() { effectstest.TestKharnSeen, effectstest.TestKharnDealt = uuid.Nil, -1 }

// a ...KeyedForEffect continuation reads Khârn's controller after the
// answer, not before it (the ruling: "happens immediately, even if you
// are in the middle of resolving the spell or ability").
func TestKharnTheBetrayerChangesControlBeforeTheRestOfTheInstruction(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := effectstest.KharnPush(g, me.ID, me.ID)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	resetKharnSeen()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureKeyedForEffect(src, kharn, 3, game.KeyedTail{Ref: effectstest.TestKharnSeenTail}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if effectstest.TestKharnSeen != uuid.Nil {
		t.Fatal("the rest of the instruction ran before the opponent was chosen")
	}
	if !g.ResolutionPaused() {
		t.Fatal("the question is not holding the boundary")
	}
	effectstest.KharnAnswer(t, g, me.ID, third.ID)
	if effectstest.TestKharnSeen != third.ID || effectstest.TestKharnDealt != 0 {
		t.Fatalf("the continuation saw controller %s and %d dealt; want %s and 0", effectstest.TestKharnSeen, effectstest.TestKharnDealt, third.ID)
	}
	if got := effectstest.KharnController(t, g, kharn); got != third.ID || effectstest.DamageMarkedOn(g, kharn) != 0 {
		t.Fatalf("controller %s, damage %d; want %s, 0", got, effectstest.DamageMarkedOn(g, kharn), third.ID)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("Sigil of Corruption: hand %d, want %d — the player who lost control draws two", me.Hand.Size(), hand+2)
	}
}

// Damaged twice in one instruction, Khârn changes hands twice, and the
// second question is asked of the player the first answer gave it to.
func TestKharnTheBetrayerSecondDamageAsksTheNewController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	kharn := effectstest.KharnPush(g, me.ID, me.ID)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureKeyedForEffect(src, kharn, 1, game.KeyedTail{Ref: effectstest.TestDealOneMoreTail, Source: src}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.KharnAnswer(t, g, me.ID, third.ID)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) != nil {
		t.Fatal("the second question went to the old controller")
	}
	effectstest.KharnAnswer(t, g, third.ID, fourth.ID)
	if got := effectstest.KharnController(t, g, kharn); got != fourth.ID {
		t.Fatalf("controller %s, want %s", got, fourth.ID)
	}
	myHand, thirdHand := me.Hand.Size(), third.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != myHand+2 || third.Hand.Size() != thirdHand+2 {
		t.Fatalf("each player who lost control draws two: me %d (want %d), third %d (want %d)",
			me.Hand.Size(), myHand+2, third.Hand.Size(), thirdHand+2)
	}
}

// Two damage instructions written as plain statements: the second runs
// while the first question is still open, so its follow-up waits in the
// queue and is asked, after the answer, of the new controller.
func TestKharnTheBetrayerHeldFollowUpAsksTheNewController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	kharn := effectstest.KharnPush(g, me.ID, me.ID)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
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
	effectstest.KharnAnswer(t, g, me.ID, third.ID)
	effectstest.KharnAnswer(t, g, third.ID, fourth.ID)
	if got := effectstest.KharnController(t, g, kharn); got != fourth.ID {
		t.Fatalf("controller %s, want %s", got, fourth.ID)
	}
}

// Fiery Confluence, "1 damage to each creature" three times: Khârn
// changes hands three times, each question asked of whoever controls it
// then, and the next bullet waits for each answer (the ruling).
func TestKharnTheBetrayerAgainstFieryConfluence(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	kharn := effectstest.KharnPush(g, opp.ID, opp.ID)
	ox := effectstest.B12Creature(g, fourth.ID, "Their Ox", "Creature — Ox", 3, 3)
	effectstest.CastModal(t, g, "Fiery Confluence", "Sorcery", effectstest.FieryConfluenceOracle, []int{0, 0, 0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, ox) != 1 {
		t.Fatalf("the ox has %d damage after the first question; want 1 — the next bullet waits", effectstest.DamageMarkedOn(g, ox))
	}
	effectstest.KharnAnswer(t, g, opp.ID, me.ID)
	if effectstest.DamageMarkedOn(g, ox) != 2 {
		t.Fatalf("the ox has %d damage after the second question; want 2", effectstest.DamageMarkedOn(g, ox))
	}
	effectstest.KharnAnswer(t, g, me.ID, third.ID)
	effectstest.KharnAnswer(t, g, third.ID, fourth.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.KharnController(t, g, kharn); got != fourth.ID || effectstest.DamageMarkedOn(g, kharn) != 0 {
		t.Fatalf("controller %s, damage %d; want %s, 0", got, effectstest.DamageMarkedOn(g, kharn), fourth.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := effectstest.KharnPush(g, me.ID, me.ID)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 2, 2)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, kharn, blocker)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
		t.Fatal("no question after combat damage")
	}
	if !g.Battlefield.Contains(blocker) {
		t.Fatal("the state-based actions ran before the question was answered")
	}
	effectstest.KharnAnswer(t, g, me.ID, third.ID)
	if g.Battlefield.Contains(blocker) {
		t.Fatal("the blocker survived lethal damage after the answer")
	}
	if got := effectstest.KharnController(t, g, kharn); got != third.ID {
		t.Fatalf("controller %s, want %s", got, third.ID)
	}
}

// A table waiting on Khârn's question is a restore point, and the
// restored game answers it and finishes the spell.
func TestKharnTheBetrayerQuestionSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kharn := effectstest.KharnPush(g, opp.ID, opp.ID)
	ox := effectstest.B12Creature(g, third.ID, "Their Ox", "Creature — Ox", 4, 4)
	effectstest.CastModal(t, g, "Fiery Confluence", "Sorcery", effectstest.FieryConfluenceOracle, []int{0, 0, 0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	r := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.KharnAnswer(t, r, opp.ID, me.ID)
	if got := effectstest.KharnController(t, r, kharn); got != me.ID {
		t.Fatalf("restored: controller %s, want %s", got, me.ID)
	}
	r2 := effectstest.RestoreRoundTrip(t, r, true)
	effectstest.KharnAnswer(t, r2, me.ID, third.ID)
	effectstest.KharnAnswer(t, r2, third.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, r2)
	if got := effectstest.KharnController(t, r2, kharn); got != opp.ID || effectstest.DamageMarkedOn(r2, ox) != 3 {
		t.Fatalf("restored: controller %s, ox damage %d; want %s, 3", got, effectstest.DamageMarkedOn(r2, ox), opp.ID)
	}
}

// "You lose control" includes leaving the battlefield (the ruling).
func TestKharnTheBetrayerSigilWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kharn := effectstest.KharnPush(g, me.ID, me.ID)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(kharn); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

// Channel Harm: damage from a source you don't control is prevented, and
// you are asked at once whether Channel Harm deals that much to its
// target; declining deals nothing, and your own source's damage is not
// prevented.
func TestChannelHarmAsksAsTheDamageIsPrevented(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 6)
	effectstest.CastCatalogSpell(t, g, "Channel Harm", "Instant", "4b3d1cfb-9f1d-4e4e-a15a-2e248ff5bdfc",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	life := me.Life
	effectstest.TestBearDamageAfter = -1
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerKeyedForEffect(src, me.ID, 3, game.KeyedTail{Ref: effectstest.TestBearDamageTail, Carry: []uuid.UUID{bear}}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	ask := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil || me.Life != life || effectstest.TestBearDamageAfter != -1 {
		t.Fatalf("want the damage prevented, a question and the instruction waiting; life %d (want %d), question %v, continuation saw %d",
			me.Life, life, ask, effectstest.TestBearDamageAfter)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if effectstest.DamageMarkedOn(g, bear) != 3 || effectstest.TestBearDamageAfter != 3 {
		t.Fatalf("bear damage %d, seen by the rest of the instruction %d; want 3 and 3", effectstest.DamageMarkedOn(g, bear), effectstest.TestBearDamageAfter)
	}
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	ask = effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil {
		t.Fatal("no question for the second instance")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if effectstest.DamageMarkedOn(g, bear) != 3 || me.Life != life {
		t.Fatalf("declined: bear %d (want 3), life %d (want %d)", effectstest.DamageMarkedOn(g, bear), me.Life, life)
	}
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 1, "W")
	effectstest.Pr7Hit(t, g, mine, me.ID, 1)
	if me.Life != life-1 {
		t.Fatalf("my own source: life %d, want %d", me.Life, life-1)
	}
}

// Immortal Coil's choice of cards is made at once, and a table waiting on
// it is a restore point.
func TestImmortalCoilChoiceSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Immortal Coil", "Artifact", effectstest.Pr8ImmortalCoil, false)
	var gy []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four"} {
		gy = append(gy, effectstest.PushGraveyardCardForTest(me, n))
	}
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 2, "R")
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	r := effectstest.RestoreRoundTrip(t, g, true)
	rme := r.Seats[0]
	pick := effectstest.LatestChoiceOfKindFor(r, game.PendingChoiceChooseCards, rme.ID)
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
