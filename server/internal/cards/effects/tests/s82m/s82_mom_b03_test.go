package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b03_test.go — S82 MOM batch 3 (#5477): backup creatures that
// grant keywords only. Every test casts the real catalog card, aims its
// backup trigger, and reads the layered result.

type s82MomB03Card struct {
	name, oracle, typeLine string
	power, toughness       int
	counters               int
	keywords               []string // printed below backup
	extra                  []string // printed above backup, not granted
}

var s82MomB03Cards = []s82MomB03Card{
	{"Boon-Bringer Valkyrie", "8652d88f-edca-4235-b942-6e7467608de8", "Creature — Angel Warrior", 4, 4, 1, []string{"flying", "first strike", "lifelink"}, nil},
	{"Consuming Aetherborn", "4d472308-5bc7-4a3c-81b5-0a757f00d9e2", "Creature — Aetherborn Vampire", 2, 2, 1, []string{"lifelink"}, nil},
	{"Fearless Skald", "f055a941-8b83-464e-a57f-8f188dc69aaa", "Creature — Dwarf Berserker", 3, 2, 1, []string{"double strike"}, nil},
	{"Golden-Scale Aeronaut", "d388080d-5ad4-4e27-b361-340f410423a7", "Creature — Dwarf Pilot", 2, 3, 1, []string{"flying"}, nil},
	{"Redcap Heelslasher", "d2134b68-d73f-4090-aaec-b2ee5c9f9f04", "Creature — Goblin Rogue", 2, 3, 1, []string{"first strike"}, nil},
	{"Saiba Cryptomancer", "15f89b06-5025-4c9f-bf51-e25bc5c2dd6b", "Creature — Moonfolk Ninja", 0, 1, 1, []string{"hexproof"}, []string{"flash"}},
	{"Serpent-Blade Assailant", "66a131ae-dbb2-4e7b-b6ce-7695f47c8f87", "Creature — Elf Warrior", 2, 1, 1, []string{"deathtouch"}, nil},
	{"Sigiled Sentinel", "79c90a68-1c31-4e16-be53-1304132da80f", "Creature — Human Knight", 2, 2, 1, []string{"vigilance"}, nil},
	{"Cragsmasher Yeti", "57120486-0316-4591-abca-438ab8283d3f", "Creature — Yeti", 4, 2, 2, []string{"trample"}, nil},
	{"Gloomfang Mauler", "5437f560-7f61-4326-be8e-15d74bf4e6ec", "Creature — Nightmare", 5, 5, 2, []string{"menace"}, nil},
	{"Death-Greeter's Champion", "81abb451-5ee0-4f90-bacc-a8ff69726a27", "Creature — Human Warrior", 2, 1, 1, []string{"double strike"}, nil},
}

func TestS82MomB03CardsAreFullWithTheirKeywords(t *testing.T) {
	for _, c := range s82MomB03Cards {
		spec, ok := effects.Lookup(c.oracle)
		if !ok || spec.Name != c.name {
			t.Fatalf("Lookup(%s) = %q, %v; want %s", c.oracle, spec.Name, ok, c.name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v", c.name, spec.Completeness, spec.Caveats)
		}
		want := append(append([]string{}, c.extra...), c.keywords...)
		if strings.Join(spec.PrintedKeywords, ",") != strings.Join(want, ",") {
			t.Errorf("%s: printed keywords %v, want %v", c.name, spec.PrintedKeywords, want)
		}
	}
}

// Backup on another creature: the counters and every printed keyword land
// on it, the card keeps its own keywords, and the grant ends with the turn.
func TestS82MomB03BackupOnAnotherCreature(t *testing.T) {
	for _, c := range s82MomB03Cards {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			id := effectstest.CastHoldCreature(t, g, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != c.counters {
				t.Fatalf("the Bear has %d +1/+1 counters, want %d", n, c.counters)
			}
			if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 0 {
				t.Errorf("the backup creature has %d counters, want 0", n)
			}
			for _, kw := range c.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
					t.Errorf("the Bear did not gain %s", kw)
				}
			}
			for _, kw := range append(append([]string{}, c.extra...), c.keywords...) {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s lacks its own %s", c.name, kw)
				}
			}
			// Printed above backup, so not granted.
			for _, kw := range c.extra {
				if effectstest.HasEffectiveKeyword(t, g, bear, kw) {
					t.Errorf("the Bear gained %s, which is not printed below backup", kw)
				}
			}
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			for _, kw := range c.keywords {
				if effectstest.HasEffectiveKeyword(t, g, bear, kw) {
					t.Errorf("the Bear still has %s after the turn ended", kw)
				}
			}
			if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != c.counters {
				t.Errorf("the counters did not stay: %d, want %d", n, c.counters)
			}
		})
	}
}

