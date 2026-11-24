package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b068_cards_test.go — Kaldheim lands (S82, set OLD batch 068).

const (
	oldB068Gnottvold  = "2e413d18-2c55-49be-9e49-54d170c978a2"
	oldB068Immersturm = "354a7376-fb4b-424d-8964-93727302dccb"
	oldB068Gates      = "6f85c26e-3c87-4112-ad1a-8a5708555a93"
	oldB068Skemfar    = "70965b80-c8ad-4718-ae20-12a4d8228898"
	oldB068GreatHall  = "75d8ef26-b745-448b-94b9-3c64668ce171"
	oldB068Frostpyre  = "965aa666-3919-4053-8584-b773bdd54f0b"
	oldB068Haven      = "f74107d5-fb4a-464b-9251-42b84d91775d"
)

var oldB068Names = map[string]string{
	oldB068Gnottvold:  "Gnottvold Slumbermound",
	oldB068Immersturm: "Immersturm Skullcairn",
	oldB068Gates:      "Gates of Istfell",
	oldB068Skemfar:    "Skemfar Elderhall",
	oldB068GreatHall:  "Great Hall of Starnheim",
	oldB068Frostpyre:  "Surtland Frostpyre",
	oldB068Haven:      "Faceless Haven",
}

func TestOldB068EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB068Names {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func TestOldB068TheSixDualPurposeLandsEnterTapped(t *testing.T) {
	for oracle, name := range oldB068Names {
		if oracle == oldB068Haven {
			continue
		}
		g := effectstest.NewCatalogGame(t)
		id := effectstest.PlayLandFromHand(t, g, name, oracle)
		if c, ok := effectstest.BattlefieldCard(g, id); !ok || !c.Tapped {
			t.Errorf("%s did not enter tapped", name)
		}
	}
}

func TestOldB068FacelessHavenEntersUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Faceless Haven", oldB068Haven)
	if c, ok := effectstest.BattlefieldCard(g, id); !ok || c.Tapped {
		t.Error("Faceless Haven entered tapped")
	}
}

func oldB068Land(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, name, "Land", oracle, false)
}

func oldB068Activate(g *game.Game, me *game.Player, land uuid.UUID, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, land, 0, params)
}

func TestOldB068GnottvoldSlumbermoundDestroysALandAndMakesATroll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := oldB068Land(g, me.ID, "Gnottvold Slumbermound", oldB068Gnottvold)
	prize := effectstest.B31Push(g, opp.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: prize}}
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{Targets: targets}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FloatMana(t, g, me, "{R}{R}{R}{R}{G}{G}{G}")
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("targeted a creature that is not a land")
	}
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.BattlefieldHas(g, land) {
		t.Error("the land was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, prize) {
		t.Error("the target land was not destroyed")
	}
	trolls := effectstest.BattlefieldIDsNamed(g, "Troll Warrior")
	if len(trolls) != 1 {
		t.Fatalf("%d Troll Warriors, want 1", len(trolls))
	}
	if p, tough := effectstest.EffectivePTOf(t, g, trolls[0]); p != 4 || tough != 4 {
		t.Errorf("Troll P/T %d/%d, want 4/4", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, trolls[0], "trample") {
		t.Error("the Troll has no trample")
	}
}

func TestOldB068ImmersturmSkullcairnBurnsAndDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := oldB068Land(g, me.ID, "Immersturm Skullcairn", oldB068Immersturm)
	spare := effectstest.HandCardFull(opp, "Spare", "Instant", "", "", nil)
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}{R}{R}{R}")
	life, hand := opp.Life, opp.Hand.Size()
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, spare)
	}
	if opp.Hand.Size() != hand-1 || opp.Hand.Contains(spare) {
		t.Errorf("the target player's hand %d -> %d, want one discarded", hand, opp.Hand.Size())
	}
}

func TestOldB068ImmersturmSkullcairnIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := oldB068Land(g, me.ID, "Immersturm Skullcairn", oldB068Immersturm)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	effectstest.FloatMana(t, g, me, "{B}{B}{R}{R}{R}")
	err := oldB068Activate(g, me, land, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if err == nil {
		t.Fatal("activated in the upkeep")
	}
}

