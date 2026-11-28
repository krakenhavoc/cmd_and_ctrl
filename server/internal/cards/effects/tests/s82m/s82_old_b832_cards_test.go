package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b832_cards_test.go — Ice Age, batch 832 (S82 OLD, #5636): activated
// abilities. Every test plays the real catalog entry through the engine.

const (
	oldB832Ranger    = "4d47e598-690e-4850-97e4-7136bc7b497e"
	oldB832Jarkeld   = "64df77f6-4f7e-442a-971c-8ad3f7d1c979"
	oldB832Shield    = "65164450-b477-4096-8146-ebb80a709787"
	oldB832Brownie   = "6d1f8073-4da2-4abc-abb4-2053c4a40bbf"
	oldB832Baton     = "879011c6-550e-4e33-a664-fffb655321d9"
	oldB832Crevasses = "88e0551a-ada0-41d9-b5c3-39257ce56c3e"
	oldB832Druid     = "943ab696-470b-488f-9567-c8cd89ec06e3"
	oldB832Chariot   = "989c698a-600e-47d8-acaf-3ca140dcd150"
	oldB832Sappers   = "9e095761-cdae-4ab2-b2d3-099b0b39600a"
	oldB832Amulet    = "a06f8476-0a23-413a-bc28-8dc2427fdc2a"
	oldB832Bear      = "Creature — Bear"
)

func TestOldB832EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB832Ranger: "Kelsinko Ranger", oldB832Jarkeld: "General Jarkeld", oldB832Shield: "Shield of the Ages",
		oldB832Brownie: "Fyndhorn Brownie", oldB832Baton: "Baton of Morale", oldB832Crevasses: "Glacial Crevasses",
		oldB832Druid: "Juniper Order Druid", oldB832Chariot: "War Chariot", oldB832Sappers: "Goblin Sappers",
		oldB832Amulet: "Amulet of Quoz",
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

func oldB832Activate(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	if mana != "" {
		effectstest.FloatMana(t, g, me, mana)
	}
	params.Strict = true
	err := g.ActivateCatalogAbility(me.ID, card, idx, params)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB832KelsinkoRangerGivesAGreenCreatureFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ranger := effectstest.B12Push(g, me.ID, "Kelsinko Ranger", "Creature — Human Ranger", oldB832Ranger, 1, 1)
	green := effectstest.B12Creature(g, me.ID, "Green Bear", oldB832Bear, 2, 2)
	effectstest.AlliSetColors(g, green, "G")
	red := effectstest.B12Creature(g, me.ID, "Red Bear", oldB832Bear, 2, 2)
	effectstest.AlliSetColors(g, red, "R")
	if err := oldB832Activate(t, g, me, ranger, 0, "{C}{W}", game.ActivateAbilityParams{Targets: effectstest.CardTargets(red)}); err == nil {
		t.Fatal("targeted a red creature")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := oldB832Activate(t, g, me, ranger, 0, "{C}{W}", game.ActivateAbilityParams{Targets: effectstest.CardTargets(green)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, green, "first strike") {
		t.Error("the green creature did not gain first strike")
	}
}

func TestOldB832GeneralJarkeldSwitchesTheBlockersOfTwoBlockedAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jarkeld := effectstest.B12Push(g, me.ID, "General Jarkeld", "Legendary Creature — Human Soldier", oldB832Jarkeld, 1, 2)
	a := effectstest.B12Creature(g, me.ID, "Attacker A", oldB832Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", oldB832Bear, 2, 2)
	x := effectstest.B12Creature(g, opp.ID, "Blocker X", oldB832Bear, 2, 2)
	y := effectstest.B12Creature(g, opp.ID, "Blocker Y", oldB832Bear, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, c := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(c, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(x, a); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareBlocker(y, b); err != nil {
		t.Fatal(err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	if err := oldB832Activate(t, g, me, jarkeld, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTargets(a, b)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	cx, _ := effectstest.BattlefieldCard(g, x)
	cy, _ := effectstest.BattlefieldCard(g, y)
	if !cx.IsBlockingAttacker(b) || cx.IsBlockingAttacker(a) {
		t.Error("Blocker X did not switch to Attacker B")
	}
	if !cy.IsBlockingAttacker(a) || cy.IsBlockingAttacker(b) {
		t.Error("Blocker Y did not switch to Attacker A")
	}
}

func TestOldB832GeneralJarkeldOnlyActsInTheDeclareBlockersStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	jarkeld := effectstest.B12Push(g, me.ID, "General Jarkeld", "Legendary Creature — Human Soldier", oldB832Jarkeld, 1, 2)
	a := effectstest.B12Creature(g, me.ID, "A", oldB832Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB832Bear, 2, 2)
	if err := oldB832Activate(t, g, me, jarkeld, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTargets(a, b)}); err == nil {
		t.Error("activated in the main phase on creatures that are not blocked attackers")
	}
}

func TestOldB832ShieldOfTheAgesPreventsTheNextOneDamageToYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	shield := effectstest.B12Push(g, me.ID, "Shield of the Ages", "Artifact", oldB832Shield, 0, 0)
	if err := oldB832Activate(t, g, me, shield, 0, "{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	life := effectstest.LifeOf(g, me.ID)
	bolt := effectstest.PutInHand(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Colors: []string{"R"}, ManaCost: "{R}"})
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life-2 {
		t.Errorf("life %d, want %d (3 damage less the 1 prevented)", got, life-2)
	}
}

func TestOldB832FyndhornBrownieUntapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	brownie := effectstest.B12Push(g, me.ID, "Fyndhorn Brownie", "Creature — Ouphe", oldB832Brownie, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB832Bear, 2, 2)
	effectstest.TapForTest(t, g, bear)
	if err := oldB832Activate(t, g, me, brownie, 0, "{C}{C}{G}", game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the Bear is still tapped")
	}
	if !effectstest.B16Tapped(t, g, brownie) {
		t.Error("the Brownie was not tapped for its cost")
	}
}

func TestOldB832BatonOfMoraleGrantsBanding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	baton := effectstest.B12Push(g, me.ID, "Baton of Morale", "Artifact", oldB832Baton, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB832Bear, 2, 2)
	if err := oldB832Activate(t, g, me, baton, 0, "{C}{C}", game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, game.KeywordBanding) {
		t.Error("the Bear did not gain banding")
	}
}

func TestOldB832WarChariotGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	chariot := effectstest.B12Push(g, me.ID, "War Chariot", "Artifact", oldB832Chariot, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB832Bear, 2, 2)
	if err := oldB832Activate(t, g, me, chariot, 0, "{C}{C}{C}", game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
	if !effectstest.B16Tapped(t, g, chariot) {
		t.Error("the Chariot was not tapped for its cost")
	}
}

func TestOldB832GlacialCrevassesSacrificesASnowMountainToFogCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crevasses := effectstest.B12Push(g, me.ID, "Glacial Crevasses", "Enchantment", oldB832Crevasses, 0, 0)
	forest := effectstest.SeedLand(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "")
	mountain := effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	snowMountain := effectstest.SeedLand(g, me.ID, "Snow-Covered Mountain", "Basic Snow Land — Mountain", "")
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB832Bear, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	life := effectstest.LifeOf(g, opp.ID)
	for _, bad := range []uuid.UUID{forest, mountain} {
		if err := g.ActivateCatalogAbility(me.ID, crevasses, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bad}}); err == nil {
			t.Fatal("sacrificed something that is not a snow Mountain")
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, crevasses, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{snowMountain}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, snowMountain) {
		t.Error("the snow Mountain was not sacrificed")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("the defender lost %d life; combat damage was not prevented", life-got)
	}
}

func TestOldB832JuniperOrderDruidUntapsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	druid := effectstest.B12Push(g, me.ID, "Juniper Order Druid", "Creature — Human Cleric Druid", oldB832Druid, 1, 1)
	land := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.TapForTest(t, g, land)
	if err := oldB832Activate(t, g, me, druid, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTargets(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.B16Tapped(t, g, land) {
		t.Error("the land is still tapped")
	}
}

func oldB832SappersTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sappers := effectstest.B12Push(g, me.ID, "Goblin Sappers", "Creature — Goblin", oldB832Sappers, 1, 1)
	ally := effectstest.B12Creature(g, me.ID, "Ally", oldB832Bear, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB832Bear, 2, 2)
	return g, me, opp, sappers, ally, blocker
}

func TestOldB832GoblinSappersMakeACreatureUnblockableAndDestroyItAtEndOfCombat(t *testing.T) {
	for idx, tc := range []struct {
		mana         string
		sappersDies  bool
		sappersLabel string
	}{{"{R}{R}", true, "cheap"}, {"{R}{R}{R}{R}", false, "dear"}} {
		g, me, opp, sappers, ally, blocker := oldB832SappersTable(t)
		effectstest.AdvanceToMain(t, g)
		if err := oldB832Activate(t, g, me, sappers, idx, tc.mana, game.ActivateAbilityParams{Targets: effectstest.CardTargets(ally)}); err != nil {
			t.Fatalf("%s: activate: %v", tc.sappersLabel, err)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(ally, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(blocker, ally); err == nil {
			t.Errorf("%s: the creature was blocked", tc.sappersLabel)
		}
		effectstest.AdvanceTo(t, g, game.StepEndCombat)
		effectstest.PassPriorityAroundTable(t, g)
		if effects.OnBattlefield(g, ally) {
			t.Errorf("%s: the creature survived end of combat", tc.sappersLabel)
		}
		if alive := effects.OnBattlefield(g, sappers); alive == tc.sappersDies {
			t.Errorf("%s: the Sappers on the battlefield = %v, want %v", tc.sappersLabel, alive, !tc.sappersDies)
		}
	}
}

func TestOldB832AmuletOfQuozFlipsACoinAndOneSideLosesTheGame(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 60 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		amulet := effectstest.B12Push(g, me.ID, "Amulet of Quoz", "Artifact", oldB832Amulet, 0, 0)
		if err := g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
			t.Fatal("activated outside the controller's upkeep")
		}
		for i := 0; i < 300; i++ {
			if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 && i > 0 {
				break
			}
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.UntapForTest(g, amulet)
		if err := g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
			t.Fatalf("activate in the upkeep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) == 0 {
			t.Fatal("no coin was flipped")
		}
		if effects.OnBattlefield(g, amulet) {
			t.Error("the Amulet was not sacrificed")
		}
		meOut, oppOut := me.Eliminated, opp.Eliminated
		if evs[len(evs)-1].Won {
			won++
			if !oppOut || meOut {
				t.Errorf("a won flip: opponent lost=%v, you lost=%v", oppOut, meOut)
			}
		} else {
			lost++
			if oppOut || !meOut {
				t.Errorf("a lost flip: opponent lost=%v, you lost=%v", oppOut, meOut)
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("saw %d won and %d lost flips; want both", won, lost)
	}
}
