package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mrd_batch4_cards_test.go — Mirrodin, batch 4 (#5606, S82): Extraplanar
// Lens, then Gate to the Aether through Hematite Golem.

const (
	mrdExtraplanarLensOracle    = "e48f9fc7-72cf-4c3a-b3b3-d237edc01310"
	mrdGateToTheAetherOracle    = "a03ab63a-83f3-4dd6-926d-08d09ce2e696"
	mrdGlimmervoidOracle        = "b92e9854-4527-4133-8615-e282a213e7e3"
	mrdGlissaSunseekerOracle    = "bf4b9330-72df-426f-820c-77e37f7e9130"
	mrdGoblinCharbelcherOracle  = "c9b1557b-70b1-44f2-9fc2-7aa0b035b187"
	mrdGoblinDirigibleOracle    = "8e69b322-4e66-4d5b-ab46-f002c40a913f"
	mrdGoblinReplicaOracle      = "0c5a7772-8358-4549-a583-f54246869e20"
	mrdGoblinStrikerOracle      = "c991d1b8-3adb-4854-9fd3-83f06aeb3941"
	mrdGoblinWarWagonOracle     = "81a38340-9558-49d3-8de6-d7d01babfd97"
	mrdGolemSkinGauntletsOracle = "2a956224-8bdb-40d8-8405-2aa29eb228e0"
	mrdGraniteShardOracle       = "db81c783-a622-41e1-9dd6-47fab1d76967"
	mrdGridMonitorOracle        = "b947b9a7-7a42-447d-b3fd-e20dba8a8f41"
	mrdGroffskithurOracle       = "32a8ca5f-0170-4f5d-97cd-63e2b74fd653"
	mrdHeartwoodShardOracle     = "64d4bdc5-ff40-469d-bb25-82de0a49dfa4"
	mrdHematiteGolemOracle      = "f2e1bb32-de4c-4abc-ac15-6e6deaa8fe5f"
)

func TestMrdBatch4CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Extraplanar Lens": mrdExtraplanarLensOracle, "Gate to the Aether": mrdGateToTheAetherOracle,
		"Glimmervoid": mrdGlimmervoidOracle, "Glissa Sunseeker": mrdGlissaSunseekerOracle,
		"Goblin Charbelcher": mrdGoblinCharbelcherOracle, "Goblin Dirigible": mrdGoblinDirigibleOracle,
		"Goblin Replica": mrdGoblinReplicaOracle, "Goblin Striker": mrdGoblinStrikerOracle,
		"Goblin War Wagon": mrdGoblinWarWagonOracle, "Golem-Skin Gauntlets": mrdGolemSkinGauntletsOracle,
		"Granite Shard": mrdGraniteShardOracle, "Grid Monitor": mrdGridMonitorOracle,
		"Groffskithur": mrdGroffskithurOracle, "Heartwood Shard": mrdHeartwoodShardOracle,
		"Hematite Golem": mrdHematiteGolemOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered as %q, %v, caveats %v; want its name, Full and none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Extraplanar Lens --------------------------------------------------------

func TestMrdExtraplanarLensAddsManaForLandsNamedLikeTheImprint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	imprinted := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	myForest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	myIsland := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	theirForest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	effectstest.B20CastCreature(t, g, me, "Extraplanar Lens", "Artifact", mrdExtraplanarLensOracle, 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetCards, theirForest) {
		t.Error("a land an opponent controls is not a legal imprint target")
	}
	effectstest.PickCard(t, g, me.ID, imprinted)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ExileHas(g, imprinted) {
		t.Fatal("the chosen land should be exiled")
	}

	effectstest.TapForMana(t, g, me.ID, myForest)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("my Forest made %v, want two green", got)
	}
	effectstest.TapForMana(t, g, opp.ID, theirForest)
	if got := effectstest.SortedPool(opp); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("an opponent's Forest made %v, want two green: 'its controller adds'", got)
	}
	effectstest.TapForMana(t, g, me.ID, myIsland)
	if got := effectstest.SortedPool(me); len(got) != 3 {
		t.Errorf("my pool after an Island is %v, want one more mana only", got)
	}
}

