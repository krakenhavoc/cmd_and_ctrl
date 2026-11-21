package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprCombineChrysalisOracle    = "a1cee8d7-f619-47b4-ba8c-faf3106d4e41"
	rprConclaveTribunalOracle    = "e938ee4c-d5df-4d93-bd61-9e518fb1dc30"
	rprConstableOfTheRealmOracle = "2a263e61-303f-449d-a431-c26cabdc68e0"
	rprCrackOpenOracle           = "412197f2-0e95-47fb-83db-ae8301c90ff0"
	rprDuergarHedgeMageOracle    = "ecb2fd58-b880-4d9e-93cb-c4d2b28233c9"
	rprEchoStormOracle           = "92442ca0-e49d-4fab-8d3a-83405c1b1437"
	rprEliteScaleguardOracle     = "b90af42d-edb6-4654-b7dc-10f0f272c883"
	rprEnduringScalelordOracle   = "270a21c1-4c44-46ef-8c4f-e08484f360e1"
	rprEphemeralShieldsOracle    = "e1a89535-7126-48e0-b335-b732a677b157"
	rprEtherswornAdjOracle       = "ca2a7669-4932-4a78-8a6c-8558b7c05ba5"
)

func rprHasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	var ok bool
	g.ReadSnapshot(func() { ok = game.HasKeyword(effectstest.FindBattlefieldCardForTest(g, id), kw) })
	return ok
}

func rprCountNamed(g *game.Game, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				n++
			}
		}
	})
	return n
}

// --- Conclave Tribunal -------------------------------------------------

func TestRprConclaveTribunalConvokesAndExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	soldiers := effectstest.PushTapCostSoldiers(g, me.ID, 4)
	tribunal, err := effectstest.CastWithTapParams(t, g, "Conclave Tribunal", "Enchantment", "{3}{W}", rprConclaveTribunalOracle,
		game.CastSpellParams{TapIDs: soldiers, Strict: true})
	if err != nil {
		t.Fatalf("convoke cast: %v", err)
	}
	for _, s := range soldiers {
		if !effectstest.TapCostTapped(g, s) {
			t.Error("convoke taps the named creatures")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tribunal) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after the Tribunal left, %d copies under its owner's control, want 1", n)
	}
}

func TestRprConclaveTribunalCantTargetYourOwnOrALand(t *testing.T) {
	spec, _ := effects.Lookup(rprConclaveTribunalOracle)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirLand := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, theirLand: false, theirs: true})
}

// --- Crack Open --------------------------------------------------------

func TestRprCrackOpenDestroysAnArtifactOrEnchantmentAndMakesATreasure(t *testing.T) {
	for _, tl := range []string{"Artifact", "Enchantment"} {
		t.Run(tl, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			target := effectstest.B12Permanent(g, opp.ID, "Their "+tl, tl)
			effectstest.CastCatalogSpell(t, g, "Crack Open", "Sorcery", rprCrackOpenOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(target) {
				t.Errorf("the %s survived", tl)
			}
			if n := effectstest.AfrTreasures(g, me.ID); n != 1 {
				t.Errorf("%d Treasures, want 1", n)
			}
		})
	}
}

func TestRprCrackOpenRefusesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Crack Open", "Sorcery", rprCrackOpenOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	if err == nil {
		t.Fatal("a creature is not a legal target")
	}
}

// --- Duergar Hedge-Mage ------------------------------------------------

func rprLands(g *game.Game, owner uuid.UUID, subtype string, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, owner, subtype, "Basic Land — "+subtype)
	}
}

