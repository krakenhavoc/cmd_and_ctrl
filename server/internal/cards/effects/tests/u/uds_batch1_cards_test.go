package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// uds_batch1_cards_test.go — Urza's Destiny, batch 1 (#5621). Each test
// plays the real catalog entry and asserts what its text says.

const (
	udsB1KingfisherOracle        = "2a70f3d7-9b5c-455e-8c4b-55870d77bd7d"
	udsB1BrassSecretaryOracle    = "0d8c2d7b-7cce-4115-a3f8-18c23731544e"
	udsB1SlinkingSkirgeOracle    = "82875793-b264-4ceb-8525-ef3b6d086072"
	udsB1HeartWardenOracle       = "babcc551-39bb-4b1c-92a0-e60b236a59b2"
	udsB1BraidwoodCupOracle      = "d52b72ff-a82d-430e-94c7-675c83b43e50"
	udsB1PhyrexianMonitorOracle  = "1eb6c28d-ebac-4016-ad07-c90081ee4983"
	udsB1CapashenTemplarOracle   = "ab9ad5e2-2bf3-450d-9d5c-c7280d368e10"
	udsB1CapashenKnightOracle    = "e295d207-128f-419a-866a-c9ac6248c199"
	udsB1ColosYearlingOracle     = "50cf702b-5647-415e-b9a8-56fa0e50da6d"
	udsB1TwistedExperimentOracle = "4066d4df-d98f-44cd-bf25-ebb5e9d9ddeb"
	udsB1MagnifyOracle           = "421c661a-24d9-4797-b538-2a4243825628"
	udsB1TrumpetBlastOracle      = "d3b408e7-4d39-47a3-b00c-846903f8ae77"
	udsB1ReliquaryMonkOracle     = "a149ac31-b792-4fbe-903a-788bbdfd5e97"
	udsB1GoblinGardenerOracle    = "29781b19-0917-4d8f-abe3-df0b7e96d8a6"
)

func TestUdsB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		udsB1KingfisherOracle:        "Kingfisher",
		udsB1BrassSecretaryOracle:    "Brass Secretary",
		udsB1SlinkingSkirgeOracle:    "Slinking Skirge",
		udsB1HeartWardenOracle:       "Heart Warden",
		udsB1BraidwoodCupOracle:      "Braidwood Cup",
		udsB1PhyrexianMonitorOracle:  "Phyrexian Monitor",
		udsB1CapashenTemplarOracle:   "Capashen Templar",
		udsB1CapashenKnightOracle:    "Capashen Knight",
		udsB1ColosYearlingOracle:     "Colos Yearling",
		udsB1TwistedExperimentOracle: "Twisted Experiment",
		udsB1MagnifyOracle:           "Magnify",
		udsB1TrumpetBlastOracle:      "Trumpet Blast",
		udsB1ReliquaryMonkOracle:     "Reliquary Monk",
		udsB1GoblinGardenerOracle:    "Goblin Gardener",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Kingfisher ----------------------------------------------------------------

func TestUdsB1KingfisherFliesAndDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	bird := effectstest.B12Push(g, me.ID, "Kingfisher", "Creature — Bird", udsB1KingfisherOracle, 2, 2)
	effectstest.AssertKeywords(t, g, bird, "flying")
	hand := me.Hand.Size()
	effectstest.B27Kill(g, bird)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Brass Secretary, Slinking Skirge, Heart Warden ------------------------------

