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

// --- Batch 2: mana rocks, mana creatures and mana triggers ------------

const (
	c2DarksteelIngotOracle  = "a2529491-7389-4cfa-92d2-145eda779603"
	c2SeanceBoardOracle     = "8331d2ff-3729-4014-8dcb-278644d2632a"
	c2LeylineDowserOracle   = "2ea2ac09-690d-4cd2-8e3a-fc01ecd7c132"
	c2BountyBoardOracle     = "e40c66cf-0a4a-4833-98bf-c685d35b9309"
	c2CitanulOracle         = "f67df84f-667e-41cf-80f9-911abafe70ac"
	c2BrainstoneOracle      = "ffac1923-99d0-48e6-b1e7-168e5c324a80"
	c2BarbflareOracle       = "bd7accec-ca8a-49c7-9ae4-fe97d99b4a34"
	c2WolfwillowHavenOracle = "7579f57d-c76e-4703-a030-34fb7160cb23"
	c2GhirapurOrreryOracle  = "c6bb3a58-131a-4d7b-8c07-173a7de23b06"
	c2ChaliceOracle         = "74027ba4-a6ba-4696-baf1-60ef0b27ee0a"
)

func TestC2DarksteelIngotIsIndestructibleAndTapsForAnyColor(t *testing.T) {
	g, me, _ := c2LandTable(t)
	ingot := pushCatalogPermanent(g, me.ID, "Darksteel Ingot", "Artifact", c2DarksteelIngotOracle, false)
	if !hasEffectiveKeyword(t, g, ingot, "indestructible") {
		t.Error("not indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ingot) })
	if !g.Battlefield.Contains(ingot) {
		t.Error("destroyed")
	}
	if err := g.ActivateManaAbility(me.ID, ingot, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("tap for {U}: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
}

func TestC2SeanceBoardGrowsOnlyWhenACreatureDied(t *testing.T) {
	g, me, _ := c2LandTable(t)
	board := pushCatalogPermanent(g, me.ID, "Séance Board", "Artifact", c2SeanceBoardOracle, false)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := counterOnBattlefieldCard(g, board, "soul"); got != 0 {
		t.Fatalf("%d soul counters with no death this turn, want 0", got)
	}

	g2, me2, _ := c2LandTable(t)
	board2 := pushCatalogPermanent(g2, me2.ID, "Séance Board", "Artifact", c2SeanceBoardOracle, false)
	bear := pushVanillaCreature(g2, me2.ID, "Bear", 2, 2)
	killCreature(t, g2, me2.ID, bear)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if got := counterOnBattlefieldCard(g2, board2, "soul"); got != 1 {
		t.Fatalf("%d soul counters after a creature died, want 1", got)
	}
}

func TestC2SeanceBoardManaIsXOfOneColorForInstantsSorceriesDemonsAndSpirits(t *testing.T) {
	g, me, _ := c2LandTable(t)
	board := pushCatalogPermanent(g, me.ID, "Séance Board", "Artifact", c2SeanceBoardOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == board {
				g.Battlefield.Cards[i].Counters = map[string]int{"soul": 2}
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, board, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if got := poolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Fatalf("pool %v, want {R}{R}", got)
	}
	bear := handSpell(me, "Red Bear", "Creature — Bear", "{R}")
	refusedForMana(t, g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}), "a plain creature spell")
	demon := handSpell(me, "Red Demon", "Creature — Demon", "{R}")
	if err := g.CastSpell(me.ID, demon, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a Demon spell: %v", err)
	}
	bolt := handSpell(me, "Bolt", "Instant", "{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an instant: %v", err)
	}
}

func TestC2SeanceBoardWithNoCountersAddsNothing(t *testing.T) {
	g, me, _ := c2LandTable(t)
	board := pushCatalogPermanent(g, me.ID, "Séance Board", "Artifact", c2SeanceBoardOracle, false)
	_ = g.ActivateManaAbility(me.ID, board, 0, game.ManaAbilityParams{Colors: []string{"R"}})
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d with no soul counters, want 0", me.ManaPool.Total())
	}
}

func TestC2LeylineDowserMillsAndUntapsWithALegend(t *testing.T) {
	g, me, _ := c2LandTable(t)
	dowser := pushCatalogPermanent(g, me.ID, "Leyline Dowser", "Artifact", c2LeylineDowserOracle, false)
	ids := seedLibrary(me, "Spell", "Other")
	_ = ids
	c2Pay(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, dowser, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("mill: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || len(pick.ChooseCards) != 1 {
		t.Fatalf("no offer for the milled instant or sorcery: %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:1]); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Size() == 0 || me.Graveyard.Size() != 0 {
		t.Errorf("hand %d, graveyard %d: the milled spell should be in hand", me.Hand.Size(), me.Graveyard.Size())
	}

	// Untap: needs an untapped legendary creature.
	if !isTapped(g, dowser) {
		t.Fatal("the Dowser did not tap")
	}
	if err := g.ActivateCatalogAbility(me.ID, dowser, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("untapped with no legendary creature")
	}
	legend := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	plain := pushVanillaCreature(g, me.ID, "Plain", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, dowser, 1, game.ActivateAbilityParams{TapIDs: []uuid.UUID{plain}, Strict: true}); err == nil {
		t.Fatal("tapped a nonlegendary creature to untap")
	}
	if err := g.ActivateCatalogAbility(me.ID, dowser, 1, game.ActivateAbilityParams{TapIDs: []uuid.UUID{legend}, Strict: true}); err != nil {
		t.Fatalf("untap: %v", err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, dowser) {
		t.Error("the Dowser is still tapped")
	}
}

func TestC2BountyBoardPaysOpponentsWhenAMarkedCreatureDies(t *testing.T) {
	g, me, opp := c2LandTable(t)
	board := pushCatalogPermanent(g, me.ID, "Bounty Board", "Artifact", c2BountyBoardOracle, false)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	other := pushVanillaCreature(g, opp.ID, "Unmarked", 2, 2)
	c2Pay(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, board, 0, game.ActivateAbilityParams{Targets: cardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("bounty: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOnBattlefieldCard(g, victim, "bounty"); got != 1 {
		t.Fatalf("%d bounty counters, want 1", got)
	}
	killCreature(t, g, opp.ID, other)
	passPriorityAroundTable(t, g)
	hand, life := me.Hand.Size(), me.Life
	killCreature(t, g, opp.ID, victim)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+2 {
		t.Errorf("me: hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life+2)
	}
}

func TestC2CitanulHierophantsGivesEveryCreatureATapForGreen(t *testing.T) {
	g, me, opp := c2LandTable(t)
	hier := gaPut(g, me.ID, "Citanul Hierophants", "Creature — Human Druid", c2CitanulOracle)
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirs := gaPut(g, opp.ID, "Their Bear", "Creature — Bear", "")
	if n := grantedManaRowsOf(g, hier); n != 1 {
		t.Errorf("the Hierophants have %d granted mana rows, want 1: it is a creature you control", n)
	}
	if n := grantedManaRowsOf(g, bear); n != 1 {
		t.Fatalf("your bear has %d granted mana rows, want 1", n)
	}
	if n := grantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's bear has %d granted rows, want 0", n)
	}
	if err := tapGrantedMana(t, g, me.ID, bear, "G"); err != nil {
		t.Fatalf("tap the bear: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
	if err := tapGrantedMana(t, g, me.ID, bear, "G"); err == nil {
		t.Error("the bear tapped twice")
	}
	fresh := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fresh Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fresh {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
	})
	if err := tapGrantedMana(t, g, me.ID, fresh, "G"); err == nil {
		t.Error("a summoning-sick creature tapped for the granted mana (CR 302.6)")
	}
}

func TestC2BrainstoneDrawsThreeAndPutsTwoBack(t *testing.T) {
	g, me, _ := c2LandTable(t)
	stone := pushCatalogPermanent(g, me.ID, "Brainstone", "Artifact", c2BrainstoneOracle, false)
	seedLibrary(me, "A", "B", "C", "D", "E")
	hand := me.Hand.Size()
	c2Pay(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(stone) {
		t.Error("Brainstone not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Fatalf("hand %d, want %d", me.Hand.Size(), hand+3)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("put-back prompt %+v, want exactly two cards", pick)
	}
	first, second := me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{first, second}); err != nil {
		t.Fatal(err)
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt after the pick")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{second, first}); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != second {
		t.Error("the first card of the answer must end on top of the library")
	}
}

func TestC2WolfwillowHavenAddsGreenAndMakesAWolf(t *testing.T) {
	g, me, opp := c2LandTable(t)
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	haven := pushAuraOnLand(t, g, me.ID, "Wolfwillow Haven", c2WolfwillowHavenOracle, forest)
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Fatalf("pool %v, want {G}{G}", got)
	}
	me.ManaPool = game.ManaPool{}
	c2Pay(t, g, me, "{C}{C}{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, haven, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countTokensControlled(g, me.ID, "Wolf") != 1 {
		t.Error("no Wolf")
	}
	_ = opp
}

func TestC2GhirapurOrreryGivesEachPlayerAnExtraLandDrop(t *testing.T) {
	g, me, opp := c2LandTable(t)
	var before [2]int
	g.ReadSnapshot(func() { before = [2]int{g.EffectiveLandDropsLocked(me), g.EffectiveLandDropsLocked(opp)} })
	pushCatalogPermanent(g, me.ID, "Ghirapur Orrery", "Artifact", c2GhirapurOrreryOracle, false)
	g.ReadSnapshot(func() {
		if n := g.EffectiveLandDropsLocked(me); n != before[0]+1 {
			t.Errorf("controller's land drops %d, want %d", n, before[0]+1)
		}
		if n := g.EffectiveLandDropsLocked(opp); n != before[1]+1 {
			t.Errorf("opponent's land drops %d, want %d: each player", n, before[1]+1)
		}
	})
}

func TestC2GhirapurOrreryDrawsThreeForTheUpkeepPlayerOnlyWithAnEmptyHand(t *testing.T) {
	for _, empty := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, "Ghirapur Orrery", "Artifact", c2GhirapurOrreryOracle, false)
		seedLibrary(opp, "A", "B", "C", "D", "E", "F")
		seedLibrary(me, "A", "B", "C", "D", "E", "F")
		if !empty {
			handSpell(opp, "Held", "Sorcery", "{1}")
		}
		g.WithWriteLock(func() {
			if empty {
				opp.Hand.Cards = nil
			}
		})
		handBefore, myHand := opp.Hand.Size(), me.Hand.Size()
		for i := 0; i < 80 && !(g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepUpkeep); i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
		}
		if g.Turn.ActiveSeat != 1 {
			t.Fatal("never reached the opponent's upkeep")
		}
		passPriorityAroundTable(t, g)
		got := opp.Hand.Size() - handBefore
		want := 0
		if empty {
			want = 3
		}
		if got != want {
			t.Errorf("empty hand %v: the upkeep player drew %d, want %d", empty, got, want)
		}
		if me.Hand.Size() > myHand+1 {
			t.Errorf("the Orrery's controller drew %d extra", me.Hand.Size()-myHand)
		}
	}
}

func TestC2ChaliceCountersAMatchingSpellOnly(t *testing.T) {
	g, me, _ := c2LandTable(t)
	chalice := b12PlayFromHand(t, g, "Chalice of the Void", "Artifact", c2ChaliceOracle, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	if got := counterOnBattlefieldCard(g, chalice, game.CounterCharge); got != 2 {
		t.Fatalf("%d charge counters, want 2", got)
	}
	c2Pay(t, g, me, "{C}{C}{C}")
	three := handSpell(me, "Three Drop", "Artifact", "{3}")
	if err := g.CastSpell(me.ID, three, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(three) || !g.Battlefield.Contains(three) {
		t.Fatal("a mana value 3 spell was countered by a Chalice on 2")
	}
	two := handSpell(me, "Two Drop", "Artifact", "{2}")
	c2Pay(t, g, me, "{C}{C}")
	if err := g.CastSpell(me.ID, two, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(two) {
		t.Error("the mana value 2 spell was not countered")
	}
}

func TestC2BarbflareGremlinDoublesAnyPlayersLandOnlyWhileTappedAndPingsThem(t *testing.T) {
	g, me, opp := c2LandTable(t)
	gremlin := pushCatalogPermanent(g, me.ID, "Barbflare Gremlin", "Creature — Gremlin", c2BarbflareOracle, false)
	myForest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	theirMountain := b31Push(g, opp.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)

	life := me.Life
	tapForMana(t, g, me.ID, myForest)
	if me.ManaPool.Total() != 1 || me.Life != life {
		t.Fatalf("an untapped Gremlin did something: pool %d, life %d", me.ManaPool.Total(), me.Life)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == gremlin {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	untap(g, myForest)
	me.ManaPool = game.ManaPool{}
	tapForMana(t, g, me.ID, myForest)
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Fatalf("pool %v, want {G}{G}", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d: the Forest deals 1 damage to its tapper", me.Life, life-1)
	}

	oppLife, myLife := opp.Life, me.Life
	tapForMana(t, g, opp.ID, theirMountain)
	if got := sortedPool(opp); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Fatalf("opponent pool %v, want {R}{R}: a player, not you", got)
	}
	if opp.Life != oppLife-1 || me.Life != myLife {
		t.Errorf("life: opponent %d (want %d), me %d (want %d)", opp.Life, oppLife-1, me.Life, myLife)
	}
}