func TestOldB068GatesOfIstfellGainsLifeAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := oldB068Land(g, me.ID, "Gates of Istfell", oldB068Gates)
	effectstest.StockLibraryFor(g, me, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{W}{W}{W}{U}{U}")
	life, hand := me.Life, me.Hand.Size()
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || me.Hand.Size() != hand+2 {
		t.Errorf("life %d -> %d, hand %d -> %d; want +2 and +2", life, me.Life, hand, me.Hand.Size())
	}
	if effectstest.BattlefieldHas(g, land) {
		t.Error("the land was not sacrificed")
	}
}

func TestOldB068SkemfarElderhallShrinksAndMakesTwoElves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := oldB068Land(g, me.ID, "Skemfar Elderhall", oldB068Skemfar)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}{G}{G}")
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Fatal("targeted a creature the activator controls")
	}
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, small) {
		t.Error("the 2/2 target survived -2/-2")
	}
	if n := len(effectstest.BattlefieldIDsNamed(g, "Elf Warrior")); n != 2 {
		t.Errorf("%d Elf Warriors, want 2", n)
	}
}

func TestOldB068SkemfarElderhallWorksWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := oldB068Land(g, me.ID, "Skemfar Elderhall", oldB068Skemfar)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}{G}{G}")
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Elf Warrior")); n != 2 {
		t.Errorf("%d Elf Warriors, want 2", n)
	}
}

func TestOldB068GreatHallSacrificesAnAngelsWorthOfCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := oldB068Land(g, me.ID, "Great Hall of Starnheim", oldB068GreatHall)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{W}{W}{B}")
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no creature to sacrifice")
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.BattlefieldHas(g, bear) || effectstest.BattlefieldHas(g, land) {
		t.Error("the creature and the land were not both sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	angels := effectstest.BattlefieldIDsNamed(g, "Angel Warrior")
	if len(angels) != 1 {
		t.Fatalf("%d Angel Warriors, want 1", len(angels))
	}
	if p, tough := effectstest.EffectivePTOf(t, g, angels[0]); p != 4 || tough != 4 {
		t.Errorf("Angel P/T %d/%d, want 4/4", p, tough)
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, angels[0], kw) {
			t.Errorf("the Angel has no %s", kw)
		}
	}
}

func TestOldB068SurtlandFrostpyreScriesThenDamagesEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := oldB068Land(g, me.ID, "Surtland Frostpyre", oldB068Frostpyre)
	effectstest.StockLibraryFor(g, me, 6)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 3, 3)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{U}{U}{U}{U}{R}")
	if err := oldB068Activate(g, me, land, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry prompt")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, small) || effectstest.BattlefieldHas(g, mine) {
		t.Error("a creature with toughness 2 or less survived 2 damage")
	}
	if !effectstest.BattlefieldHas(g, big) {
		t.Error("the 3/3 died")
	}
}

func TestOldB068FacelessHavenBecomesAFourThreeForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	haven := effectstest.PushCatalogPermanent(g, me.ID, "Faceless Haven", "Snow Land", oldB068Haven, false)
	effectstest.AdvanceToMain(t, g)
	// Plain lands make no snow mana.
	for i := 0; i < 3; i++ {
		effectstest.B31Push(g, me.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	}
	if err := g.ActivateCatalogAbility(me.ID, haven, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{S}{S}{S} paid with nonsnow lands")
	}
	for i := 0; i < 3; i++ {
		effectstest.BlbB1Push(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "", 0, 0)
	}
	if err := g.ActivateCatalogAbility(me.ID, haven, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, haven)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("not a land creature: %v", c.Effective().Types)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, haven); p != 4 || tough != 3 {
		t.Errorf("P/T %d/%d, want 4/3", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, haven, "vigilance") {
		t.Error("no vigilance")
	}
	if !c.Effective().AllCreatureTypes {
		t.Error("not every creature type")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	c, _ = effectstest.BattlefieldCard(g, haven)
	if c.IsCreature() {
		t.Error("still a creature after the turn ended")
	}
}