func TestUdsB1SacrificeToDrawRows(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Brass Secretary", "Artifact Creature — Construct", udsB1BrassSecretaryOracle},
		{"Slinking Skirge", "Creature — Phyrexian Imp", udsB1SlinkingSkirgeOracle},
		{"Heart Warden", "Creature — Elf Druid", udsB1HeartWardenOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedLibrary(me, "A", "B")
			// Summoning sick: the row has no {T}, so it is still usable.
			id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, true)
			row := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})) - 1
			hand := me.Hand.Size()
			effectstest.FloatForTest(g, me, "C")
			if err := g.ActivateCatalogAbility(me.ID, id, row, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("activated for {1}; it costs {2}")
			}
			effectstest.FloatForTest(g, me, "C")
			if err := g.ActivateCatalogAbility(me.ID, id, row, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
				t.Error("the creature was not sacrificed as the cost")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("drew %d, want 1", got)
			}
		})
	}
}

func TestUdsB1SlinkingSkirgeHasFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	imp := effectstest.B12Push(g, g.Seats[0].ID, "Slinking Skirge", "Creature — Phyrexian Imp", udsB1SlinkingSkirgeOracle, 2, 1)
	effectstest.AssertKeywords(t, g, imp, "flying")
}

func TestUdsB1HeartWardenTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Heart Warden", "Creature — Elf Druid", udsB1HeartWardenOracle, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a summoning-sick Heart Warden tapped for mana (CR 302.6)")
	}
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Heart Warden", "Creature — Elf Druid", udsB1HeartWardenOracle, false)
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want G", got)
	}
}

// --- Braidwood Cup ---------------------------------------------------------------

func TestUdsB1BraidwoodCupTapsForOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	cup := effectstest.PushCatalogPermanent(g, me.ID, "Braidwood Cup", "Artifact", udsB1BraidwoodCupOracle, true)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, cup, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the turn it arrived (a noncreature artifact): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	if !effectstest.B13Tapped(t, g, cup) {
		t.Error("Braidwood Cup did not tap")
	}
	if err := g.ActivateCatalogAbility(me.ID, cup, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a tapped Cup")
	}
}

// --- Phyrexian Monitor -----------------------------------------------------------

func TestUdsB1PhyrexianMonitorRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	monitor := effectstest.B12Push(g, me.ID, "Phyrexian Monitor", "Creature — Phyrexian Skeleton", udsB1PhyrexianMonitorOracle, 2, 2)
	if err := effectstest.UsgB2Activate(t, g, me, monitor, "", nil); err == nil {
		t.Fatal("activated without paying {B}")
	}
	if err := effectstest.UsgB2Activate(t, g, me, monitor, "B", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.RegenShieldsOn(t, g, monitor); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	effectstest.DestroyIt(t, g, monitor)
	c, ok := effectstest.BattlefieldCard(g, monitor)
	if !ok {
		t.Fatal("the shield did not replace the destruction")
	}
	if !c.Tapped {
		t.Error("a regenerated creature is tapped")
	}
}

// --- Capashen Templar, Capashen Knight, Colos Yearling ---------------------------

func TestUdsB1SelfPumpRows(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana, keyword string
		power, toughness, dp, dt              int
	}{
		{"Capashen Templar", "Creature — Human Knight", udsB1CapashenTemplarOracle, "W", "", 2, 2, 0, 1},
		{"Capashen Knight", "Creature — Human Knight", udsB1CapashenKnightOracle, "CW", "first strike", 1, 1, 1, 0},
		{"Colos Yearling", "Creature — Goat Beast", udsB1ColosYearlingOracle, "R", "mountainwalk", 1, 1, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if tc.keyword != "" {
				effectstest.AssertKeywords(t, g, id, tc.keyword)
			}
			for i := 1; i <= 2; i++ {
				if err := effectstest.UsgB2Activate(t, g, me, id, tc.mana, nil); err != nil {
					t.Fatalf("activation %d: %v", i, err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				want := [2]int{tc.power + i*tc.dp, tc.toughness + i*tc.dt}
				if p, tough := effectstest.EffectivePTOf(t, g, id); p != want[0] || tough != want[1] {
					t.Errorf("after %d activations %d/%d, want %d/%d", i, p, tough, want[0], want[1])
				}
			}
			if err := effectstest.UsgB2Activate(t, g, me, id, "", nil); err == nil {
				t.Error("activated with no mana")
			}
		})
	}
}

