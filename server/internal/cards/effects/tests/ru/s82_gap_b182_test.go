package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b182_test.go — GAP batch 182 (#5948, Starter Commander Decks):
// every card driven through the engine.

const (
	gap182DisintegrationOracle = "eaa2ff39-c3cf-428b-a2f1-a6a99a2c806b"
	gap182ScatterOracle        = "c17d5153-cc74-4fbb-9d75-f7ee7b557b0a"
	gap182ProvokeOracle        = "ac844303-487b-49e9-9170-02892153ca6a"
	gap182RicochetOracle       = "8c35fd11-be45-4984-bd83-6e4f3fbc47a9"
	gap182SlagheapOracle       = "33587cb2-0fd3-4e4c-bc5e-e7299cc9dab5"
	gap182BloodhallOracle      = "8220c5fa-28dc-40d0-a38a-d8eefc2795d6"
	gap182VolcanoOracle        = "018129dd-f578-4fc9-b6f8-b13ed0c75a8d"
	gap182JwarOracle           = "8b96f837-7c32-473a-b5ae-1d66527eaf7b"
	gap182SaltMarshOracle      = "6a40823f-130c-4416-a987-75e205d3e1dd"
)

func TestS82Gap182CardsAreRegisteredFull(t *testing.T) {
	for _, id := range []string{gap182DisintegrationOracle, gap182ScatterOracle, gap182ProvokeOracle, gap182RicochetOracle,
		gap182SlagheapOracle, gap182BloodhallOracle, gap182VolcanoOracle, gap182JwarOracle, gap182SaltMarshOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- the lands -------------------------------------------------------------

func TestS82Gap182TappedDualsEnterTappedAndTapForEitherColour(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Urborg Volcano", gap182VolcanoOracle},
		{"Salt Marsh", gap182SaltMarshOracle},
		{"Jwar Isle Refuge", gap182JwarOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		land := effectstest.PlayLandFromHand(t, g, c.name, c.oracle)
		if !effectstest.LayeredCard(t, g, land).Tapped {
			t.Errorf("%s did not enter tapped", c.name)
		}
		spec, _ := effects.Lookup(c.oracle)
		if len(spec.ManaAbilities) != 1 || len(spec.ManaAbilities[0].Produced) == 0 {
			t.Errorf("%s: mana abilities %+v", c.name, spec.ManaAbilities)
		}
	}
}

func TestS82Gap182JwarIsleRefugeGainsOneLifeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.PlayLandFromHand(t, g, "Jwar Isle Refuge", gap182JwarOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before+1 {
		t.Errorf("life %d -> %d after the land entered, want +1", before, got)
	}
}

func TestS82Gap182MoltenSlagheapStoresCountersAndPaysBlackOrRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Molten Slagheap", "Land", gap182SlagheapOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("store: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, land, game.CounterStorage); got != 1 {
		t.Fatalf("storage counters %d, want 1", got)
	}
	effectstest.AddCountersForTest(g, land, game.CounterStorage, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, land).Tapped = false })
	effectstest.B06AddMana(me, "R")
	before := me.ManaPool.Total()
	effectstest.ActivateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{CounterCounts: []int{2}})
	for i := 0; i < 2; i++ {
		var pick *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
				pick = c
			}
		}
		if pick == nil {
			break
		}
		color := "B"
		if i == 1 {
			color = "R"
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, color); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	// The {1} was paid from the pool and two mana came out.
	if got := me.ManaPool.Total(); got != before-1+2 {
		t.Errorf("pool %d, want %d (paid {1}, got 2 for 2 counters)", got, before-1+2)
	}
	if got := effectstest.CounterCount(g, land, game.CounterStorage); got != 0 {
		t.Errorf("%d storage counters left, want 0", got)
	}
}

func TestS82Gap182StensiaBloodhallHitsAPlayerNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Stensia Bloodhall", "Land", gap182BloodhallOracle, false)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToMain(t, g)
	for _, c := range []string{"C", "C", "C", "B", "R"} {
		effectstest.B06AddMana(me, c)
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("the Bloodhall targeted a creature")
	}
	before := effectstest.LifeOf(g, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-2 {
		t.Errorf("life %d -> %d, want -2", before, got)
	}
}

func TestS82Gap182StensiaBloodhallRefusesAPoolThatCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Stensia Bloodhall", "Land", gap182BloodhallOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("the Bloodhall was activated without {3}{B}{R}")
	}
}

// --- Scatter the Seeds ----------------------------------------------------

