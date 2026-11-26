package testss82m

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b114_test.go — Ikoria: Lair of Behemoths, older-sets batch 114
// (#5512). Every test plays the real catalog entry through the engine.

const (
	oldB114HelicaGlider          = "8bcdb107-3827-4280-9e5e-7f1e771ef0c5"
	oldB114CallOfTheDeathDweller = "97773064-6534-4761-8c73-3ffd713f4569"
	oldB114EssenceSymbiote       = "9875cb1b-d67e-4ac4-8c16-164f28d4e4d4"
	oldB114Pyroceratops          = "b5306bf4-0cba-48dd-8af8-7c7821db7e02"
	oldB114HuntedNightmare       = "b5aae2f9-ced6-41d3-9121-418831ef5b8f"
	oldB114DuskfangMentor        = "d8b770c2-0406-4a2d-95d9-e14bb5d3bdc6"
	oldB114WillOfTheAllHunter    = "dec281b4-e6d6-4b78-b755-358e07eb7c06"
	oldB114BootNipper            = "e6337404-2978-47a8-a5ae-4a70add2df69"
	oldB114HornbashMentor        = "ec3031ff-93db-48e9-9e09-335786cb85c2"
	oldB114WingfoldPteron        = "f7db3bd9-ef67-4d64-a503-08a3c03404b2"
)

func TestOldB114EveryCardIsRegisteredFull(t *testing.T) {
	effectstest.NewCatalogGame(t)
	for _, id := range []string{oldB114HelicaGlider, oldB114CallOfTheDeathDweller, oldB114EssenceSymbiote, oldB114Pyroceratops, oldB114HuntedNightmare, oldB114DuskfangMentor, oldB114WillOfTheAllHunter, oldB114BootNipper, oldB114HornbashMentor, oldB114WingfoldPteron} {
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

func TestOldB114EntersWithYourChoiceOfCounter(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
		answers                []string
	}{
		{"Helica Glider", "Creature — Nightmare Squirrel", oldB114HelicaGlider, 2, 2, []string{game.CounterFlying, game.CounterFirstStrike}},
		{"Boot Nipper", "Creature — Beast", oldB114BootNipper, 2, 1, []string{game.CounterDeathtouch, game.CounterLifelink}},
		{"Wingfold Pteron", "Creature — Dinosaur", oldB114WingfoldPteron, 3, 6, []string{game.CounterFlying, game.CounterHexproof}},
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
				for _, kind := range tc.answers {
					wanted := 0
					if kind == want {
						wanted = 1
					}
					if got := effectstest.CounterOn(g, id, kind); got != wanted {
						t.Errorf("answer %d: %d %s counters, want %d", i, got, kind, wanted)
					}
				}
				if !effectstest.HasEffectiveKeyword(t, g, id, want) {
					t.Errorf("answer %d: no %s", i, want)
				}
			})
		}
	}
}

func TestOldB114MentorsPutAKeywordCounterOnANonHuman(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, kind string
		mana                         []string
	}{
		{"Duskfang Mentor", "Creature — Human Cleric", oldB114DuskfangMentor, game.CounterLifelink, []string{"C", "B"}},
		{"Hornbash Mentor", "Creature — Human Warrior", oldB114HornbashMentor, game.CounterTrample, []string{"C", "C", "G"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
			human := effectstest.B12Push(g, me.ID, "Soldier", "Creature — Human Soldier", "", 2, 2)
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 1)
			effectstest.AnswerAnyPendingTargetPrompts(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.CounterOn(g, bear, tc.kind) != 1 || effectstest.CounterOn(g, human, tc.kind) != 0 {
				t.Fatalf("the Bear has %d %s counters and the Human %d, want 1 and 0",
					effectstest.CounterOn(g, bear, tc.kind), tc.kind, effectstest.CounterOn(g, human, tc.kind))
			}
			var mentor uuid.UUID
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].Name == tc.name {
						mentor = g.Battlefield.Cards[i].InstanceID
						g.Battlefield.Cards[i].SummonedThisTurn = false
					}
				}
			})
			me.ManaPool.EmptyPool()
			effectstest.ApaMana(me, tc.mana...)
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

func TestOldB114PyroceratopsGrowsOnNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	pyro := effectstest.PushCatalogPermanent(g, me.ID, "Pyroceratops", "Creature — Elemental Dinosaur", oldB114Pyroceratops, false)
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, pyro, game.CounterPlusOne); got != 0 {
		t.Fatalf("a creature spell gave %d counters", got)
	}
	effectstest.CastCatalogSpell(t, g, "Will of the All-Hunter", "Instant", oldB114WillOfTheAllHunter, effectstest.CardTarget(pyro))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, pyro, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after a noncreature spell, want 1", got)
	}
}

