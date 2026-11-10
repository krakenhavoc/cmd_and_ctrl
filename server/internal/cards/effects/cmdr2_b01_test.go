package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander deck wave 2, batch 1 (simple lands) and batch 2 (#5368):
// the cards' tests.

const (
	c2AncientZigguratOracle     = "0baabe39-72ae-47bd-a095-cbf7eb8a6361"
	c2SunhomeOracle             = "a9f8344c-1705-4254-81d6-aa05e0c69c29"
	c2AbundantCountrysideOracle = "e3eb6f90-ccfc-41e7-bff6-0b378226bc7e"
	c2LabyrinthOracle           = "9ec5a487-d8ed-459a-8f58-56f6e9a2dfe8"
	c2DarkmossBridgeOracle      = "2065cada-4078-41c4-9e06-2460d2a2e8ee"
	c2VituGhaziOracle           = "88fb9e82-28b9-4275-a1e2-cb3a9bfda127"
	c2HalimarDepthsOracle       = "42d121a2-5266-483a-ab16-e0a8073cd6a3"
	c2TocasiasDigSiteOracle     = "a91f93fd-e428-4a36-b1b3-604b47a34287"
	c2SwarmyardOracle           = "4b508087-99da-4eb1-8b12-29162f2ec85d"
	c2DesertIndomitableOracle   = "852d4dc3-404d-4565-99e9-1eac8f6eca5e"
	c2DesertFerventOracle       = "311f38a6-f68f-4d30-bc4e-62339f1e0d88"
	c2PaintedBluffsOracle       = "b66deeb5-7371-4f06-b10e-d65165bc07b2"
)

func c2LandTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func c2Land(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushCatalogPermanent(g, owner, name, typeLine, oracle, false)
}

func c2Pay(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
}

func TestC2AncientZigguratManaPaysOnlyForCreatureSpells(t *testing.T) {
	g, me, _ := c2LandTable(t)
	zig := c2Land(g, me.ID, "Ancient Ziggurat", "Land", c2AncientZigguratOracle)
	if err := g.ActivateManaAbility(me.ID, zig, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap for {G}: %v", err)
	}
	cost, _ := game.ParseCost("{G}")
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(mustBattlefield(t, g, zig))) {
		t.Error("the Ziggurat's mana pays an ability: stronger than printed")
	}
	sorcery := handSpell(me, "Green Sorcery", "Sorcery", "{G}")
	refusedForMana(t, g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}), "a noncreature spell")
	bear := handSpell(me, "Green Bear", "Creature — Bear", "{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a creature spell with the Ziggurat's mana: %v", err)
	}
}

func mustBattlefield(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c
}

func TestC2SunhomeGivesDoubleStrike(t *testing.T) {
	g, me, _ := c2LandTable(t)
	land := c2Land(g, me.ID, "Sunhome, Fortress of the Legion", "Land", c2SunhomeOracle)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("activated without {2}{R}{W}")
	}
	c2Pay(t, g, me, "{C}{C}{R}{W}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !isTapped(g, land) {
		t.Error("Sunhome did not tap as a cost")
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the target did not gain double strike")
	}
}

func TestC2AbundantCountrysideMakesAChangelingShapeshifter(t *testing.T) {
	g, me, _ := c2LandTable(t)
	land := c2Land(g, me.ID, "Abundant Countryside", "Land", c2AbundantCountrysideOracle)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {6}")
	}
	c2Pay(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Shapeshifter"); got != 1 {
		t.Fatalf("%d Shapeshifters, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Shapeshifter" {
			if c.Power != 1 || c.Toughness != 1 || !c.IsColorless() || !hasEffectiveKeyword(t, g, c.InstanceID, "changeling") {
				t.Errorf("Shapeshifter %d/%d, want a colorless 1/1 with changeling", c.Power, c.Toughness)
			}
		}
	}
}

func TestC2AbundantCountrysideHasBothManaAbilities(t *testing.T) {
	g, me, _ := c2LandTable(t)
	land := c2Land(g, me.ID, "Abundant Countryside", "Land", c2AbundantCountrysideOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{C}: %v", err)
	}
	cost, _ := game.ParseCost("{1}")
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(mustBattlefield(t, g, land))) {
		t.Error("the colourless mana is restricted")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	me.ManaPool = game.ManaPool{}
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("any colour: %v", err)
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(mustBattlefield(t, g, land))) {
		t.Error("the any-colour mana pays an ability")
	}
}