func TestMrdExtraplanarLensDeclinedImprintAddsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")

	effectstest.B20CastCreature(t, g, me, "Extraplanar Lens", "Artifact", mrdExtraplanarLensOracle, 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, forest)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ExileHas(g, forest) {
		t.Fatal("declining exiles nothing")
	}
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 1 {
		t.Errorf("with nothing exiled a Forest made %v, want one green", got)
	}
}

// --- Gate to the Aether ------------------------------------------------------

func TestMrdGateToTheAetherLetsThatPlayerPutAPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gate := effectstest.Mrd2Catalog(g, me.ID, "Gate to the Aether", "Artifact", mrdGateToTheAetherOracle, "{6}", 0, 0)
	beast := effectstest.PlTop(opp, "Their Beast", "Creature — Beast", "{3}{G}")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, gate) == nil {
		t.Fatal("no trigger at an opponent's upkeep: 'each player's upkeep'")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the Gate's controller is asked; 'the player' whose upkeep it is chooses")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, beast)
	c, ok := effectstest.BattlefieldCard(g, beast)
	if !ok || c.Controller != opp.ID {
		t.Fatal("the revealed creature card should enter under that player's control")
	}
}

func TestMrdGateToTheAetherLeavesOtherCardsOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.Mrd2Catalog(g, me.ID, "Gate to the Aether", "Artifact", mrdGateToTheAetherOracle, "{6}", 0, 0)
	bolt := effectstest.PlTop(opp, "Their Bolt", "Instant", "{R}")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("an instant card is offered to the battlefield")
	}
	if top, _ := opp.Library.Top(); top.InstanceID != bolt {
		t.Error("the instant card should stay on top of the library")
	}
}

// --- Glimmervoid -------------------------------------------------------------

func TestMrdGlimmervoidIsSacrificedAtTheEndStepWithoutArtifacts(t *testing.T) {
	for _, withArtifact := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		void := effectstest.Mrd2Catalog(g, me.ID, "Glimmervoid", "Land", mrdGlimmervoidOracle, "", 0, 0)
		if withArtifact {
			effectstest.MrdPermanent(g, me.ID, "My Trinket", "Artifact", "{1}", 0, 0)
		}
		effectstest.AdvanceToEndStepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if kept := g.Battlefield.Contains(void); kept != withArtifact {
			t.Errorf("controls an artifact = %v: Glimmervoid on the battlefield = %v", withArtifact, kept)
		}
	}
}

// --- Glissa Sunseeker --------------------------------------------------------

func TestMrdGlissaSunseekerDestroysAnArtifactMatchingTheUnspentMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	glissa := effectstest.Mrd2Catalog(g, me.ID, "Glissa Sunseeker", "Legendary Creature — Elf Warrior", mrdGlissaSunseekerOracle, "{2}{G}{G}", 3, 2)
	rock := effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)

	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, glissa, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("one unspent mana: a mana value 2 artifact survives")
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("the unspent mana is counted, not spent: pool %d", me.ManaPool.Total())
	}

	effectstest.FindBattlefieldCardForTest(g, glissa).Tapped = false
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, glissa, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("two unspent mana: the mana value 2 artifact is destroyed")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, glissa, "first strike") {
		t.Error("Glissa has first strike")
	}
}

// --- Goblin Charbelcher ------------------------------------------------------

