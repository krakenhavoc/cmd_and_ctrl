package testss82m

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b113_test.go — Ikoria: Lair of Behemoths, older-sets batch 113
// (#5512). Every test plays the real catalog entry through the engine.

const (
	oldB113FlycatcherGiraffid = "0264b059-75cb-41e6-9e51-262ac22b4c60"
	oldB113UnbreakableBond    = "36cfdbd6-f971-4ee5-aa1b-fa3ddad5b4b1"
	oldB113FrillscareMentor   = "3c9d217e-d7f8-43df-9bb6-58d29f93e2c4"
	oldB113FlourishingFox     = "44add927-4e6b-408e-907f-6a665add8b16"
	oldB113FerociousTigorilla = "4a7fef1e-3f11-4350-86da-6f1315fedc53"
	oldB113Grimdancer         = "4afd88fa-f8bb-46d4-9ba7-09419e6f858c"
	oldB113KeensightMentor    = "500a17a0-8332-414c-9c3a-7157a7800887"
	oldB113MomentumRumbler    = "502230fb-3a7a-439c-adfa-f7820deacf87"
	oldB113Chevill            = "50863075-a30b-4238-8e31-5af32a39d886"
	oldB113OminousSeas        = "58ea9b0a-1bef-4ccb-acb7-f03c8daf3df3"
	oldB113BloodCurdle        = "8b50de69-121a-48c4-bd7c-c8858f35f6ed"
)

func TestOldB113EveryCardIsRegisteredFull(t *testing.T) {
	effectstest.NewCatalogGame(t)
	for _, id := range []string{
		oldB113FlycatcherGiraffid, oldB113UnbreakableBond, oldB113FrillscareMentor, oldB113FlourishingFox,
		oldB113FerociousTigorilla, oldB113Grimdancer, oldB113KeensightMentor, oldB113MomentumRumbler,
		oldB113Chevill, oldB113OminousSeas, oldB113BloodCurdle,
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// The creature is asked what it enters with, and the answer's counters are
// on it, with their keywords.
func TestOldB113EntersWithYourChoiceOfCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
		answers                [][]string
	}{
		{"Flycatcher Giraffid", "Creature — Antelope Lizard", oldB113FlycatcherGiraffid, 3, 5,
			[][]string{{game.CounterReach}, {game.CounterVigilance}}},
		{"Ferocious Tigorilla", "Creature — Cat Ape", oldB113FerociousTigorilla, 4, 3,
			[][]string{{game.CounterTrample}, {game.CounterMenace}}},
		{"Grimdancer", "Creature — Nightmare", oldB113Grimdancer, 3, 3,
			[][]string{{game.CounterMenace, game.CounterDeathtouch}, {game.CounterMenace, game.CounterLifelink}, {game.CounterDeathtouch, game.CounterLifelink}}},
	} {
		for i, want := range tc.answers {
			t.Run(fmt.Sprintf("%s/%d", tc.name, i), func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
				if effectstest.LatestOptionPickFor(g, me.ID) == nil {
					t.Fatal("the controller was not asked")
				}
				effectstest.AnswerOptionPick(t, g, me.ID, i)
				effectstest.PassPriorityAroundTable(t, g)
				all := []string{game.CounterReach, game.CounterVigilance, game.CounterTrample, game.CounterMenace, game.CounterDeathtouch, game.CounterLifelink}
				for _, kind := range all {
					wanted := 0
					for _, w := range want {
						if w == kind {
							wanted = 1
						}
					}
					if got := effectstest.CounterOn(g, id, kind); got != wanted {
						t.Errorf("answer %d: %d %s counters, want %d", i, got, kind, wanted)
					}
					if wanted == 1 && !effectstest.HasEffectiveKeyword(t, g, id, kind) {
						t.Errorf("answer %d: no %s", i, kind)
					}
				}
			})
		}
	}
}