func TestS82Gap182ScatterTheSeedsMakesThreeSaprolingsAndConvokes(t *testing.T) {
	spec, ok := effects.Lookup(gap182ScatterOracle)
	if !ok || spec.TapCost == nil {
		t.Fatal("Scatter the Seeds must declare convoke as its tap cost")
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Scatter the Seeds", "Instant", gap182ScatterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Saproling", me.ID); n != 3 {
		t.Errorf("%d Saproling tokens, want 3", n)
	}
}

// --- Unlicensed Disintegration ------------------------------------------

func gap182Disintegrate(t *testing.T, withArtifact bool) (victimLife, before int, dead bool) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, opp.ID)
	if withArtifact {
		effectstest.PushArtifactForTest(g, me.ID, "Rock")
	}
	before = effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Unlicensed Disintegration", "Instant", gap182DisintegrationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	return effectstest.LifeOf(g, opp.ID), before, !g.Battlefield.Contains(bear)
}

func TestS82Gap182UnlicensedDisintegrationHitsTheControllerOnlyWithAnArtifact(t *testing.T) {
	life, before, dead := gap182Disintegrate(t, true)
	if !dead {
		t.Error("the creature was not destroyed")
	}
	if life != before-3 {
		t.Errorf("life %d -> %d with an artifact, want -3", before, life)
	}
	life, before, dead = gap182Disintegrate(t, false)
	if !dead {
		t.Error("the creature was not destroyed without an artifact")
	}
	if life != before {
		t.Errorf("life %d -> %d without an artifact, want no change", before, life)
	}
}

func TestS82Gap182UnlicensedDisintegrationRefusesANonCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	if err := effectstest.CastCatalogSpellErr(t, g, "Unlicensed Disintegration", "Instant", gap182DisintegrationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("Unlicensed Disintegration was cast at an artifact")
	}
}

// --- Provoke the Trolls ---------------------------------------------------

func TestS82Gap182ProvokeTheTrollsPumpsOnlyACreatureItDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Big Bear", 2, 5)
	effectstest.CastCatalogSpell(t, g, "Provoke the Trolls", "Instant", gap182ProvokeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 7 {
		t.Errorf("the damaged creature has power %d, want 7", got)
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("the pump outlived the turn: power %d", got)
	}
}

func TestS82Gap182ProvokeTheTrollsKillsAndDamagesAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Provoke the Trolls", "Instant", gap182ProvokeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-3 {
		t.Errorf("life %d -> %d, want -3", before, got)
	}
}

// --- Wild Ricochet ----------------------------------------------------------

func gap182RicochetSetup(t *testing.T) (g *game.Game, me, victim, other uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, victim, other = g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	effectstest.CastCatalogSpell(t, g, "Wild Ricochet", "Instant", gap182RicochetOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	for i := 0; i < 8 && effectstest.LatestRetarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.LatestRetarget(g, me) == nil {
		t.Fatal("Wild Ricochet opened no retarget prompt for the spell")
	}
	return g, me, victim, other
}

// The new targets are chosen for the spell first, and the copy is then made
// from the spell with its new target: both hit the new target.
func TestS82Gap182WildRicochetRetargetsTheSpellThenCopiesIt(t *testing.T) {
	g, me, victim, other := gap182RicochetSetup(t)
	before, victimBefore := effectstest.LifeOf(g, other), effectstest.LifeOf(g, victim)
	effectstest.AnswerRetargetWith(t, g, me, effectstest.RtPlayerRef(other))
	effectstest.SettleThroughCopyRetarget(t, g, me, other)
	if got := effectstest.LifeOf(g, other); got != before-6 {
		t.Errorf("the new target has %d life, want %d (the Bolt and its copy)", got, before-6)
	}
	if got := effectstest.LifeOf(g, victim); got != victimBefore {
		t.Errorf("the original target has %d life, want %d", got, victimBefore)
	}
}

// Declining the first offer leaves the Bolt where it was; the copy is then
// pointed at another player.
func TestS82Gap182WildRicochetMayDeclineAndRetargetOnlyTheCopy(t *testing.T) {
	g, me, victim, other := gap182RicochetSetup(t)
	before, victimBefore := effectstest.LifeOf(g, other), effectstest.LifeOf(g, victim)
	effectstest.AnswerRetargetWith(t, g, me, nil)
	effectstest.SettleThroughCopyRetarget(t, g, me, other)
	if got := effectstest.LifeOf(g, victim); got != victimBefore-3 {
		t.Errorf("the original target has %d life, want %d", got, victimBefore-3)
	}
	if got := effectstest.LifeOf(g, other); got != before-3 {
		t.Errorf("the copy's target has %d life, want %d", got, before-3)
	}
}
