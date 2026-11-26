package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 490 (Worldwake, #5582): Sejiri Steppe, Terra Eternal and
// Razor Boomerang.

const (
	oldB490Steppe    = "3dfbf95e-a91b-429c-96e2-95ac777e7027"
	oldB490Terra     = "a5ce8125-da02-48b0-83cd-871c934352b0"
	oldB490Boomerang = "f2649e48-4252-43f3-bd1e-c4c6fcb0aa3a"
)

func TestOldB490EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB490Steppe: "Sejiri Steppe", oldB490Terra: "Terra Eternal", oldB490Boomerang: "Razor Boomerang",
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

func TestOldB490SejiriSteppeEntersTappedAndGivesProtectionFromAChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	steppe := effectstest.PlayLandFromHand(t, g, "Sejiri Steppe", oldB490Steppe)
	if !effectstest.TappedForTest(t, g, steppe) {
		t.Error("Sejiri Steppe did not enter tapped")
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		effectstest.PassPriorityUntilChoice(t, g)
		p = effectstest.LatestPickTarget(g, me.ID)
	}
	if p == nil {
		t.Fatalf("no target prompt for the enters trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("pick the creature: %v", err)
	}
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "G")
	c, _ := effectstest.BattlefieldCard(g, bear)
	if !game.HasKeyword(&c, game.ProtectionFromColor("G")) {
		t.Error("the creature did not gain protection from green")
	}
	if game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("the creature gained protection from a colour that was not chosen")
	}
}

func TestOldB490SejiriSteppeTapsForWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	steppe := effectstest.PushCatalogPermanent(g, me.ID, "Sejiri Steppe", "Land", oldB490Steppe, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, steppe, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool %d, want one mana", got)
	}
	if !effectstest.TappedForTest(t, g, steppe) {
		t.Error("the land did not tap")
	}
}

func TestOldB490TerraEternalMakesEveryLandIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	creature := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	baseline := effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.B15Destroy(t, g, baseline)
	if g.Battlefield.Contains(baseline) {
		t.Fatal("setup: a land survived destroy before Terra Eternal was out")
	}
	effectstest.CastCatalogSpell(t, g, "Terra Eternal", "Enchantment", oldB490Terra, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, land := range []uuid.UUID{mine, theirs} {
		if !effectstest.HasEffectiveKeyword(t, g, land, "indestructible") {
			t.Errorf("land %s lacks indestructible", land)
		}
	}
	effectstest.B15Destroy(t, g, theirs)
	if !g.Battlefield.Contains(theirs) {
		t.Error("an opponent's land was destroyed under Terra Eternal")
	}
	if effectstest.HasEffectiveKeyword(t, g, creature, "indestructible") {
		t.Error("a nonland creature gained indestructible")
	}
}

func oldB490Attach(t *testing.T, g *game.Game, attachment, host uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: host})
	})
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
}

func TestOldB490RazorBoomerangThrowsAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	boomerang := effectstest.SeedEquipment(g, me.ID, "Razor Boomerang", oldB490Boomerang)
	oldB490Attach(t, g, boomerang, bear)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.AttachedTo(t, g, boomerang, bear) {
		t.Error("the Boomerang was not unattached as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if !me.Hand.Contains(boomerang) || effectstest.BattlefieldHas(g, boomerang) {
		t.Error("the Boomerang did not return to its owner's hand")
	}
	if !effectstest.TappedOnBattlefield(t, g, bear) {
		t.Error("the creature was not tapped")
	}
}

func TestOldB490RazorBoomerangRefusesWhileTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	boomerang := effectstest.SeedEquipment(g, me.ID, "Razor Boomerang", oldB490Boomerang)
	oldB490Attach(t, g, boomerang, bear)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("activated while tapped")
	}
	if !effectstest.AttachedTo(t, g, boomerang, bear) {
		t.Error("a refused activation unattached the Boomerang")
	}
}

func TestOldB490RazorBoomerangEquipsForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)
	boomerang := effectstest.SeedEquipment(g, me.ID, "Razor Boomerang", oldB490Boomerang)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, boomerang, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("equipped for {1}")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, boomerang, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("equip for {2}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, boomerang, bear) {
		t.Error("the Boomerang did not attach")
	}
}
