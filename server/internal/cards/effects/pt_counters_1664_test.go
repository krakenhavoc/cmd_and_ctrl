package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pt_counters_1664_test.go — #1664: every P/T counter kind changes
// power and toughness (CR 122.1a). One test per proof card, one
// refusal for Contagion, and the replacement checks: Hardened Scales
// ("+1/+1 counters") leaves a +1/+2 alone, Doubling Season ("one or
// more counters") doubles it.

const (
	contagionOracle        = "433e86ce-df04-4182-bb63-ad2607d67dbd"
	wallOfRootsOracle      = "3a21a6ae-b2f2-4f0c-acfd-5f3e8d63fd2f"
	dwarvenArmorerOracle   = "5bbd27b1-0afd-4d98-a73c-c348c8f08625"
	armorThrullOracle      = "35453c9e-e1ae-4fe9-926d-75724deb0555"
	lightningSerpentOracle = "d07b4bb6-0c8d-44ea-a5b3-eeb6e36f3631"
)

func TestPTCounters1664CardsRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		contagionOracle:        "Contagion",
		wallOfRootsOracle:      "Wall of Roots",
		dwarvenArmorerOracle:   "Dwarven Armorer",
		armorThrullOracle:      "Armor Thrull",
		lightningSerpentOracle: "Lightning Serpent",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// ptOf is a battlefield creature's current power and toughness.
func ptOf(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tt int
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				p, tt, found = c.PowerForComparison(), c.CurrentToughness(), true
			}
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return p, tt
}

// --- Contagion ------------------------------------------------------

// Pitched for 1 life and a black card, the two -2/-1 counters split
// one each: a 3/3 becomes a 1/2, and an X/1 dies to the toughness
// check.
func TestContagionPitchesAndShrinksWithMinusTwoMinusOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Big", "Creature — Beast", 3, 3)
	small := b12Creature(g, opp.ID, "Small", "Creature — Beast", 4, 1)
	pitch := handCardFull(me, "Their Black Card", "Sorcery", "{B}", "", []string{"B"})
	life := me.Life

	d1658Cast(t, g, "Contagion", "Instant", contagionOracle, game.CastSpellParams{
		AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch},
		Targets: cardRefs(big, small), Distribution: map[uuid.UUID]int{big: 1, small: 1},
	})
	if me.Hand.Contains(pitch) {
		t.Error("the pitched card should have left the hand")
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	passPriorityAroundTable(t, g)

	if n := e2Card(t, g, big).Counters["-2/-1"]; n != 1 {
		t.Fatalf("big: %d -2/-1 counters, want 1", n)
	}
	if p, tt := ptOf(t, g, big); p != 1 || tt != 2 {
		t.Errorf("a 3/3 with a -2/-1 counter is %d/%d, want 1/2", p, tt)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == small {
			t.Error("a 4/1 with a -2/-1 counter is a 2/0 and should have died")
		}
	}
}

// Both counters on one creature is a legal split of two among "one or
// two" targets.
func TestContagionBothCountersOnOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Big", "Creature — Beast", 5, 5)
	d1658Cast(t, g, "Contagion", "Instant", contagionOracle, game.CastSpellParams{
		Targets: cardRefs(big), Distribution: map[uuid.UUID]int{big: 2},
	})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, big); p != 1 || tt != 3 {
		t.Errorf("a 5/5 with two -2/-1 counters is %d/%d, want 1/3", p, tt)
	}
}

// The pitch needs a BLACK card: a red one is refused at announce, and
// the spell stays in hand.
func TestContagionRefusesANonBlackPitch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Big", "Creature — Beast", 3, 3)
	red := handCardFull(me, "Their Red Card", "Sorcery", "{R}", "", []string{"R"})
	life := me.Life
	err := d1658CastErr(t, g, "Contagion", "Instant", contagionOracle, game.CastSpellParams{
		AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{red},
		Targets: cardRefs(big), Distribution: map[uuid.UUID]int{big: 2},
	})
	if err == nil {
		t.Fatal("a red card paid Contagion's pitch")
	}
	if !me.Hand.Contains(red) {
		t.Error("the refused pitch card left the hand")
	}
	if me.Life != life {
		t.Errorf("a refused cast cost life: %d, want %d", me.Life, life)
	}
}

