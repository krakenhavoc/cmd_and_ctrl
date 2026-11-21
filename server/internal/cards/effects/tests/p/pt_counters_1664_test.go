package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPTCounters1664CardsRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.ContagionOracle:        "Contagion",
		effectstest.WallOfRootsOracle:      "Wall of Roots",
		effectstest.DwarvenArmorerOracle:   "Dwarven Armorer",
		effectstest.ArmorThrullOracle:      "Armor Thrull",
		effectstest.LightningSerpentOracle: "Lightning Serpent",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- Contagion ------------------------------------------------------

// Pitched for 1 life and a black card, the two -2/-1 counters split
// one each: a 3/3 becomes a 1/2, and an X/1 dies to the toughness
// check.
func TestContagionPitchesAndShrinksWithMinusTwoMinusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Beast", 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Beast", 4, 1)
	pitch := effectstest.HandCardFull(me, "Their Black Card", "Sorcery", "{B}", "", []string{"B"})
	life := me.Life

	effectstest.D1658Cast(t, g, "Contagion", "Instant", effectstest.ContagionOracle, game.CastSpellParams{
		AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch},
		Targets: effectstest.CardRefs(big, small), Distribution: map[uuid.UUID]int{big: 1, small: 1},
	})
	if me.Hand.Contains(pitch) {
		t.Error("the pitched card should have left the hand")
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.E2Card(t, g, big).Counters["-2/-1"]; n != 1 {
		t.Fatalf("big: %d -2/-1 counters, want 1", n)
	}
	if p, tt := effectstest.PtOf(t, g, big); p != 1 || tt != 2 {
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Beast", 5, 5)
	effectstest.D1658Cast(t, g, "Contagion", "Instant", effectstest.ContagionOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(big), Distribution: map[uuid.UUID]int{big: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, big); p != 1 || tt != 3 {
		t.Errorf("a 5/5 with two -2/-1 counters is %d/%d, want 1/3", p, tt)
	}
}

// The pitch needs a BLACK card: a red one is refused at announce, and
// the spell stays in hand.
func TestContagionRefusesANonBlackPitch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Beast", 3, 3)
	red := effectstest.HandCardFull(me, "Their Red Card", "Sorcery", "{R}", "", []string{"R"})
	life := me.Life
	err := effectstest.D1658CastErr(t, g, "Contagion", "Instant", effectstest.ContagionOracle, game.CastSpellParams{
		AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{red},
		Targets: effectstest.CardRefs(big), Distribution: map[uuid.UUID]int{big: 2},
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	effectstest.D1658Cast(t, g, "Contagion", "Instant", effectstest.ContagionOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(bear), Distribution: map[uuid.UUID]int{bear: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.E2Card(t, g, bear)
	if c.Counters[game.CounterPlusOne] != 1 || c.Counters["-2/-1"] != 2 {
		t.Fatalf("counters %v, want one +1/+1 and two -2/-1", c.Counters)
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != -1 || tt != 1 {
		t.Errorf("a 2/2 with +1/+1 and two -2/-1 is %d/%d, want -1/1", p, tt)
	}
}

// --- Wall of Roots --------------------------------------------------

// Putting a -0/-1 counter on the Wall adds {G}, shrinks its toughness,
// works while tapped, and only once a turn.
func TestWallOfRootsPaysToughnessForGreenOnceATurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wall := effectstest.B12Push(g, me.ID, "Wall of Roots", "Creature — Plant Wall", effectstest.WallOfRootsOracle, 0, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Tap(g, wall)

	before := me.ManaPool.Total()
	if err := g.ActivateManaAbility(me.ID, wall, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.Total() - before; got != 1 || me.ManaPool.Units()[me.ManaPool.Total()-1].Color != "G" {
		t.Fatalf("pool %v, want one more {G}", me.ManaPool)
	}
	if n := effectstest.CounterCount(g, wall, "-0/-1"); n != 1 {
		t.Fatalf("%d -0/-1 counters, want 1", n)
	}
	if p, tt := effectstest.PtOf(t, g, wall); p != 0 || tt != 4 {
		t.Errorf("the Wall is %d/%d, want 0/4", p, tt)
	}
	if err := g.ActivateManaAbility(me.ID, wall, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a second activation the same turn was allowed")
	}
	if n := effectstest.CounterCount(g, wall, "-0/-1"); n != 1 {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			armorer := effectstest.B12Push(g, me.ID, "Dwarven Armorer", "Creature — Dwarf", effectstest.DwarvenArmorerOracle, 0, 2)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			discard := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, "R")

			effectstest.B16Activate(t, g, me.ID, armorer, 0, game.ActivateAbilityParams{
				Targets: effectstest.CardRefs(bear), DiscardIDs: []uuid.UUID{discard},
			})
			if me.Hand.Contains(discard) {
				t.Error("the discard is part of the cost and should be paid at announce")
			}
			if c := effectstest.SettleUntilOptionPick(t, g, me.ID, bear); c == nil {
				t.Fatal("no counter choice was asked")
			}
			effectstest.AnswerOptionPick(t, g, me.ID, tc.index)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CounterCount(g, bear, tc.kind); n != 1 {
				t.Fatalf("%d %s counters, want 1 (all: %v)", n, tc.kind, effectstest.E2Card(t, g, bear).Counters)
			}
			if p, tt := effectstest.PtOf(t, g, bear); p != tc.wantP || tt != tc.wantT {
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
		{"Hardened Scales leaves it alone", effectstest.HardenedScalesOracle, "Hardened Scales", 1},
		{"Doubling Season doubles it", effectstest.DoublingSeasonOracle, "Doubling Season", 2},
	} {
		t.Run(tc.label, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedReplacementPermanent(g, tc.oracle, tc.name, me.ID)
			thrull := effectstest.B12Push(g, me.ID, "Armor Thrull", "Creature — Thrull", effectstest.ArmorThrullOracle, 1, 3)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			effectstest.AdvanceToMain(t, g)

			effectstest.B16Activate(t, g, me.ID, thrull, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CounterCount(g, bear, "+1/+2"); n != tc.wantN {
				t.Fatalf("%d +1/+2 counters, want %d (all: %v)", n, tc.wantN, effectstest.E2Card(t, g, bear).Counters)
			}
			if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 0 {
				t.Errorf("%d +1/+1 counters appeared from nowhere", n)
			}
			if p, tt := effectstest.PtOf(t, g, bear); p != 2+tc.wantN || tt != 2+2*tc.wantN {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lightning Serpent", TypeLine: "Creature — Elemental Serpent",
		OracleID: effectstest.LightningSerpentOracle, ManaCost: "{X}{R}", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast Lightning Serpent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CounterCount(g, id, "+1/+0"); n != 3 {
		t.Fatalf("%d +1/+0 counters, want 3", n)
	}
	if p, tt := effectstest.PtOf(t, g, id); p != 5 || tt != 1 {
		t.Errorf("the Serpent is %d/%d, want 5/1", p, tt)
	}
	abilities := effectstest.EffectiveAbilities(t, g, id)
	if !effectstest.ContainsString(abilities, "haste") || !effectstest.ContainsString(abilities, "trample") {
		t.Errorf("printed keywords missing: %v", abilities)
	}

	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myPower := effectstest.PushCounterCreature(g, me.ID, "Mine +1/+0", "+1/+0", 1)
	myWall := effectstest.PushCounterCreature(g, me.ID, "Mine -0/-1", "-0/-1", 1)
	theirContagion := effectstest.PushCounterCreature(g, opp.ID, "Theirs -2/-1", "-2/-1", 1)
	theirPower := effectstest.PushCounterCreature(g, opp.ID, "Theirs +1/+0", "+1/+0", 1)

	cards, _ := effects.BeneficialProliferateChoice(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shambler := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nested Shambler", TypeLine: "Creature — Zombie",
		OracleID: effectstest.B39NestedShamblerOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(shambler, "+1/+0", 2) })
	if p, _ := effectstest.PtOf(t, g, shambler); p != 3 {
		t.Fatalf("setup: a 1/1 with two +1/+0 counters has power 3, got %d", p)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shambler) })
	effectstest.PassPriorityAroundTable(t, g)
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