func TestRprDuergarHedgeMageDestroysAnArtifactWithTwoMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rprLands(g, me.ID, "Mountain", 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	effectstest.CastAndResolveCreature(t, g, "Duergar Hedge-Mage", "Creature — Dwarf Shaman", rprDuergarHedgeMageOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetCards, ench) || !effectstest.HasID(p.PickTargetCards, rock) {
		t.Fatalf("legal targets wrong: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if !g.Battlefield.Contains(ench) {
		t.Error("the enchantment was destroyed with only Mountains")
	}
}

func TestRprDuergarHedgeMageDeclinedLeavesTheArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rprLands(g, me.ID, "Mountain", 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastAndResolveCreature(t, g, "Duergar Hedge-Mage", "Creature — Dwarf Shaman", rprDuergarHedgeMageOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed after answering no")
	}
}

func TestRprDuergarHedgeMageNeedsTwoOfTheLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rprLands(g, me.ID, "Mountain", 1)
	rprLands(g, me.ID, "Plains", 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	effectstest.CastAndResolveCreature(t, g, "Duergar Hedge-Mage", "Creature — Dwarf Shaman", rprDuergarHedgeMageOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a trigger asked for a target with one Mountain and one Plains: %+v", p)
	}
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(ench) {
		t.Error("something was destroyed without two of a land type")
	}
}

func TestRprDuergarHedgeMageTwoPlainsDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rprLands(g, me.ID, "Plains", 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	effectstest.CastAndResolveCreature(t, g, "Duergar Hedge-Mage", "Creature — Dwarf Shaman", rprDuergarHedgeMageOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed with only Plains")
	}
}

// --- Elite Scaleguard --------------------------------------------------

func TestRprEliteScaleguardBolstersTheWeakestCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	weak := effectstest.B12Creature(g, me.ID, "Weak", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "Strong", "Creature — Bear", 4, 4)
	effectstest.CastHoldCreature(t, g, "Elite Scaleguard", "Creature — Human Soldier", rprEliteScaleguardOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	var n int
	g.ReadSnapshot(func() { n = effectstest.PlusOneCounters(g, weak) })
	if n != 2 {
		t.Errorf("the least-toughness creature has %d +1/+1 counters, want 2", n)
	}
}

func rprScaleguardTable(t *testing.T) (g *game.Game, me, opp *game.Player, countered, plain, wall, third uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Elite Scaleguard", "Creature — Human Soldier", rprEliteScaleguardOracle, 2, 3)
	countered = effectstest.B12Creature(g, me.ID, "Countered", "Creature — Bear", 2, 2)
	plain = effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(countered, game.CounterPlusOne, 1) })
	wall = effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	third = effectstest.B12Creature(g, g.Seats[2].ID, "Third Seat", "Creature — Wall", 0, 4)
	return
}

func TestRprEliteScaleguardIgnoresAnAttackerWithNoCounter(t *testing.T) {
	g, me, opp, _, plain, wall, _ := rprScaleguardTable(t)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(plain, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("an attacker with no counter triggered the Scaleguard: %+v", p)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TapCostTapped(g, wall) {
		t.Error("a creature was tapped")
	}
}

func TestRprEliteScaleguardTapsADefenderWhenACounteredCreatureAttacks(t *testing.T) {
	g, me, opp, countered, _, wall, third := rprScaleguardTable(t)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(countered, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, wall) || effectstest.HasID(p.PickTargetCards, third) {
		t.Fatalf("legal targets must be only the defending player's creatures: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TapCostTapped(g, wall) {
		t.Error("the defending player's creature was not tapped")
	}
	if effectstest.TapCostTapped(g, third) {
		t.Error("another player's creature was tapped")
	}
}

// --- Enduring Scalelord ------------------------------------------------

func TestRprEnduringScalelordGrowsWhenAnotherCreatureYouControlGetsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B12Push(g, me.ID, "Enduring Scalelord", "Creature — Dragon", rprEnduringScalelordOracle, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	// Counters on an opponent's creature do nothing.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, lord); n != 0 {
		t.Fatalf("the Scalelord has %d counters after an opponent's creature grew", n)
	}

	// Two counters at once are one placement: one trigger, one counter.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, game.CounterPlusOne, 2) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, lord); n != 1 {
		t.Fatalf("the Scalelord has %d counters, want 1", n)
	}

	// Its own counters never fire it.
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, lord); n != 1 {
		t.Errorf("the Scalelord chained itself to %d counters", n)
	}
}

func TestRprEnduringScalelordMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lord := effectstest.B12Push(g, me.ID, "Enduring Scalelord", "Creature — Dragon", rprEnduringScalelordOracle, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, lord); n != 0 {
		t.Errorf("the Scalelord has %d counters after declining", n)
	}
	if !rprHasKeyword(g, lord, "flying") {
		t.Error("the Scalelord has no flying")
	}
}

// --- Constable of the Realm --------------------------------------------