func TestOldB113UnbreakableBondReturnsACreatureWithLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Unbreakable Bond", "Sorcery", oldB113UnbreakableBond, effectstest.CardTarget(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the Bear did not return")
	}
	if effectstest.CounterOn(g, dead, game.CounterLifelink) != 1 || !effectstest.HasEffectiveKeyword(t, g, dead, "lifelink") {
		t.Error("the Bear has no lifelink counter")
	}
}

func TestOldB113UnbreakableBondRefusesANonCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Unbreakable Bond", TypeLine: "Sorcery", OracleID: oldB113UnbreakableBond, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(land)}); err == nil {
		t.Error("a land card was a legal target")
	}
}

// The mentors put the counter on a non-Human creature only, and their tap
// ability counts every creature with the keyword, counters included.
func TestOldB113MentorsPutAKeywordCounterOnANonHuman(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, kind, mana string
	}{
		{"Frillscare Mentor", "Creature — Human Warrior", oldB113FrillscareMentor, game.CounterMenace, "R"},
		{"Keensight Mentor", "Creature — Human Cleric", oldB113KeensightMentor, game.CounterVigilance, "W"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
			human := effectstest.B12Push(g, me.ID, "Soldier", "Creature — Human Soldier", "", 2, 2)
			// A Human is not a legal target for the counter.
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 1)
			effectstest.AnswerAnyPendingTargetPrompts(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.CounterOn(g, bear, tc.kind) != 1 {
				t.Errorf("the Bear has %d %s counters, want 1", effectstest.CounterOn(g, bear, tc.kind), tc.kind)
			}
			if effectstest.CounterOn(g, human, tc.kind) != 0 {
				t.Errorf("the Human got a %s counter", tc.kind)
			}
			// The tap ability: the Bear (counter) grows, the Human does not.
			var mentor uuid.UUID
			g.ReadSnapshot(func() {
				for _, c := range g.Battlefield.Cards {
					if c.Name == tc.name {
						mentor = c.InstanceID
					}
				}
			})
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == mentor {
						g.Battlefield.Cards[i].SummonedThisTurn = false
					}
				}
			})
			effectstest.ApaMana(me, "C", "C", tc.mana)
			if tc.name == "Keensight Mentor" {
				me.ManaPool.EmptyPool()
				effectstest.ApaMana(me, "C", tc.mana)
			}
			effectstest.ApaActivate(t, g, me, mentor, 0, game.ActivateAbilityParams{})
			if effectstest.CounterOn(g, bear, game.CounterPlusOne) != 1 {
				t.Errorf("the Bear has %d +1/+1 counters, want 1", effectstest.CounterOn(g, bear, game.CounterPlusOne))
			}
			if effectstest.CounterOn(g, human, game.CounterPlusOne) != 0 || effectstest.CounterOn(g, mentor, game.CounterPlusOne) != 0 {
				t.Error("a creature without the keyword got a +1/+1 counter")
			}
		})
	}
}

func TestOldB113FlourishingFoxGrowsWhenYouCycleAnotherCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fox := effectstest.PushCatalogPermanent(g, me.ID, "Flourishing Fox", "Creature — Fox", oldB113FlourishingFox, false)
	effectstest.CycleFromHand(t, g, "Imposing Vantasaur", "Creature — Dinosaur", "a0d5fce0-a439-496a-9fcf-9271503f3ce8", "{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, fox, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after one cycle, want 1", got)
	}
}

func TestOldB113FlourishingFoxCyclesForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g, "Flourishing Fox", "Creature — Fox", oldB113FlourishingFox, "{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Fox is not in the graveyard")
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the draw", hand, me.Hand.Size())
	}
}