// A -2/-1 does not annihilate a +1/+1 (CR 704.5q names -1/-1 only).
func TestContagionCounterDoesNotAnnihilateAPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	d1658Cast(t, g, "Contagion", "Instant", contagionOracle, game.CastSpellParams{
		Targets: cardRefs(bear), Distribution: map[uuid.UUID]int{bear: 2},
	})
	passPriorityAroundTable(t, g)
	c := e2Card(t, g, bear)
	if c.Counters[game.CounterPlusOne] != 1 || c.Counters["-2/-1"] != 2 {
		t.Fatalf("counters %v, want one +1/+1 and two -2/-1", c.Counters)
	}
	if p, tt := ptOf(t, g, bear); p != -1 || tt != 1 {
		t.Errorf("a 2/2 with +1/+1 and two -2/-1 is %d/%d, want -1/1", p, tt)
	}
}

// --- Wall of Roots --------------------------------------------------

// Putting a -0/-1 counter on the Wall adds {G}, shrinks its toughness,
// works while tapped, and only once a turn.
func TestWallOfRootsPaysToughnessForGreenOnceATurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wall := b12Push(g, me.ID, "Wall of Roots", "Creature — Plant Wall", wallOfRootsOracle, 0, 5)
	advanceToMain(t, g)
	b16Tap(g, wall)

	before := len(me.ManaPool)
	if err := g.ActivateManaAbility(me.ID, wall, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := len(me.ManaPool) - before; got != 1 || me.ManaPool[len(me.ManaPool)-1].Color != "G" {
		t.Fatalf("pool %v, want one more {G}", me.ManaPool)
	}
	if n := counterCount(g, wall, "-0/-1"); n != 1 {
		t.Fatalf("%d -0/-1 counters, want 1", n)
	}
	if p, tt := ptOf(t, g, wall); p != 0 || tt != 4 {
		t.Errorf("the Wall is %d/%d, want 0/4", p, tt)
	}
	if err := g.ActivateManaAbility(me.ID, wall, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a second activation the same turn was allowed")
	}
	if n := counterCount(g, wall, "-0/-1"); n != 1 {
		t.Errorf("the refused activation still placed a counter: %d", n)
	}
}

// --- Dwarven Armorer ------------------------------------------------

// The controller picks +1/+0 as the ability resolves, and the target
// gets exactly that.
func TestDwarvenArmorerPutsTheChosenCounter(t *testing.T) {
	for _, tc := range []struct {
		index        int
		kind         string
		wantP, wantT int
	}{
		{0, "+0/+1", 2, 3},
		{1, "+1/+0", 3, 2},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			armorer := b12Push(g, me.ID, "Dwarven Armorer", "Creature — Dwarf", dwarvenArmorerOracle, 0, 2)
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			discard := handCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
			advanceToMain(t, g)
			floatForTest(g, me, "R")

			b16Activate(t, g, me.ID, armorer, 0, game.ActivateAbilityParams{
				Targets: cardRefs(bear), DiscardIDs: []uuid.UUID{discard},
			})
			if me.Hand.Contains(discard) {
				t.Error("the discard is part of the cost and should be paid at announce")
			}
			if c := settleUntilOptionPick(t, g, me.ID, bear); c == nil {
				t.Fatal("no counter choice was asked")
			}
			answerOptionPick(t, g, me.ID, tc.index)
			passPriorityAroundTable(t, g)
			if n := counterCount(g, bear, tc.kind); n != 1 {
				t.Fatalf("%d %s counters, want 1 (all: %v)", n, tc.kind, e2Card(t, g, bear).Counters)
			}
			if p, tt := ptOf(t, g, bear); p != tc.wantP || tt != tc.wantT {
				t.Errorf("bear is %d/%d, want %d/%d", p, tt, tc.wantP, tc.wantT)
			}
		})
	}
}

// --- Armor Thrull + Hardened Scales / Doubling Season --------------

