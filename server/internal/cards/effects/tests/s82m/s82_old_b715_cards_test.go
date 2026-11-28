package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b715_cards_test.go — Urza's Legacy, pump and combat tricks, batch
// 715 (S82, #5623). Each test plays the real catalog entry through the engine.

const (
	oldB715WarCryOracle    = "0439df39-0324-4110-b3f4-4a32393021de"
	oldB715SickOracle      = "070d82db-0851-4f6d-8bf7-1a79c4c2cfee"
	oldB715SustainerOracle = "2c99d52d-f909-4494-aeee-937ae779fd4a"
	oldB715HopeOracle      = "506ea048-aca5-439a-a5b5-059ddda0672a"
	oldB715SilkNetOracle   = "58511cb8-348d-409e-8fd6-772be20e57cd"
	oldB715PyrosaurOracle  = "667c39bc-92d6-4fee-bced-95f0e1f21455"
	oldB715RadiantOracle   = "859cbe11-e1e6-4309-94da-8725db1b3fc2"
	oldB715ElkOracle       = "b99d6d51-3ead-4b81-8e0c-f458619ebb24"
)

func TestOldB715EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB715WarCryOracle: "Ghitu War Cry", oldB715SickOracle: "Sick and Tired",
		oldB715SustainerOracle: "Sustainer of the Realm", oldB715HopeOracle: "Hope and Glory",
		oldB715SilkNetOracle: "Silk Net", oldB715PyrosaurOracle: "Pygmy Pyrosaur",
		oldB715RadiantOracle: "Radiant, Archangel", oldB715ElkOracle: "Gang of Elk",
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

func oldB715Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle, mana string, targets ...uuid.UUID) error {
	t.Helper()
	effectstest.FloatForTest(g, p, mana)
	var refs []game.TargetRef
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: refs})
}

func TestOldB715GhituWarCryPumpsTheTargetEachTimeItIsActivated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cry := effectstest.PushCatalogPermanent(g, me.ID, "Ghitu War Cry", "Enchantment", oldB715WarCryOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	for i := 0; i < 2; i++ {
		effectstest.FloatForTest(g, me, "R")
		if err := g.ActivateCatalogAbility(me.ID, cry, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tt := effectstest.MkmMPPT(t, g, bear); p != 4 || tt != 2 {
		t.Errorf("Bear is %d/%d, want 4/2", p, tt)
	}
	if err := g.ActivateCatalogAbility(me.ID, cry, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Error("activated with no {R} to pay")
	}
}

func TestOldB715SickAndTiredShrinksBothTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	if err := oldB715Cast(t, g, me, "Sick and Tired", "Instant", "{2}{B}", oldB715SickOracle, "CCB", a); err == nil {
		t.Fatal("Sick and Tired cast with one target")
	}
	if err := oldB715Cast(t, g, me, "Sick and Tired", "Instant", "{2}{B}", oldB715SickOracle, "CCB", a, b); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 2, b: 1, c: 2} {
		if p, tt := effectstest.MkmMPPT(t, g, id); p != want || tt != want {
			t.Errorf("a creature is %d/%d, want %d/%d", p, tt, want, want)
		}
	}
}

func TestOldB715SustainerOfTheRealmGainsToughnessWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	sustainer := effectstest.B12Push(g, opp.ID, "Sustainer of the Realm", "Creature — Angel", oldB715SustainerOracle, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, sustainer, "flying") {
		t.Error("no flying")
	}
	atk := effectstest.B12Creature(g, g.Seats[0].ID, "Attacker", "Creature — Bear", 1, 1)
	effectstest.RampageBlock(t, g, atk, sustainer)
	if p, tt := effectstest.MkmMPPT(t, g, sustainer); p != 2 || tt != 5 {
		t.Errorf("Sustainer is %d/%d, want 2/5", p, tt)
	}
}