func TestC2LabyrinthRemovesAnAttackerOrBlockerFromCombat(t *testing.T) {
	g, me, opp := c2LandTable(t)
	land := c2Land(g, opp.ID, "Labyrinth of Skophos", "Land", c2LabyrinthOracle)
	idle := pushVanillaCreature(g, opp.ID, "Idle", 2, 2)
	att := pushVanillaCreature(g, me.ID, "Attacker", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(att, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	c2Pay(t, g, opp, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(opp.ID, land, 0, game.ActivateAbilityParams{Targets: cardRefs(idle), Strict: true}); err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	if err := g.ActivateCatalogAbility(opp.ID, land, 0, game.ActivateAbilityParams{Targets: cardRefs(att), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := mustBattlefield(t, g, att); c.AttackingTarget != uuid.Nil {
		t.Error("the attacker is still attacking")
	}
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	if opp.Life != life {
		t.Errorf("life %d, want %d: the removed attacker dealt damage", opp.Life, life)
	}
}

func TestC2DarkmossBridgeEntersTappedAndIsIndestructible(t *testing.T) {
	g, _, _ := c2LandTable(t)
	id := playLandFromHand(t, g, "Darkmoss Bridge", c2DarkmossBridgeOracle)
	if !isTapped(g, id) {
		t.Error("Darkmoss Bridge entered untapped")
	}
	if !hasEffectiveKeyword(t, g, id, "indestructible") {
		t.Error("Darkmoss Bridge is not indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	if !g.Battlefield.Contains(id) {
		t.Error("Darkmoss Bridge was destroyed")
	}
}

func TestC2VituGhaziMakesASaproling(t *testing.T) {
	g, me, _ := c2LandTable(t)
	land := c2Land(g, me.ID, "Vitu-Ghazi, the City-Tree", "Land", c2VituGhaziOracle)
	c2Pay(t, g, me, "{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {W}")
	}
	c2Pay(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Saproling"); got != 1 {
		t.Fatalf("%d Saprolings, want 1", got)
	}
}

func TestC2HalimarDepthsEntersTappedAndOrdersTheTopThree(t *testing.T) {
	g, me, _ := c2LandTable(t)
	lib := seedLibrary(me, "A", "B", "C", "D")
	id := playLandFromHand(t, g, "Halimar Depths", c2HalimarDepthsOracle)
	if !isTapped(g, id) {
		t.Error("Halimar Depths entered untapped")
	}
	passPriorityAroundTable(t, g)
	ask := putInLibraryChoiceFor(g, me.ID)
	if ask == nil || len(ask.ScryCards) != 3 || ask.LibraryPlacement != game.LibraryPlaceTop {
		t.Fatalf("order prompt %+v, want the top three cards to put back on top", ask)
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, []uuid.UUID{lib[2], lib[0], lib[1]}); err != nil {
		t.Fatal(err)
	}
	if got := libraryTopNames(me, 1); len(got) != 1 || got[0] != "C" {
		t.Errorf("top card %v, want C: the card chosen first", got)
	}
}

func TestC2TocasiasDigSiteSurveils(t *testing.T) {
	g, me, _ := c2LandTable(t)
	land := c2Land(g, me.ID, "Tocasia's Dig Site", "Land", c2TocasiasDigSiteOracle)
	seedLibrary(me, "A", "B")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {3}")
	}
	c2Pay(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil && c.Chooser == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("no surveil prompt")
	}
}

func TestC2SwarmyardRegeneratesOnlyItsCreatureTypes(t *testing.T) {
	g, me, _ := c2LandTable(t)
	land := c2Land(g, me.ID, "Swarmyard", "Land", c2SwarmyardOracle)
	bear := enterCard(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("targeted a Bear")
	}
	for i, typ := range []string{"Insect", "Rat", "Spider", "Squirrel"} {
		l := c2Land(g, me.ID, "Swarmyard", "Land", c2SwarmyardOracle)
		c := enterCard(t, g, me.ID, game.Card{Name: typ, TypeLine: "Creature — " + typ, Power: 1, Toughness: 1})
		if err := g.ActivateCatalogAbility(me.ID, l, 0, game.ActivateAbilityParams{Targets: cardRefs(c), Strict: true}); err != nil {
			t.Fatalf("%s: %v", typ, err)
		}
		passPriorityAroundTable(t, g)
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(c) })
		if !g.Battlefield.Contains(c) {
			t.Errorf("%d: the regenerated %s was destroyed", i, typ)
		}
	}
}

func TestC2DesertsEnterTappedAndCycle(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Desert of the Indomitable", c2DesertIndomitableOracle, "{C}{G}"},
		{"Desert of the Fervent", c2DesertFerventOracle, "{C}{R}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := c2LandTable(t)
			seedLibrary(me, "A", "B")
			id := playLandFromHand(t, g, tc.name, tc.oracle)
			if !isTapped(g, id) {
				t.Error("entered untapped")
			}
			cid, _ := cycleFromHand(t, g, tc.name, "Land — Desert", tc.oracle, tc.mana)
			passPriorityAroundTable(t, g)
			if me.Hand.Contains(cid) {
				t.Error("cycled card still in hand")
			}
			if !me.Graveyard.Contains(cid) {
				t.Error("cycled card is not in the graveyard")
			}
		})
	}
}

func TestC2PaintedBluffsFiltersAnyColor(t *testing.T) {
	g, me, _ := c2LandTable(t)
	land := c2Land(g, me.ID, "Painted Bluffs", "Land — Desert", c2PaintedBluffsOracle)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("the filter ability paid its {1} from nothing")
	}
	c2Pay(t, g, me, "{C}")
	before := me.ManaPool.Total()
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("filter: %v", err)
	}
	if me.ManaPool.Total() != before {
		t.Errorf("pool %d, want %d: {1} in, one mana out", me.ManaPool.Total(), before)
	}
	cost, _ := game.ParseCost("{R}")
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(mustBattlefield(t, g, land))) {
		t.Error("the filtered mana is not red")
	}
}