func TestOldB113MomentumRumblerGetsFirstStrikeThenDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rumbler := effectstest.PushCatalogPermanent(g, me.ID, "Momentum Rumbler", "Creature — Dinosaur", oldB113MomentumRumbler, false)
	attack := func() {
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(rumbler, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	attack()
	if effectstest.CounterOn(g, rumbler, game.CounterFirstStrike) != 1 {
		t.Fatalf("%d first strike counters after the first attack, want 1", effectstest.CounterOn(g, rumbler, game.CounterFirstStrike))
	}
	if effectstest.HasEffectiveKeyword(t, g, rumbler, "double strike") {
		t.Error("double strike on the first attack")
	}
	// Next turn it attacks again: it has first strike, so it gains double strike.
	effectstest.AdvanceToMainOf(t, g, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rumbler {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	attack()
	if !effectstest.HasEffectiveKeyword(t, g, rumbler, "double strike") {
		t.Error("no double strike on the second attack")
	}
	if effectstest.CounterOn(g, rumbler, game.CounterFirstStrike) != 1 {
		t.Error("a second first strike counter was put on it")
	}
}

func TestOldB113ChevillMarksAnOpponentsCreatureAndPaysOnItsDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Chevill, Bane of Monsters", "Legendary Creature — Human Rogue", oldB113Chevill, false)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, victim, "bounty"); got != 1 {
		t.Fatalf("%d bounty counters, want 1", got)
	}
	life, hand := me.Life, me.Hand.Size()
	effectstest.B27Kill(g, victim)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB113ChevillDoesNotMarkWhileAnOpponentsPermanentHasABounty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Chevill, Bane of Monsters", "Legendary Creature — Human Rogue", oldB113Chevill, false)
	marked := effectstest.B12Push(g, opp.ID, "Marked", "Creature — Bear", "", 2, 2)
	other := effectstest.B12Push(g, opp.ID, "Other", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(marked, "bounty", 1) })
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, other, "bounty") != 0 || effectstest.CounterOn(g, marked, "bounty") != 1 {
		t.Error("Chevill put a second bounty counter while one was out")
	}
}

func TestOldB113ChevillIgnoresYourOwnBountyPermanentDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Chevill, Bane of Monsters", "Legendary Creature — Human Rogue", oldB113Chevill, false)
	mine := effectstest.B12Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, "bounty", 1) })
	life := me.Life
	effectstest.B27Kill(g, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d: your own permanent paid out", me.Life, life)
	}
}

func TestOldB113OminousSeasCountsDrawsAndMakesAKraken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	seas := effectstest.PushCatalogPermanent(g, me.ID, "Ominous Seas", "Enchantment", oldB113OminousSeas, false)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, seas, "foreshadow"); got != 3 {
		t.Fatalf("%d foreshadow counters after drawing three, want 3", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, seas, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability was paid with three counters")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(seas, "foreshadow", 5) })
	effectstest.ApaActivate(t, g, me, seas, 0, game.ActivateAbilityParams{})
	if got := effectstest.CounterOn(g, seas, "foreshadow"); got != 0 {
		t.Errorf("%d foreshadow counters left, want 0", got)
	}
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Kraken" && c.Controller == me.ID && c.IsToken() {
				if p, tt := c.Power, c.Toughness; p != 8 || tt != 8 {
					t.Errorf("Kraken is %d/%d, want 8/8", p, tt)
				}
				n++
			}
		}
	})
	if n != 1 {
		t.Errorf("%d Kraken tokens, want 1", n)
	}
}

func TestOldB113OminousSeasCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.PushCatalogHandCard(me, "Ominous Seas", "Enchantment", oldB113OminousSeas)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
}

func TestOldB113BloodCurdleDestroysAndMarksACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
	mine := effectstest.B12Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blood Curdle", "Instant", oldB113BloodCurdle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target was not destroyed")
	}
	if effectstest.CounterOn(g, mine, game.CounterMenace) != 1 || !effectstest.HasEffectiveKeyword(t, g, mine, "menace") {
		t.Error("my creature has no menace counter")
	}
}

// Targeting your own only creature destroys it first, so there is none left
// to put the counter on.
func TestOldB113BloodCurdleOnYourOnlyCreatureLeavesNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mine := effectstest.B12Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blood Curdle", "Instant", oldB113BloodCurdle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the target was not destroyed")
	}
}