// Backup on itself: the counters only (CR 702.165a "if that's another creature").
func TestS82MomB03BackupOnItselfPutsCountersOnly(t *testing.T) {
	for _, c := range s82MomB03Cards {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.CastHoldCreature(t, g, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.AnswerTriggerTargets(t, g, me.ID, id)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != c.counters {
				t.Fatalf("%s has %d counters, want %d", c.name, n, c.counters)
			}
			for _, kw := range c.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s lost its printed %s", c.name, kw)
				}
			}
		})
	}
}

// Backup may target an opponent's creature (it is "target creature").
func TestS82MomB03BackupMayTargetAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c := s82MomB03Cards[0]
	effectstest.CastHoldCreature(t, g, c.name, c.typeLine, c.oracle, c.power, c.toughness)
	effectstest.AnswerTriggerTargets(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, theirs, game.CounterPlusOne) != 1 || !effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("the opponent's creature should get the counter and the keywords")
	}
}

func TestS82MomB03CragsmasherYetiMountaincycles(t *testing.T) {
	s82MomB03Cycles(t, 8, "Mountain", "Island")
}

func TestS82MomB03GloomfangMaulerSwampcycles(t *testing.T) { s82MomB03Cycles(t, 9, "Swamp", "Island") }

func s82MomB03Cycles(t *testing.T, idx int, land, other string) {
	t.Helper()
	c := s82MomB03Cards[idx]
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	want := effectstest.PushLibraryCardForTest(me, game.Card{Name: land, TypeLine: "Basic Land — " + land})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: other, TypeLine: "Basic Land — " + other})
	cid, p := effectstest.CycleFromHand(t, g, c.name, c.typeLine, c.oracle, "{C}{C}")
	if !p.Graveyard.Contains(cid) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.SearchChoiceFor(g, p.ID)
	if ch == nil || len(ch.SearchCards) != 1 || ch.SearchCards[0] != want {
		t.Fatalf("want a search offering only the %s, got %+v", land, ch)
	}
	if err := g.ResolveSearchLibrary(ch.ID, p.ID, []uuid.UUID{want}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !p.Hand.Contains(want) {
		t.Errorf("the %s was not put into hand", land)
	}
}

// Death-Greeter's Champion dashed for {3}{R}: haste, then back to hand.
func TestS82MomB03DeathGreetersChampionDash(t *testing.T) {
	c := s82MomB03Cards[10]
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, c.name, c.typeLine, "{2}{R}", c.oracle, []string{"R"})
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = c.power, c.toughness
			}
		}
	})
	effectstest.FloatForTest(g, me, "CCCR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "dash"}); err != nil {
		t.Fatalf("dash cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("a dashed creature has haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, id) || !me.Hand.Contains(id) {
		t.Error("the dashed creature should return to its owner's hand at the end step")
	}
}

// Strict: a pool that cannot pay refuses the cast.
func TestS82MomB03CannotCastWithoutMana(t *testing.T) {
	c := s82MomB03Cards[3]
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, c.name, c.typeLine, "{4}{W}", c.oracle, []string{"W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("casting with an empty pool should be refused")
	}
}
