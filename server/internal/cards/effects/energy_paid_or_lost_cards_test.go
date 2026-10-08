package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_paid_or_lost_cards_test.go — ADR 0129 PR 5 (#1995): the cards
// that trigger on getting energy (WheneverYouGetEnergy) and the ones
// that read the energy paid or lost this turn
// (PlayerTurnTally.EnergyPaidOrLost), through the real catalog.

const (
	izzetGeneratoriumOracle = "6a618c4a-e604-4dce-a078-755fd35ac5d9"
	blasterHulkOracle       = "4b8ca7fd-3c8e-44fe-ab0a-0ddc2f2b047c"
	aetherRevoltOracle      = "b73661a6-d136-4c65-804c-461e83484f4b"
	brotherhoodScribeOracle = "e22bb590-9c9a-4a99-8cd7-1dbe742f4cdd"
	fabricationModuleOracle = "b057c8ff-f169-41cd-a594-026f1a6cb0f9"
	territorialGorgerOracle = "fcf9c013-7a79-42b1-8aee-6d40ff475fbf"
)

func TestEnergyPaidOrLostCardsAreFull(t *testing.T) {
	for _, oracle := range []string{izzetGeneratoriumOracle, blasterHulkOracle, aetherRevoltOracle,
		brotherhoodScribeOracle, fabricationModuleOracle, territorialGorgerOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// getEnergy puts n energy on p as one placement, the way an effect's
// "you get" does.
func getEnergy(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddPlayerCounterByForEffect(p.ID, p.ID, game.CounterEnergy, n) })
	if err != nil {
		t.Fatalf("get energy: %v", err)
	}
}

// loseEnergy takes n energy off p by an effect.
func loseEnergy(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddPlayerCounterForEffect(p.ID, game.CounterEnergy, -n) })
	if err != nil {
		t.Fatalf("lose energy: %v", err)
	}
}

// pickTargetRefs answers the newest pick_target owed by `chooser`.
func pickTargetRefs(t *testing.T, g *game.Game, chooser uuid.UUID, dist map[uuid.UUID]int, refs ...game.TargetRef) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	var err error
	if dist != nil {
		err = g.ResolvePickTargetsDivided(p.ID, chooser, refs, dist)
	} else {
		err = g.ResolvePickTargets(p.ID, chooser, refs)
	}
	if err != nil {
		t.Fatalf("answer targets: %v", err)
	}
}

// Territorial Gorger: +2/+2 per placement of energy, however many
// counters it put on; paying energy is not getting it.
func TestTerritorialGorgerGrowsEachTimeYouGetEnergy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	gorger := pushDiesCreatureForTest(g, me.ID, "Territorial Gorger", territorialGorgerOracle, "Creature — Gremlin", 2, 2)
	getEnergy(t, g, me, 3)
	passPriorityAroundTable(t, g)
	if c := apaLive(g, gorger); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Fatalf("after one placement of three: %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	getEnergy(t, g, me, 1)
	passPriorityAroundTable(t, g)
	loseEnergy(t, g, me, 2)
	passPriorityAroundTable(t, g)
	if c := apaLive(g, gorger); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("after a second placement and a removal: %d/%d, want 6/6", c.CurrentPower(), c.CurrentToughness())
	}
	// An opponent getting energy is not you getting it.
	getEnergy(t, g, g.Seats[1], 2)
	passPriorityAroundTable(t, g)
	if c := apaLive(g, gorger); c.CurrentPower() != 6 {
		t.Errorf("an opponent's energy pumped it to %d", c.CurrentPower())
	}
}