// Hardened Scales says "+1/+1 counters": a +1/+2 counter is not one,
// so the target gets exactly one. Doubling Season says "one or more
// counters": it doubles.
func TestArmorThrullCounterAndTheCounterReplacements(t *testing.T) {
	for _, tc := range []struct {
		label        string
		oracle, name string
		wantN        int
	}{
		{"Hardened Scales leaves it alone", hardenedScalesOracle, "Hardened Scales", 1},
		{"Doubling Season doubles it", doublingSeasonOracle, "Doubling Season", 2},
	} {
		t.Run(tc.label, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedReplacementPermanent(g, tc.oracle, tc.name, me.ID)
			thrull := b12Push(g, me.ID, "Armor Thrull", "Creature — Thrull", armorThrullOracle, 1, 3)
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			advanceToMain(t, g)

			b16Activate(t, g, me.ID, thrull, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
			passPriorityAroundTable(t, g)
			if n := counterCount(g, bear, "+1/+2"); n != tc.wantN {
				t.Fatalf("%d +1/+2 counters, want %d (all: %v)", n, tc.wantN, e2Card(t, g, bear).Counters)
			}
			if n := counterCount(g, bear, game.CounterPlusOne); n != 0 {
				t.Errorf("%d +1/+1 counters appeared from nowhere", n)
			}
			if p, tt := ptOf(t, g, bear); p != 2+tc.wantN || tt != 2+2*tc.wantN {
				t.Errorf("bear is %d/%d, want %d/%d", p, tt, 2+tc.wantN, 2+2*tc.wantN)
			}
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == thrull {
					t.Error("the Thrull is sacrificed as a cost")
				}
			}
		})
	}
}

// --- Lightning Serpent ----------------------------------------------

// Cast for X=3 it enters with three +1/+0 counters and swings as a
// 5/1, then is sacrificed at the end step.
func TestLightningSerpentEntersWithXPowerCountersAndLeavesAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lightning Serpent", TypeLine: "Creature — Elemental Serpent",
		OracleID: lightningSerpentOracle, ManaCost: "{X}{R}", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast Lightning Serpent: %v", err)
	}
	passPriorityAroundTable(t, g)

	if n := counterCount(g, id, "+1/+0"); n != 3 {
		t.Fatalf("%d +1/+0 counters, want 3", n)
	}
	if p, tt := ptOf(t, g, id); p != 5 || tt != 1 {
		t.Errorf("the Serpent is %d/%d, want 5/1", p, tt)
	}
	abilities := effectiveAbilities(t, g, id)
	if !containsString(abilities, "haste") || !containsString(abilities, "trample") {
		t.Errorf("printed keywords missing: %v", abilities)
	}

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			t.Fatal("the Serpent should be sacrificed at the beginning of the end step")
		}
	}
}

// --- the other P/T readers ------------------------------------------

// The automatic proliferate pick reads a -2/-1 or -0/-1 as the harmful
// kind it is, and a +1/+0 as a helpful one.
func TestProliferatePickReadsEveryPTCounterKind(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myPower := pushCounterCreature(g, me.ID, "Mine +1/+0", "+1/+0", 1)
	myWall := pushCounterCreature(g, me.ID, "Mine -0/-1", "-0/-1", 1)
	theirContagion := pushCounterCreature(g, opp.ID, "Theirs -2/-1", "-2/-1", 1)
	theirPower := pushCounterCreature(g, opp.ID, "Theirs +1/+0", "+1/+0", 1)

	cards, _ := BeneficialProliferateChoice(g, me.ID)
	chosen := map[uuid.UUID]bool{}
	for _, id := range cards {
		chosen[id] = true
	}
	if !chosen[myPower] || !chosen[theirContagion] {
		t.Errorf("want my +1/+0 creature and their -2/-1 creature chosen: %v", cards)
	}
	if chosen[myWall] || chosen[theirPower] {
		t.Errorf("my -0/-1 creature and their +1/+0 creature must not be chosen: %v", cards)
	}
}

// A dies trigger reading "this creature's power" reads every P/T
// counter it had: a Nested Shambler with two +1/+0 counters was a
// 3/1 and makes three Squirrels.
func TestLastKnownPowerReadsEveryPTCounterKind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shambler := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nested Shambler", TypeLine: "Creature — Zombie",
		OracleID: b39NestedShamblerOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(shambler, "+1/+0", 2) })
	if p, _ := ptOf(t, g, shambler); p != 3 {
		t.Fatalf("setup: a 1/1 with two +1/+0 counters has power 3, got %d", p)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shambler) })
	passPriorityAroundTable(t, g)
	squirrels := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Squirrel" {
			squirrels++
		}
	}
	if squirrels != 3 {
		t.Errorf("%d Squirrels, want 3 for a last-known power of 3", squirrels)
	}
}