func TestOldB715HopeAndGloryUntapsAndPumpsBothTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.B02bSetTapped(g, a, true)
	effectstest.B02bSetTapped(g, b, true)
	if err := oldB715Cast(t, g, me, "Hope and Glory", "Instant", "{1}{W}", oldB715HopeOracle, "CW", a); err == nil {
		t.Fatal("Hope and Glory cast with one target")
	}
	if err := oldB715Cast(t, g, me, "Hope and Glory", "Instant", "{1}{W}", oldB715HopeOracle, "CW", a, b); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.TappedOnBattlefield(t, g, id) {
			t.Error("a target stayed tapped")
		}
		if p, tt := effectstest.MkmMPPT(t, g, id); p != 3 || tt != 3 {
			t.Errorf("a target is %d/%d, want 3/3", p, tt)
		}
	}
}

func TestOldB715SilkNetGivesPlusOneAndReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB715Cast(t, g, me, "Silk Net", "Instant", "{G}", oldB715SilkNetOracle, "G", bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.MkmMPPT(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("Bear is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("no reach")
	}
}

func TestOldB715PygmyPyrosaurCantBlockAndFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	dino := effectstest.PushCatalogPermanent(g, opp.ID, "Pygmy Pyrosaur", "Creature — Lizard", oldB715PyrosaurOracle, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(dino, atk); err == nil {
		t.Error("the Pyrosaur blocked")
	}
	for i := 0; i < 2; i++ {
		effectstest.FloatForTest(g, opp, "R")
		if err := g.ActivateCatalogAbility(opp.ID, dino, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tt := effectstest.MkmMPPT(t, g, dino); p != 3 || tt != 1 {
		t.Errorf("Pyrosaur is %d/%d, want 3/1", p, tt)
	}
}

func TestOldB715RadiantCountsEveryOtherFlyerOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	radiant := effectstest.B12Push(g, me.ID, "Radiant, Archangel", "Legendary Creature — Angel", oldB715RadiantOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, radiant, "flying") || !effectstest.HasEffectiveKeyword(t, g, radiant, "vigilance") {
		t.Error("missing flying or vigilance")
	}
	if p, tt := effectstest.MkmMPPT(t, g, radiant); p != 3 || tt != 3 {
		t.Errorf("alone Radiant is %d/%d, want 3/3 (it does not count itself)", p, tt)
	}
	effectstest.PorFlyer(g, me.ID, "Mine")
	effectstest.PorFlyer(g, opp.ID, "Theirs")
	effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Bear", 2, 2)
	if p, tt := effectstest.MkmMPPT(t, g, radiant); p != 5 || tt != 5 {
		t.Errorf("Radiant is %d/%d, want 5/5 (two other flyers, either controller)", p, tt)
	}
}

func TestOldB715GangOfElkGainsTwoTwoForEachBlocker(t *testing.T) {
	for _, n := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		elk := effectstest.B12Push(g, g.Seats[0].ID, "Gang of Elk", "Creature — Elk Beast", oldB715ElkOracle, 5, 4)
		effectstest.RampageBlock(t, g, elk, effectstest.RampageBlockers(g, n)...)
		want := 5 + 2*n
		if p, tt := effectstest.MkmMPPT(t, g, elk); p != want || tt != 4+2*n {
			t.Errorf("%d blockers: Gang of Elk is %d/%d, want %d/%d", n, p, tt, want, 4+2*n)
		}
	}
}

func TestOldB715GangOfElkUnblockedGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	elk := effectstest.B12Push(g, g.Seats[0].ID, "Gang of Elk", "Creature — Elk Beast", oldB715ElkOracle, 5, 4)
	effectstest.AttackIntoBlocks(t, g, g.Seats[1].ID, elk)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.MkmMPPT(t, g, elk); p != 5 || tt != 4 {
		t.Errorf("an unblocked Gang of Elk is %d/%d, want 5/4", p, tt)
	}
}