func TestMrdGoblinCharbelcherCountsNonlandsAndDoublesForAMountain(t *testing.T) {
	for _, tc := range []struct {
		land, typeLine string
		damage         int
	}{{"Forest", "Basic Land — Forest", 2}, {"Mountain", "Basic Land — Mountain", 4}} {
		t.Run(tc.land, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			belcher := effectstest.Mrd2Catalog(g, me.ID, "Goblin Charbelcher", "Artifact", mrdGoblinCharbelcherOracle, "{4}", 0, 0)
			under := effectstest.PlTop(me, "Under", "Sorcery", "{1}")
			land := effectstest.PlTop(me, tc.land, tc.typeLine, "")
			shock := effectstest.PlTop(me, "Shock", "Instant", "{R}")
			ponder := effectstest.PlTop(me, "Ponder", "Sorcery", "{U}")
			life := opp.Life

			effectstest.B06AddMana(me, "C", "C", "C")
			if err := g.ActivateCatalogAbility(me.ID, belcher, 0, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true,
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Life != life-tc.damage {
				t.Errorf("opponent's life %d, want %d", opp.Life, life-tc.damage)
			}
			order := effectstest.PutInLibraryChoiceFor(g, me.ID)
			if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 3 {
				t.Fatalf("the three revealed cards are put on the bottom in any order: %+v", order)
			}
			if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{land, ponder, shock}, nil); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if got := effectstest.LibraryBottomIDs(me, 3); !effectstest.SameIDs(got, []uuid.UUID{land, ponder, shock}) {
				t.Errorf("the bottom three are %v, in the chosen order", got)
			}
			if top, _ := me.Library.Top(); top.InstanceID != under {
				t.Error("the first unrevealed card is now on top")
			}
		})
	}
}

// --- Goblin Dirigible and Goblin War Wagon -----------------------------------

func TestMrdGoblinDirigibleAndWarWagonStayTappedUnlessTheirControllerPays(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		cost                   []string
	}{
		{"Goblin Dirigible", "Artifact Creature — Construct", mrdGoblinDirigibleOracle, []string{"C", "C", "C", "C"}},
		{"Goblin War Wagon", "Artifact Creature — Juggernaut", mrdGoblinWarWagonOracle, []string{"C", "C"}},
	} {
		for _, pay := range []bool{false, true} {
			g := effectstest.NewCatalogGame(t)
			owner := g.Seats[1]
			id := effectstest.B12Push(g, owner.ID, tc.name, tc.typeLine, tc.oracle, 4, 4)
			effectstest.ArnTap(t, g, id)
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			if !effectstest.ArnTapped(t, g, id) {
				t.Fatalf("%s untapped in its controller's untap step", tc.name)
			}
			if effectstest.TriggerOnStack(g, id) == nil {
				t.Fatalf("%s: no upkeep trigger", tc.name)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if pay {
				effectstest.B06AddMana(owner, tc.cost...)
			}
			effectstest.AnswerPayUnless(t, g, owner.ID, pay)
			if untapped := !effectstest.ArnTapped(t, g, id); untapped != pay {
				t.Errorf("%s, pay=%v: untapped = %v", tc.name, pay, untapped)
			}
		}
	}
}

// --- Goblin Replica ----------------------------------------------------------

func TestMrdGoblinReplicaSacrificesToDestroyAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	replica := effectstest.Mrd2Catalog(g, me.ID, "Goblin Replica", "Artifact Creature — Goblin", mrdGoblinReplicaOracle, "{3}", 2, 2)
	bear := effectstest.SeedBear(g, opp.ID)
	rock := effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)

	effectstest.B06AddMana(me, "C", "C", "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, replica, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("a creature was accepted as the target artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, replica, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(replica) {
		t.Error("the Replica is sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
}

// --- Goblin Striker and Grid Monitor -----------------------------------------

func TestMrdGoblinStrikerHasFirstStrikeAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	striker := effectstest.CastCatalogSpell(t, g, "Goblin Striker", "Creature — Goblin Berserker", mrdGoblinStrikerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"first strike", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, striker, kw) {
			t.Errorf("Goblin Striker has %s", kw)
		}
	}
	effectstest.DeclareAttack(t, g, opp.ID, striker)
}

func TestMrdGridMonitorStopsItsControllerCastingCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	monitor := effectstest.Mrd2Catalog(g, me.ID, "Grid Monitor", "Artifact Creature — Construct", mrdGridMonitorOracle, "{4}", 4, 6)
	if err := effectstest.CastCatalogSpellErr(t, g, "Goblin Striker", "Creature — Goblin Berserker", mrdGoblinStrikerOracle, nil); err == nil {
		t.Fatal("a creature spell was cast under Grid Monitor")
	}
	effectstest.P1Destroy(t, g, monitor)
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Goblin Striker", "Creature — Goblin Berserker", mrdGoblinStrikerOracle, nil); err != nil {
		t.Errorf("with the Monitor gone, the creature spell should be castable: %v", err)
	}
}

