package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 older sets, Commander 2013, batch 385 (#5560).

const (
	oldB385SkywardEyeOracle = "45bef776-121b-4489-9c46-f7b4fd4c3c0d"
	oldB385MayaelOracle     = "7f546d54-584d-4bec-8fbb-1ea2f8ab277e"
	oldB385WarCadenceOracle = "49d0fdd6-cc8f-4fe1-a6bd-4321dac18404"
	oldB385DjinnOracle      = "f9de4cea-27c4-4343-8a7a-09b8f346c3b5"
	oldB385RakeclawOracle   = "8dbb4a8f-78e9-4ceb-824d-bb67bdf939db"
)

func TestOldB385EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB385SkywardEyeOracle: "Skyward Eye Prophets",
		oldB385MayaelOracle:     "Mayael the Anima",
		oldB385WarCadenceOracle: "War Cadence",
		oldB385DjinnOracle:      "Djinn of Infinite Deceits",
		oldB385RakeclawOracle:   "Rakeclaw Gargantuan",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q completeness %v", name, spec.Name, spec.Completeness)
		}
	}
}

func TestOldB385SkywardEyeProphetsPutsALandOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prophets := effectstest.B12Push(g, me.ID, "Skyward Eye Prophets", "Creature — Human Wizard", oldB385SkywardEyeOracle, 3, 3)
	land := effectstest.B02TopOfLibrary(me, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, prophets, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("the revealed land did not enter the battlefield")
	}
	if !effectstest.HasEffectiveKeyword(t, g, prophets, "vigilance") {
		t.Error("Skyward Eye Prophets lacks vigilance")
	}
}

func TestOldB385SkywardEyeProphetsPutsANonlandIntoHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prophets := effectstest.B12Push(g, me.ID, "Skyward Eye Prophets", "Creature — Human Wizard", oldB385SkywardEyeOracle, 3, 3)
	spell := effectstest.B02TopOfLibrary(me, "Bear Cub", "Creature — Bear")
	if err := g.ActivateCatalogAbility(me.ID, prophets, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spell) {
		t.Error("a nonland card entered the battlefield")
	}
	if !me.Hand.Contains(spell) {
		t.Error("the revealed nonland card did not go to hand")
	}
}

func oldB385MayaelTable(t *testing.T) (g *game.Game, me *game.Player, mayael uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	mayael = effectstest.B12Push(g, me.ID, "Mayael the Anima", "Legendary Creature — Elf Shaman", oldB385MayaelOracle, 2, 3)
	return g, me, mayael
}

func TestOldB385MayaelPutsAPowerFiveCreatureOntoTheBattlefield(t *testing.T) {
	g, me, mayael := oldB385MayaelTable(t)
	big := uuid.New()
	small := uuid.New()
	for i := 0; i < 4; i++ {
		effectstest.B02TopOfLibrary(me, "Filler", "Land")
	}
	me.Library.PushTop(game.Card{InstanceID: small, Name: "Four Power", TypeLine: "Creature — Ox", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: big, Name: "Five Power", TypeLine: "Creature — Ox", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "C", "C", "C", "R", "G", "W")
	if err := g.ActivateCatalogAbility(me.ID, mayael, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsPromptFrom(g, mayael)
	if prompt == nil {
		t.Fatal("no pick offered")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{small}); err == nil {
		t.Fatal("a creature with power 4 was accepted")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{big}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("the power-5 creature did not enter")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the power-4 creature entered")
	}
}

func TestOldB385MayaelRefusesWithoutAllItsMana(t *testing.T) {
	g, me, mayael := oldB385MayaelTable(t)
	if err := effectstest.ConB1Activate(t, g, me, mayael, 0, "CCCRG", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without {W}")
	}
}

func TestOldB385RakeclawTargetsOnlyPowerFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	garg := effectstest.B12Push(g, me.ID, "Rakeclaw Gargantuan", "Creature — Beast", oldB385RakeclawOracle, 5, 3)
	bear := effectstest.SeedBear(g, me.ID)
	if err := effectstest.ConB1Activate(t, g, me, garg, 0, "C", game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("a 2/2 was a legal target")
	}
	if err := effectstest.ConB1Activate(t, g, me, garg, 0, "C", game.ActivateAbilityParams{Targets: effectstest.CardTarget(garg)}); err != nil {
		t.Fatalf("activate on itself: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, garg, "first strike") {
		t.Error("the Gargantuan did not gain first strike")
	}
}

func TestOldB385DjinnExchangesTwoNonlegendaryCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	djinn := effectstest.B12Push(g, me.ID, "Djinn of Infinite Deceits", "Creature — Djinn", oldB385DjinnOracle, 2, 7)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	legend := effectstest.SeedLegendaryCreature(g, opp.ID, "Some Legend")
	if err := effectstest.ConB1Activate(t, g, me, djinn, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTargets(mine, legend)}); err == nil {
		t.Fatal("a legendary creature was a legal target")
	}
	if err := effectstest.ConB1Activate(t, g, me, djinn, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTargets(mine, theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, mine) != opp.ID || effectstest.ControllerOf(t, g, theirs) != me.ID {
		t.Error("control was not exchanged")
	}
}

func TestOldB385DjinnCantBeActivatedDuringCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	djinn := effectstest.B12Push(g, me.ID, "Djinn of Infinite Deceits", "Creature — Djinn", oldB385DjinnOracle, 2, 7)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := effectstest.ConB1Activate(t, g, me, djinn, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTargets(mine, theirs)}); err == nil {
		t.Fatal("activated during combat")
	}
}

func TestOldB385WarCadenceTaxesEachBlockerForX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cadence := effectstest.SeedEnchantment(g, oldB385WarCadenceOracle, "War Cadence", "Enchantment", me.ID)
	wall := effectstest.SeedBear(g, opp.ID)
	effectstest.B06AddMana(me, "C", "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, cadence, 0, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var price game.BlockTaxPrice
	g.WithWriteLock(func() { price = g.PriceBlockersForEffect([]uuid.UUID{wall}) })
	if price.Cost != "{2}" || len(price.Lines) != 1 {
		t.Fatalf("price = %+v, want {2} for one blocker", price)
	}
}