// --- Twisted Experiment ------------------------------------------------------------

func TestUdsB1TwistedExperimentGivesPlusThreeMinusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Twisted Experiment", udsB1TwistedExperimentOracle, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 1 {
		t.Errorf("the enchanted Bear is %d/%d, want 5/1", p, tough)
	}
	elf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	effectstest.AuraCast(t, g, "Twisted Experiment", udsB1TwistedExperimentOracle, elf)
	if g.Battlefield.Contains(elf) || !opp.Graveyard.Contains(elf) {
		t.Error("a 1-toughness creature survived at toughness 0 (CR 704.5f)")
	}
}

// --- Magnify --------------------------------------------------------------------------

func TestUdsB1MagnifyPumpsEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Magnify", "Instant", udsB1MagnifyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("a creature is %d/%d, want 3/3", p, tough)
		}
	}
	late := effectstest.B12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	if p, tough := effectstest.PtOf(t, g, late); p != 2 || tough != 2 {
		t.Errorf("a creature that arrived after resolution is %d/%d, want 2/2", p, tough)
	}
}

// --- Trumpet Blast -------------------------------------------------------------------

func TestUdsB1TrumpetBlastPumpsEveryAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home", "Creature — Bear", 2, 2)
	effectstest.ArnSetCombat(g, mine, opp.ID, uuid.Nil)
	effectstest.ArnSetCombat(g, theirs, me.ID, uuid.Nil)
	effectstest.CastCatalogSpell(t, g, "Trumpet Blast", "Instant", udsB1TrumpetBlastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 2 {
			t.Errorf("attacker %d/%d, want 4/2", p, tough)
		}
	}
	if p, tough := effectstest.PtOf(t, g, home); p != 2 || tough != 2 {
		t.Errorf("a creature at home %d/%d, want 2/2", p, tough)
	}
}

// --- Reliquary Monk, Goblin Gardener -----------------------------------------------

func TestUdsB1DiesTriggersDestroyTheirTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		victim, wrong          string // the target's type line, and one it can't take
	}{
		{"Reliquary Monk", "Creature — Human Monk Cleric", udsB1ReliquaryMonkOracle, 2, 2, "Enchantment", "Basic Land — Forest"},
		{"Goblin Gardener", "Creature — Goblin", udsB1GoblinGardenerOracle, 2, 1, "Basic Land — Forest", "Artifact"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.JmpB7Seats(g)
			dier := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			victim := effectstest.B12Permanent(g, opp.ID, "Victim", tc.victim)
			wrong := effectstest.B12Permanent(g, opp.ID, "Wrong", tc.wrong)
			effectstest.AdvanceToMain(t, g)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dier) })
			effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
			if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
				t.Errorf("the %s target was not destroyed", tc.victim)
			}
			if !g.Battlefield.Contains(wrong) {
				t.Errorf("the %s was destroyed too", tc.wrong)
			}
		})
	}
}

func TestUdsB1ReliquaryMonkTakesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.JmpB7Seats(g)
	monk := effectstest.B12Push(g, me.ID, "Reliquary Monk", "Creature — Human Monk Cleric", udsB1ReliquaryMonkOracle, 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, monk)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: rock}, true)
	if g.Battlefield.Contains(rock) {
		t.Error("the target artifact is still on the battlefield")
	}
}

func TestUdsB1DiesTriggersIgnoreOtherDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.JmpB7Seats(g)
	effectstest.B12Push(g, me.ID, "Goblin Gardener", "Creature — Goblin", udsB1GoblinGardenerOracle, 2, 1)
	effectstest.B12Push(g, me.ID, "Reliquary Monk", "Creature — Human Monk Cleric", udsB1ReliquaryMonkOracle, 2, 2)
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("another creature dying triggered a dies trigger")
	}
}