func TestOldB114EssenceSymbioteGrowsAMutatedCreatureAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Essence Symbiote", "Creature — Beast", oldB114EssenceSymbiote, false)
	host := effectstest.B12Push(g, me.ID, "Host", "Creature — Bear", "", 2, 2)
	life := me.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventMutate, CardID: host, Actor: me.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, host, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the mutated creature, want 1", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	// An opponent's creature mutating does nothing.
	theirs := effectstest.B12Push(g, g.Seats[1].ID, "Theirs", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventMutate, CardID: theirs, Actor: g.Seats[1].ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, theirs, game.CounterPlusOne) != 0 || me.Life != life+2 {
		t.Error("an opponent's mutating creature triggered the Symbiote")
	}
}

func TestOldB114HuntedNightmareHasTheOpponentCounterTheirOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	a := effectstest.B12Push(g, opp.ID, "A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, opp.ID, "B", "Creature — Bear", "", 2, 2)
	effectstest.CastHoldCreature(t, g, "Hunted Nightmare", "Creature — Nightmare", oldB114HuntedNightmare, 4, 5)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the opponent was not asked (me %s): %+v", me.ID, g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, b, game.CounterDeathtouch) != 1 || effectstest.CounterOn(g, a, game.CounterDeathtouch) != 0 {
		t.Errorf("deathtouch counters: A %d, B %d, want 0 and 1", effectstest.CounterOn(g, a, game.CounterDeathtouch), effectstest.CounterOn(g, b, game.CounterDeathtouch))
	}
}

func TestOldB114HuntedNightmareWithOneCreatureAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := g.Seats[1]
	a := effectstest.B12Push(g, opp.ID, "A", "Creature — Bear", "", 2, 2)
	effectstest.CastHoldCreature(t, g, "Hunted Nightmare", "Creature — Nightmare", oldB114HuntedNightmare, 4, 5)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, a, game.CounterDeathtouch) != 1 {
		t.Error("the only creature did not get the counter")
	}
}

func TestOldB114WillOfTheAllHunterPumpsOrCountersABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Will of the All-Hunter", "Instant", oldB114WillOfTheAllHunter, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("not blocking: %d/%d, want 4/4", p, tt)
	}
	if effectstest.CounterOn(g, bear, game.CounterPlusOne) != 0 {
		t.Error("a creature that is not blocking got counters")
	}
	// A blocker gets two counters instead.
	blocker := effectstest.B12Push(g, me.ID, "Blocker", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == blocker {
				g.Battlefield.Cards[i].BlockingTarget = uuid.New()
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Will of the All-Hunter", "Instant", oldB114WillOfTheAllHunter, effectstest.CardTarget(blocker))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, blocker, game.CounterPlusOne) != 2 {
		t.Errorf("%d +1/+1 counters on the blocker, want 2", effectstest.CounterOn(g, blocker, game.CounterPlusOne))
	}
	if p, tt := effectstest.PtOf(t, g, blocker); p != 4 || tt != 4 {
		t.Errorf("blocker is %d/%d, want 4/4 from counters alone", p, tt)
	}
}

func TestOldB114WillOfTheAllHunterCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g, "Will of the All-Hunter", "Instant", oldB114WillOfTheAllHunter, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the draw", hand, me.Hand.Size())
	}
}

func oldB114Dead(g *game.Game, me *game.Player, name string, mv string) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: mv, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	return id
}

func TestOldB114CallOfTheDeathDwellerReturnsTwoAndCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	a := oldB114Dead(g, me, "A", "{1}{G}")
	b := oldB114Dead(g, me, "B", "{G}")
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	effectstest.CastCatalogSpell(t, g, "Call of the Death-Dweller", "Sorcery", oldB114CallOfTheDeathDweller, targets)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("both creatures did not return")
	}
	// Deathtouch on A, then menace on A too (either of them, the same one).
	pick := func(id uuid.UUID) {
		c := effectstest.LatestChooseCardsFor(g, me.ID)
		if c == nil {
			t.Fatal("no creature was asked for")
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{id}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	pick(a)
	pick(a)
	if effectstest.CounterOn(g, a, game.CounterDeathtouch) != 1 || effectstest.CounterOn(g, a, game.CounterMenace) != 1 {
		t.Error("A did not get both counters")
	}
	if effectstest.CounterOn(g, b, game.CounterDeathtouch) != 0 || effectstest.CounterOn(g, b, game.CounterMenace) != 0 {
		t.Error("B got a counter")
	}
}

func TestOldB114CallOfTheDeathDwellerWithOneCreatureAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	a := oldB114Dead(g, me, "A", "{2}")
	effectstest.CastCatalogSpell(t, g, "Call of the Death-Dweller", "Sorcery", oldB114CallOfTheDeathDweller, effectstest.CardTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, a, game.CounterDeathtouch) != 1 || effectstest.CounterOn(g, a, game.CounterMenace) != 1 {
		t.Error("the only creature did not get both counters")
	}
}

func TestOldB114CallOfTheDeathDwellerRefusesATotalOverThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	a := oldB114Dead(g, me, "A", "{2}{G}")
	b := oldB114Dead(g, me, "B", "{G}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Call of the Death-Dweller", TypeLine: "Sorcery", OracleID: oldB114CallOfTheDeathDweller, Owner: me.ID, Controller: me.ID})
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
		t.Error("creature cards with total mana value 4 were a legal pair")
	}
}
