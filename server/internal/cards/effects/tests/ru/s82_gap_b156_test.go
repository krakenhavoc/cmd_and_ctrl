package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b156_test.go — GAP batch 156 (Jumpstart 2022 red creatures,
// #5948), played through the catalog and the real stack.

const (
	gapB156Flameblade  = "b1c721e6-d0b1-4443-937d-89f7c3a459e6"
	gapB156Whelp       = "8422f1e0-00ca-4ffb-a6b2-e2c9f96d7f23"
	gapB156Grotag      = "623e261c-d4b9-4dea-930d-e0573a1deaee"
	gapB156Ib          = "6742c4b9-8fa6-439f-844d-04a09fa20c45"
	gapB156Flamefiend  = "e93089b1-e790-402b-a994-69a7e422e4f2"
	gapB156LavaSerpent = "4d94d807-97d9-4cdf-a63c-d8b82c2acc6e"
	gapB156Bear        = "Creature — Bear"
)

func TestGapB156FlamebladeAdeptGetsOnePerCycledOrDiscardedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	adept := effectstest.B12Push(g, me.ID, "Flameblade Adept", "Creature — Jackal Warrior", gapB156Flameblade, 1, 2)
	if !effectstest.P1g2HasKeyword(g, adept, "menace") {
		t.Error("the Adept lacks menace")
	}
	serpent := uuid.New()
	plain := uuid.New()
	theirs := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: serpent, Name: "Lava Serpent", TypeLine: "Creature — Elemental Serpent",
			OracleID: gapB156LavaSerpent, ManaCost: "{5}{R}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
		me.Hand.PushTop(game.Card{InstanceID: plain, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
		opp.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Their Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, serpent, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle the Serpent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.ApaLive(g, adept); live.CurrentPower() != 2 {
		t.Fatalf("cycling one card made it %d/%d, want 2/2 (one trigger, not two)", live.CurrentPower(), live.CurrentToughness())
	}
	g.WithWriteLock(func() { _ = g.DiscardCardsKeyedForEffect(me.ID, []uuid.UUID{plain}, uuid.Nil, game.KeyedTail{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.ApaLive(g, adept); live.CurrentPower() != 3 {
		t.Fatalf("a discard too: %d/%d, want 3/2", live.CurrentPower(), live.CurrentToughness())
	}
	g.WithWriteLock(func() { _ = g.DiscardCardsKeyedForEffect(opp.ID, []uuid.UUID{theirs}, uuid.Nil, game.KeyedTail{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.ApaLive(g, adept); live.CurrentPower() != 3 {
		t.Errorf("an opponent's discard pumped it: %d/%d", live.CurrentPower(), live.CurrentToughness())
	}
}

func TestGapB156FurnaceWhelpFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	whelp := effectstest.B12Push(g, me.ID, "Furnace Whelp", "Creature — Dragon", gapB156Whelp, 2, 2)
	if !effectstest.P1g2HasKeyword(g, whelp, "flying") {
		t.Error("the Whelp lacks flying")
	}
	for i := 0; i < 3; i++ {
		effectstest.FloatForTest(g, me, "R")
		effectstest.ApaActivate(t, g, me, whelp, 0, game.ActivateAbilityParams{})
	}
	if live := effectstest.ApaLive(g, whelp); live.CurrentPower() != 5 || live.CurrentToughness() != 2 {
		t.Errorf("after three pumps it is %d/%d, want 5/2", live.CurrentPower(), live.CurrentToughness())
	}
}

func TestGapB156GrotagExilesTheTopCardAndLetsYouPlayItThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	grotag := effectstest.B12Push(g, me.ID, "Grotag Night-Runner", "Creature — Goblin Rogue", gapB156Grotag, 2, 3)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	effectstest.DealCombatDamageToPlayer(g, grotag, opp.ID, 2)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 0 {
		t.Fatalf("the top card was not exiled (library %d)", me.Library.Size())
	}
	if !g.Exile.Contains(top) {
		t.Fatal("the top card is not in exile")
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the exiled land: %v", err)
	}
	if !g.Battlefield.Contains(top) {
		t.Error("the land did not enter the battlefield")
	}
}

func TestGapB156GrotagTriggersOnlyForCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	grotag := effectstest.B12Push(g, me.ID, "Grotag Night-Runner", "Creature — Goblin Rogue", gapB156Grotag, 2, 3)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	effectstest.DealNoncombatDamage(g, me.ID, grotag, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Error("noncombat damage exiled the top card")
	}
}

func TestGapB156IbSacrificesABlockedGoblinToHitItsBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ib Halfheart, Goblin Tactician", "Legendary Creature — Goblin Advisor", gapB156Ib, false)
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin Warrior", 2, 2)
	a := effectstest.PushVanillaCreature(g, them.ID, "Wall A", 0, 5)
	b := effectstest.PushVanillaCreature(g, them.ID, "Wall B", 0, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(goblin, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(w, goblin); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	if g.Battlefield.Contains(goblin) {
		t.Error("the blocked Goblin was not sacrificed")
	}
	for _, w := range []uuid.UUID{a, b} {
		if got := effectstest.FindBattlefieldCardForTest(g, w).DamageMarked; got != 4 {
			t.Errorf("a blocker took %d damage, want 4", got)
		}
	}
}

func TestGapB156IbIgnoresNonGoblinsAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	ib := effectstest.PushCatalogPermanent(g, me.ID, "Ib Halfheart, Goblin Tactician", "Legendary Creature — Goblin Advisor", gapB156Ib, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", gapB156Bear, 2, 2)
	wall := effectstest.PushVanillaCreature(g, them.ID, "Wall", 0, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{ib, bear} {
		if err := g.DeclareAttacker(id, them.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, ib); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	if !g.Battlefield.Contains(ib) {
		t.Error("Ib sacrificed itself")
	}
	if got := effectstest.FindBattlefieldCardForTest(g, wall).DamageMarked; got >= 4 {
		t.Errorf("the wall took %d damage: Ib's 4 damage was dealt for a non-Goblin or for itself", got)
	}
}

func TestGapB156IbSacrificesTwoMountainsForTwoGoblins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ib := effectstest.B12Push(g, me.ID, "Ib Halfheart, Goblin Tactician", "Legendary Creature — Goblin Advisor", gapB156Ib, 3, 2)
	m1 := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	m2 := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, ib, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{m1, forest}}); err == nil {
		t.Fatal("sacrificed a Forest for a Mountain")
	}
	effectstest.ApaActivate(t, g, me, ib, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{m1, m2}})
	if g.Battlefield.Contains(m1) || g.Battlefield.Contains(m2) {
		t.Error("the Mountains were not sacrificed")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Goblin"); n != 2 {
		t.Errorf("%d Goblin tokens, want 2", n)
	}
}

func TestGapB156FlamefiendDealsFourDividedOnlyIfAnArtifactIsSacrificed(t *testing.T) {
	for _, sacrifice := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
		rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
		effectstest.CastCatalogSpell(t, g, "Kuldotha Flamefiend", "Creature — Elemental", gapB156Flamefiend, nil)
		effectstest.PassPriorityAroundTable(t, g)
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			t.Fatal("the enters trigger asks for targets")
		}
		life := opp.Life
		refs := append(effectstest.CardRefs(a), game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
		if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{a: 1, opp.ID: 3}); err != nil {
			t.Fatalf("a legal division: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		var ask *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceOwnPermanents && pc.Chooser == me.ID {
				ask = pc
			}
		}
		if ask == nil {
			t.Fatal("the Flamefiend did not ask whether to sacrifice an artifact")
		}
		picks := []uuid.UUID(nil)
		if sacrifice {
			picks = []uuid.UUID{rock}
		}
		if err := g.ResolveOwnPermanents(ask.ID, me.ID, picks); err != nil {
			t.Fatalf("answer the sacrifice prompt: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		d, lifeNow := effectstest.FindBattlefieldCardForTest(g, a).DamageMarked, opp.Life
		if sacrifice && (d != 1 || lifeNow != life-3 || g.Battlefield.Contains(rock)) {
			t.Errorf("sacrificed: %d marked, life %d → %d, rock still there %v; want 1, -3, gone", d, life, lifeNow, g.Battlefield.Contains(rock))
		}
		if !sacrifice && (d != 0 || lifeNow != life || !g.Battlefield.Contains(rock)) {
			t.Errorf("declined: %d marked, life %d → %d, rock there %v; want nothing", d, life, lifeNow, g.Battlefield.Contains(rock))
		}
	}
}