func TestRprConstableExilesWhenItGetsCountersAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	constable := effectstest.B12Push(g, me.ID, "Constable of the Realm", "Creature — Giant Soldier", rprConstableOfTheRealmOracle, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(constable, game.CounterPlusOne, 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetCards, constable) || !effectstest.HasID(p.PickTargetCards, victim) {
		t.Fatalf("legal targets wrong: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Fatal("the permanent is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(constable) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after the Constable left, %d copies under its owner's control, want 1", n)
	}
}

func TestRprConstableRenownAttackPutsTwoCountersAndTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	constable := effectstest.PushRenownCard(g, me.ID, "Constable of the Realm", rprConstableOfTheRealmOracle, 3, 3, "renown 2")
	victim := effectstest.B12Creature(g, g.Seats[2].ID, "Third Threat", "Creature — Bear", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(constable, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	var n int
	g.ReadSnapshot(func() { n = effectstest.PlusOneCounters(g, constable) })
	if n != 2 || !effectstest.RenownedNow(g, constable) {
		t.Errorf("%d counters, renowned %v; want 2 and renowned", n, effectstest.RenownedNow(g, constable))
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the permanent was not exiled by the renown counters")
	}
}

// --- Ephemeral Shields -------------------------------------------------

func TestRprEphemeralShieldsConvokesAndGrantsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	soldiers := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	_, err := effectstest.CastWithTapParams(t, g, "Ephemeral Shields", "Instant", "{1}{W}", rprEphemeralShieldsOracle,
		game.CastSpellParams{TapIDs: soldiers, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if err != nil {
		t.Fatalf("convoke cast: %v", err)
	}
	for _, s := range soldiers {
		if !effectstest.TapCostTapped(g, s) {
			t.Error("convoke taps the named creatures")
		}
	}
	if rprHasKeyword(g, bear, "indestructible") {
		t.Error("the target is indestructible before the spell resolves")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !rprHasKeyword(g, bear, "indestructible") {
		t.Error("the target did not gain indestructible")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	g.WithWriteLock(func() {})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if rprHasKeyword(g, bear, "indestructible") {
		t.Error("indestructible outlasted the turn")
	}
}

// --- Ethersworn Adjudicator --------------------------------------------

func TestRprEtherswornAdjudicatorDestroysAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adj := effectstest.B12Push(g, me.ID, "Ethersworn Adjudicator", "Artifact Creature — Vedalken Knight", rprEtherswornAdjOracle, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)

	effectstest.Std51Pay(g, me, "{1}{W}{B}")
	if err := g.ActivateCatalogAbility(me.ID, adj, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err == nil {
		t.Fatal("an artifact is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, adj, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	if !effectstest.TapCostTapped(g, adj) {
		t.Error("the Adjudicator is not tapped")
	}

	// Tapped, the destroy can't be paid again until it untaps.
	effectstest.Std51Pay(g, me, "{1}{W}{B}")
	if err := g.ActivateCatalogAbility(me.ID, adj, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}}); err == nil {
		t.Fatal("the tapped Adjudicator destroyed again")
	}

	effectstest.Std51Pay(g, me, "{2}{U}")
	if err := g.ActivateCatalogAbility(me.ID, adj, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("untap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TapCostTapped(g, adj) {
		t.Fatal("the Adjudicator is still tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, adj, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}}); err != nil {
		t.Fatalf("second destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
}

// --- Combine Chrysalis -------------------------------------------------

func TestRprCombineChrysalisGivesTokensFlyingAndMakesABeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chrysalis := effectstest.B12Push(g, me.ID, "Combine Chrysalis", "Artifact", rprCombineChrysalisOracle, 0, 0)
	var mine, theirs uuid.UUID
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 black Zombie"), 1)
		_ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("2/2 black Zombie"), 1)
	})
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Zombie" && c.Controller == me.ID {
				mine = c.InstanceID
			}
			if c.Name == "Zombie" && c.Controller == opp.ID {
				theirs = c.InstanceID
			}
		}
	})
	nontoken := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !rprHasKeyword(g, mine, "flying") {
		t.Error("your creature token has no flying")
	}
	if rprHasKeyword(g, theirs, "flying") {
		t.Error("an opponent's token has flying")
	}
	if rprHasKeyword(g, nontoken, "flying") {
		t.Error("a nontoken creature has flying")
	}

	effectstest.AdvanceToMain(t, g)
	effectstest.Std51Pay(g, me, "{2}{G}{U}")
	if err := g.ActivateCatalogAbility(me.ID, chrysalis, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{mine}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if rprCountNamed(g, "Beast") != 1 {
		t.Errorf("%d Beasts, want 1", rprCountNamed(g, "Beast"))
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the sacrificed token is still on the battlefield")
	}
}

func TestRprCombineChrysalisNeedsATokenToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chrysalis := effectstest.B12Push(g, me.ID, "Combine Chrysalis", "Artifact", rprCombineChrysalisOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.Std51Pay(g, me, "{2}{G}{U}")
	if err := g.ActivateCatalogAbility(me.ID, chrysalis, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("activated with no token to sacrifice")
	}
}

// --- Echo Storm --------------------------------------------------------

func TestRprEchoStormCopiesOncePerCommanderCast(t *testing.T) {
	for _, casts := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		me.CommanderCasts[uuid.New()] = casts
		rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		effectstest.CastCatalogSpell(t, g, "Echo Storm", "Sorcery", rprEchoStormOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
		for i := 0; i < 8; i++ {
			if c := effectstest.LatestPickTarget(g, me.ID); c != nil {
				effectstest.PickCard(t, g, me.ID, rock)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		if got, want := rprCountNamed(g, "Their Rock"), 1+1+casts; got != want {
			t.Errorf("with %d commander casts there are %d Rocks, want %d (the original plus a token per spell)", casts, got, want)
		}
	}
}