// --- Golem-Skin Gauntlets ----------------------------------------------------

func TestMrdGolemSkinGauntletsCountEveryEquipmentOnTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	gauntlets := effectstest.SeedEquipment(g, me.ID, "Golem-Skin Gauntlets", mrdGolemSkinGauntletsOracle)

	effectstest.EquipTo(t, g, me.ID, gauntlets, bear)
	if p, tt := effectstest.Mrd2Size(t, g, bear); p != 3 || tt != 2 {
		t.Fatalf("with the Gauntlets alone the bear is %d/%d, want 3/2", p, tt)
	}
	saw := effectstest.SeedEquipment(g, me.ID, "Bone Saw", effectstest.BoneSawOracle)
	effectstest.EquipTo(t, g, me.ID, saw, bear)
	if p, tt := effectstest.Mrd2Size(t, g, bear); p != 5 || tt != 2 {
		t.Errorf("with two Equipment the bear is %d/%d, want 5/2 (2 + 2 counted + Bone Saw's 1)", p, tt)
	}
}

// --- Granite Shard and Heartwood Shard ---------------------------------------

func TestMrdGraniteShardPingsWithEitherCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shard := effectstest.Mrd2Catalog(g, me.ID, "Granite Shard", "Artifact", mrdGraniteShardOracle, "{3}", 0, 0)
	life := opp.Life

	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, shard, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true,
	}); err != nil {
		t.Fatalf("activate the {R} row: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FindBattlefieldCardForTest(g, shard).Tapped = false
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shard, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true,
	}); err != nil {
		t.Fatalf("activate the {3} row: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent's life %d, want %d after two pings", opp.Life, life-2)
	}
}

func TestMrdHeartwoodShardGivesTrampleUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	shard := effectstest.Mrd2Catalog(g, me.ID, "Heartwood Shard", "Artifact", mrdHeartwoodShardOracle, "{3}", 0, 0)
	bear := effectstest.SeedBear(g, me.ID)

	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, shard, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate the {G} row: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the creature should have trample")
	}
	if !effectstest.TappedForTest(t, g, shard) {
		t.Error("the Shard taps to pay the cost")
	}
}

// --- Groffskithur ------------------------------------------------------------

func TestMrdGroffskithurReturnsItsNamesakeWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	groff := effectstest.Mrd2Catalog(g, me.ID, "Groffskithur", "Creature — Beast", mrdGroffskithurOracle, "{5}{G}", 3, 3)
	other := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: other, Name: "Another Beast", TypeLine: "Creature — Beast", Owner: me.ID, Controller: me.ID})
	namesake := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: namesake, Name: "Groffskithur", TypeLine: "Creature — Beast", Owner: me.ID, Controller: me.ID})
	blocker := effectstest.SeedBear(g, opp.ID)

	effectstest.Mrd3BlockedBy(t, g, opp.ID, groff, blocker)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetCards, other) {
		t.Error("a card with another name is not a legal target")
	}
	effectstest.PickCard(t, g, me.ID, namesake)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if !me.Hand.Contains(namesake) {
		t.Error("the card named Groffskithur should return to its owner's hand")
	}
}

// --- Hematite Golem ----------------------------------------------------------

func TestMrdHematiteGolemPumpsTwoPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	golem := effectstest.Mrd2Catalog(g, me.ID, "Hematite Golem", "Artifact Creature — Golem", mrdHematiteGolemOracle, "{4}", 1, 4)
	for i := 0; i < 2; i++ {
		effectstest.B06AddMana(me, "C", "R")
		if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tt := effectstest.Mrd2Size(t, g, golem); p != 5 || tt != 4 {
		t.Errorf("after two activations the Golem is %d/%d, want 5/4", p, tt)
	}
}
