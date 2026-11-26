package testss82m

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b355_test.go — S82 older sets, batch 355: the Magic 2015 Auras
// (#5556).

const (
	oldB355EnsoulArtifact = "02628058-c9d8-4265-afd1-3fb697ec89f7"
	oldB355VerdantHaven   = "09381494-393d-4ddf-b79f-7c0ae5dce912"
	oldB355Encrust        = "2e056743-78ff-45b6-80df-7486d52d5df6"
	oldB355OppressiveRays = "4a17fb46-70db-4c9a-ac0e-dfc774345673"
	oldB355Invisibility   = "de26b0c6-dfb7-45a8-9d7f-f8d45522d675"
	oldB355SpectraWard    = "effb99e6-db1b-4ada-879d-431d9d13d4b7"
	oldB355AirServant     = "49638d96-e83e-4aab-bafb-d6183a6e3ff6"
)

func TestOldB355EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB355EnsoulArtifact: "Ensoul Artifact", oldB355VerdantHaven: "Verdant Haven",
		oldB355Encrust: "Encrust", oldB355OppressiveRays: "Oppressive Rays",
		oldB355Invisibility: "Invisibility", oldB355SpectraWard: "Spectra Ward",
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

func oldB355Attach(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.PushCatalogPermanent(g, owner, name, effectstest.AuraTypeLine, oracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach %s: %v", name, err)
		}
	})
	return aura
}

func TestOldB355EnsoulArtifactMakesAFiveFiveCreatureWhileAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	aura := oldB355Attach(t, g, me.ID, "Ensoul Artifact", oldB355EnsoulArtifact, rock)
	if p, tough := effectstest.PtOf(t, g, rock); p != 5 || tough != 5 {
		t.Errorf("the artifact is %d/%d, want 5/5", p, tough)
	}
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, rock)
		if !c.IsCreature() || !c.IsArtifact() {
			t.Errorf("the enchanted permanent is not an artifact creature: %q", c.TypeLine)
		}
	})
	effectstest.KillCreature(t, g, me.ID, aura)
	g.WithWriteLock(func() {
		if effectstest.FindBattlefieldCardForTest(g, rock).IsCreature() {
			t.Error("the artifact is still a creature without the Aura")
		}
	})
}

func TestOldB355EnsoulArtifactCannotEnchantANonartifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Ensoul Artifact", effectstest.AuraTypeLine, oldB355EnsoulArtifact, effectstest.CardTarget(bear)); err == nil {
		t.Error("Ensoul Artifact was cast on a nonartifact creature")
	}
}

func TestOldB355VerdantHavenGainsLifeAndAddsAnAdditionalMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Verdant Haven", effectstest.AuraTypeLine, oldB355VerdantHaven, effectstest.CardTarget(forest))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	effectstest.TapForMana(t, g, me.ID, forest)
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick for the additional mana")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "U" {
		t.Errorf("pool = %v, want {G} and the chosen {U}", got)
	}
}

func TestOldB355EncrustLocksUntapAndActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	servant := effectstest.B12Push(g, opp.ID, "Air Servant", "Creature — Elemental", oldB355AirServant, 4, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, servant).Tapped = true })
	oldB355Attach(t, g, opp.ID, "Encrust", oldB355Encrust, servant)
	effectstest.AdvanceToUpkeepOfSeat(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if !effectstest.IsTapped(g, servant) {
		t.Error("the enchanted creature untapped")
	}
	// Its activated abilities can't be activated, even with the mana paid.
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, opp, "CCU")
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Keywords: []string{"flying"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	if err := g.ActivateCatalogAbility(opp.ID, servant, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(flier)}); err == nil {
		t.Error("the enchanted creature's ability was activated")
	}
}

func TestOldB355EncrustEnchantsAnArtifactToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Encrust", effectstest.AuraTypeLine, oldB355Encrust, effectstest.CardTarget(land)); err == nil {
		t.Error("Encrust was cast on a land")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Encrust", effectstest.AuraTypeLine, oldB355Encrust, effectstest.CardTarget(rock)); err != nil && !strings.Contains(err.Error(), "mana") {
		t.Errorf("Encrust refused an artifact for a reason other than mana: %v", err)
	}
}

func TestOldB355InvisibilityCanOnlyBeBlockedByWalls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Rogue", 2, 2)
	oldB355Attach(t, g, me.ID, "Invisibility", oldB355Invisibility, runner)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(runner, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(bear, runner), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(wall, runner); err != nil {
		t.Errorf("a Wall could not block: %v", err)
	}
}

func TestOldB355SpectraWardPumpsAndGivesProtectionFromEveryColorWithoutShakingAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Invisibility", effectstest.AuraTypeLine, oldB355Invisibility, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(other, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	ward := oldB355Attach(t, g, me.ID, "Spectra Ward", oldB355SpectraWard, host)
	effectstest.AdvanceToMain(t, g)
	g.RunStateChecksForTest()
	if p, tough := effectstest.PtOf(t, g, host); p != 4 || tough != 4 {
		t.Errorf("the host is %d/%d, want 4/4", p, tough)
	}
	for _, kw := range effects.OdyProtectionFromEachColor() {
		found := false
		for _, a := range effectstest.EffectiveAbilities(t, g, host) {
			if strings.HasPrefix(a, kw) {
				found = true
			}
		}
		if !found {
			t.Errorf("the host lacks %q", kw)
		}
	}
	if !g.Battlefield.Contains(ward) || !g.Battlefield.Contains(other) {
		t.Error("an Aura fell off a creature protected from its colour, though the card says it doesn't remove Auras")
	}
}

func TestOldB355OppressiveRaysTaxesAttackingBlockingAndActivating(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	servant := effectstest.B12Push(g, me.ID, "Air Servant", "Creature — Elemental", oldB355AirServant, 4, 3)
	free := effectstest.B12Creature(g, me.ID, "Free", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	oldB355Attach(t, g, opp.ID, "Oppressive Rays", oldB355OppressiveRays, servant)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	// The attack tax is {3} for the enchanted creature alone.
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: servant, Target: opp.ID}}); p.Total.Generic != 3 {
		t.Errorf("attacking with the enchanted creature costs %d, want 3", p.Total.Generic)
	}
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: free, Target: opp.ID}}); !p.IsFree() {
		t.Errorf("attacking with another creature costs %+v, want nothing", p.Total)
	}
	if err := g.DeclareAttacker(servant, opp.ID); !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("attacking with no mana: %v, want ErrAttackTaxUnpaid", err)
	}
	effectstest.FillPool(me, 3)
	if err := g.DeclareAttacker(servant, opp.ID); err != nil {
		t.Fatalf("attacking with {3}: %v", err)
	}
	_ = blocker
	// Its activated abilities cost {3} more: {2}{U} becomes six.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, servant).Tapped = false })
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Keywords: []string{"flying"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, servant, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(flier)}); err == nil {
		t.Error("Air Servant was activated for its printed cost despite Oppressive Rays")
	}
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, servant, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(flier)}); err != nil {
		t.Errorf("Air Servant was not activatable for {3} more: %v", err)
	}
}