// Fabrication Module: its own {4}, {T} gets energy, which triggers it.
func TestFabricationModuleCountersOnGettingEnergy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	mod := pushPermanentForTest(g, me.ID, "Fabrication Module", fabricationModuleOracle, "Artifact")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := 0; i < 4; i++ {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, mod, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	settleEnergy(t, g, me.ID, erPlay{target: bear})
	if energyOf(me) != 1 {
		t.Errorf("energy = %d, want 1", energyOf(me))
	}
	if got := apaLive(g, bear).Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// Brotherhood Scribe: the tap ability needs three artifacts; the pump
// is only on its controller's turn.
func TestBrotherhoodScribe(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	scribe := pushDiesCreatureForTest(g, me.ID, "Brotherhood Scribe", brotherhoodScribeOracle, "Creature — Human Artificer", 1, 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	pushPermanentForTest(g, me.ID, "Relic A", "", "Artifact")
	pushPermanentForTest(g, me.ID, "Relic B", "", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, scribe, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("metalcraft: activated with two artifacts")
	}
	pushPermanentForTest(g, me.ID, "Relic C", "", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, scribe, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with three artifacts: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 1 {
		t.Errorf("energy = %d, want 1", energyOf(me))
	}
	if c := apaLive(g, bear); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := apaLive(g, scribe); c.CurrentPower() != 2 {
		t.Errorf("the Scribe is %d power, want 2", c.CurrentPower())
	}
	if c := apaLive(g, theirs); c.CurrentPower() != 2 {
		t.Errorf("an opponent's creature got the pump: %d power", c.CurrentPower())
	}

	// On an opponent's turn, getting energy does nothing.
	advanceToMainOf(t, g, 1)
	getEnergy(t, g, me, 1)
	passPriorityAroundTable(t, g)
	if c := apaLive(g, bear); c.CurrentPower() != 2 {
		t.Errorf("on an opponent's turn the Bear is %d power, want 2", c.CurrentPower())
	}
}

// Izzet Generatorium: one more energy each time you get some, and the
// draw needs four paid or lost this turn.
func TestIzzetGeneratorium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	gen := pushPermanentForTest(g, me.ID, "Izzet Generatorium", izzetGeneratoriumOracle, "Artifact")
	getEnergy(t, g, me, 2)
	if energyOf(me) != 3 {
		t.Fatalf("getting two gave %d, want 3", energyOf(me))
	}
	// An opponent's energy is not replaced.
	getEnergy(t, g, g.Seats[1], 2)
	if energyOf(g.Seats[1]) != 2 {
		t.Errorf("an opponent got %d, want 2", energyOf(g.Seats[1]))
	}

	loseEnergy(t, g, me, 3)
	if err := g.ActivateCatalogAbility(me.ID, gen, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with three paid or lost")
	}
	getEnergy(t, g, me, 1) // 2 with the Generatorium
	loseEnergy(t, g, me, 1)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gen, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with four paid or lost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// Aether Revolt: "that much damage" is the energy that landed; with
// revolt, its noncombat damage to an opponent is 2 more, and damage to
// your own creature is not.
func TestAetherRevolt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	pushPermanentForTest(g, me.ID, "Aether Revolt", aetherRevoltOracle, "Enchantment")
	mine := pushVanillaCreature(g, me.ID, "Mine", 0, 30)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)

	life := opp.Life
	getEnergy(t, g, me, 2)
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("no revolt: opponent lost %d, want 2", life-opp.Life)
	}

	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(fodder); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	life = opp.Life
	getEnergy(t, g, me, 1)
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("revolt: opponent lost %d, want 1 + 2", life-opp.Life)
	}

	getEnergy(t, g, me, 1)
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: mine})
	passPriorityAroundTable(t, g)
	if d := apaLive(g, mine).DamageMarked; d != 1 {
		t.Errorf("your own creature took %d, want 1 (no bonus)", d)
	}
}

// Blaster Hulk: {1} less per energy paid or lost this turn, generic only.
func TestBlasterHulkCostsLessForEnergyPaidOrLost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hulk := game.Card{InstanceID: uuid.New(), Name: "Blaster Hulk", TypeLine: "Artifact Creature — Pirate",
		ManaCost: "{6}{R}{R}", OracleID: blasterHulkOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(hulk)
	price := func() int {
		t.Helper()
		p, err := g.PriceCast(me.ID, hulk, game.CastSpellParams{})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		return p.Total.ManaValue()
	}
	if got := price(); got != 8 {
		t.Fatalf("nothing paid or lost: mana value %d, want 8", got)
	}
	setEnergy(t, g, me, 10)
	loseEnergy(t, g, me, 4)
	if got := price(); got != 4 {
		t.Fatalf("four lost: mana value %d, want 4", got)
	}
	loseEnergy(t, g, me, 6)
	if got := price(); got != 2 {
		t.Errorf("ten lost: mana value %d, want 2 ({R}{R})", got)
	}
}

// Blaster Hulk attacks: it gets {E}{E}, pays eight, and the reflexive
// trigger divides 8 damage as announced.
func TestBlasterHulkAttacksAndDividesEight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulk := pushDiesCreatureForTest(g, me.ID, "Blaster Hulk", blasterHulkOracle, "Artifact Creature — Pirate", 8, 8)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 30)
	setEnergy(t, g, me, 6)
	life := opp.Life
	declareAttack(t, g, opp.ID, hulk)
	passPriorityAroundTable(t, g)
	c := energyPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no energy prompt after the attack trigger resolved")
	}
	if energyOf(me) != 8 {
		t.Fatalf("energy before paying = %d, want 6 + 2", energyOf(me))
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, map[uuid.UUID]int{opp.ID: 5, wall: 3},
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, game.TargetRef{Kind: game.TargetCard, ID: wall})
	passPriorityAroundTable(t, g)
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
	if opp.Life != life-5 {
		t.Errorf("the player's share: lost %d, want 5", life-opp.Life)
	}
	if d := apaLive(g, wall).DamageMarked; d != 3 {
		t.Errorf("the wall's share: %d, want 3", d)
	}
	if got := g.TurnTallyFor(me.ID).EnergyPaidOrLost; got != 8 {
		t.Errorf("paid or lost this turn = %d, want 8", got)
	}
}
