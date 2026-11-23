package testss82v

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSlice294cCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.Slice294cAccessTunnelOracle:       "Access Tunnel",
		effectstest.Slice294cCanyonSloughOracle:       "Canyon Slough",
		effectstest.Slice294cShelteredThicketOracle:   "Sheltered Thicket",
		effectstest.Slice294cCommercialDistrictOracle: "Commercial District",
		effectstest.Slice294cHushwoodVergeOracle:      "Hushwood Verge",
		effectstest.Slice294cMortuaryMireOracle:       "Mortuary Mire",
		effectstest.Slice294cOpalPalaceOracle:         "Opal Palace",
		effectstest.Slice294cViridescentBogOracle:     "Viridescent Bog",
		effectstest.Slice294cWarRoomOracle:            "War Room",
		effectstest.Slice294cShiftingWoodlandOracle:   "Shifting Woodland",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- Access Tunnel ---------------------------------------------------

func TestAccessTunnelMakesASmallCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tunnel := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Access Tunnel",
		TypeLine:   "Land",
		OracleID:   effectstest.Slice294cAccessTunnelOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	small := effectstest.PushRestrictionBear(g, me, "Small Beast") // 2/2
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, tunnel, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertRestrictions(t, g, small, game.CantBeBlocked)
}

// A creature with power over 3 is not a legal target at all.
func TestAccessTunnelRefusesACreatureOverPowerThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tunnel := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Access Tunnel",
		TypeLine:   "Land",
		OracleID:   effectstest.Slice294cAccessTunnelOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Big Beast",
		TypeLine:   "Creature — Beast",
		Power:      4,
		Toughness:  4,
		Owner:      me.ID,
		Controller: me.ID,
	})
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	err := g.ActivateCatalogAbility(me.ID, tunnel, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a power-4 creature: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Canyon Slough & Sheltered Thicket --------------------------------

func TestCanyonSloughAndShelteredThicketEnterTappedCycleAndTapForColours(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, a, b string
	}{
		{"Canyon Slough", effectstest.Slice294cCanyonSloughOracle, "B", "R"},
		{"Sheltered Thicket", effectstest.Slice294cShelteredThicketOracle, "R", "G"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]

			id := effectstest.PlayLandFromHand(t, g, tc.name, tc.oracle)
			card, ok := effectstest.BattlefieldCard(g, id)
			if !ok || !card.Tapped {
				t.Fatalf("%s must enter tapped", tc.name)
			}
			if n := effectstest.TapEventsFor(g, id); n != 0 {
				t.Errorf("%d tap events; the land should have ENTERED tapped", n)
			}

			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
			effectstest.B28TapForMana(t, g, me.ID, id, tc.a)
			if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{tc.a}) {
				t.Errorf("pool = %v, want {%s}", got, tc.a)
			}

			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
			me.ManaPool = nil
			effectstest.B28TapForMana(t, g, me.ID, id, tc.b)
			if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{tc.b}) {
				t.Errorf("pool = %v, want {%s}", got, tc.b)
			}
		})
	}
}

func TestCanyonSloughAndShelteredThicketCycle(t *testing.T) {
	for _, tc := range []struct{ name, oracle, typeLine string }{
		{"Canyon Slough", effectstest.Slice294cCanyonSloughOracle, "Land — Swamp Mountain"},
		{"Sheltered Thicket", effectstest.Slice294cShelteredThicketOracle, "Land — Mountain Forest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			handBefore := me.Hand.Size()
			_, _ = effectstest.CycleFromHand(t, g, tc.name, tc.typeLine, tc.oracle, "{C}{C}")
			if got := me.Hand.Size(); got != handBefore-1+1 {
				// The card leaves and a new one is drawn: net zero.
				t.Errorf("hand size = %d, want %d after cycling", got, handBefore)
			}
			if me.Graveyard.Size() == 0 {
				t.Error("cycled land did not land in the graveyard")
			}
		})
	}
}

// --- Commercial District ----------------------------------------------

func TestCommercialDistrictEntersTappedAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Next Draw")

	id := effectstest.PlayLandFromHand(t, g, "Commercial District", effectstest.Slice294cCommercialDistrictOracle)
	effectstest.PassPriorityAroundTable(t, g)

	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("Commercial District must enter tapped")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("playing Commercial District did not queue a surveil")
	}
}

// --- Mortuary Mire -----------------------------------------------------

func TestMortuaryMireEntersTappedAndOffersToRebuyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zombie := effectstest.Batch01GraveyardCard(me, "Rotting Zombie", "Creature")

	id := effectstest.PlayLandFromHand(t, g, "Mortuary Mire", effectstest.Slice294cMortuaryMireOracle)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("Mortuary Mire must enter tapped")
	}

	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no pick_target prompt after the \"you may\"")
	}
	effectstest.PickCard(t, g, me.ID, zombie)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Graveyard.Contains(zombie) {
		t.Error("the creature is still in the graveyard")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != zombie {
		t.Error("the creature did not go on top of the library")
	}
}

// Declining the "you may" leaves the graveyard alone.
func TestMortuaryMireDeclinedLeavesTheGraveyardAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zombie := effectstest.Batch01GraveyardCard(me, "Rotting Zombie", "Creature")

	effectstest.PlayLandFromHand(t, g, "Mortuary Mire", effectstest.Slice294cMortuaryMireOracle)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(zombie) {
		t.Error("declining the trigger still moved the creature")
	}
}

// --- Hushwood Verge ------------------------------------------------------

func TestHushwoodVergeSecondColourNeedsAForestOrAPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	verge := effectstest.SeedPermanentWithOracle(g, me.ID, "Hushwood Verge", "Land", effectstest.Slice294cHushwoodVergeOracle)
	// An Island is a land, but not one the clause names.
	effectstest.SeedManaLand(g, me.ID, "Island", "Basic Land — Island", "U")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("gated {W} with no Forest or Plains: err = %v, want ErrConditionNotMet", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, verge); c.Tapped {
		t.Fatal("a refused activation tapped the verge")
	}
	if err := g.ActivateManaAbility(me.ID, verge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("unconditional {G}: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}
}

func TestHushwoodVergeGatedColourWithAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	verge := effectstest.SeedPermanentWithOracle(g, me.ID, "Hushwood Verge", "Land", effectstest.Slice294cHushwoodVergeOracle)
	effectstest.SeedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("gated ability: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}
}

// --- Opal Palace ---------------------------------------------------------

func TestOpalPalaceTapsForColourlessAndFilteredCommanderIdentity(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RiderGiveImportedCommander(g, me, "Test Commander", "{W}{U}", []string{"W", "U"})
	palace := effectstest.SeedPermanentWithOracle(g, me.ID, "Opal Palace", "Land", effectstest.Slice294cOpalPalaceOracle)

	if err := g.ActivateManaAbility(me.ID, palace, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("colourless ability: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want {C}", got)
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(palace) })
	me.ManaPool = nil
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, palace, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility (filter ability): %v", err)
	}
	if n := effectstest.B10ResolveAllManaPicks(t, g, me.ID, "W"); n != 1 {
		t.Fatalf("resolved %d mana picks, want 1", n)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}
}

// --- Viridescent Bog -------------------------------------------------

func TestViridescentBogAddsBothColoursForOneGenericAndATap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bog := effectstest.SeedPermanentWithOracle(g, me.ID, "Viridescent Bog", "Land", effectstest.Slice294cViridescentBogOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateManaAbility(me.ID, bog, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := effectstest.PoolColors(me)
	if len(got) != 2 || got[0] != "B" || got[1] != "G" {
		t.Errorf("pool = %v, want [B G]", got)
	}
}

// Without the {1} in the pool, the land stays untapped.
func TestViridescentBogRefusesWithoutTheGenericMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bog := effectstest.SeedPermanentWithOracle(g, me.ID, "Viridescent Bog", "Land", effectstest.Slice294cViridescentBogOracle)

	err := g.ActivateManaAbility(me.ID, bog, 0, game.ManaAbilityParams{})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("err = %v, want ErrInsufficientMana", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, bog); c.Tapped {
		t.Error("a refused activation tapped the land")
	}
}

// --- War Room ------------------------------------------------------------

// The draw ability landed with #1594 (computed_life_cost_test.go); this
// test keeps the mana half and pins that the draw is costed by the
// computed count, never by a flat number or not at all.
func TestWarRoomTapsForColourlessAndCostsItsDrawByTheCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	room := effectstest.SeedPermanentWithOracle(g, me.ID, "War Room", "Land", effectstest.Slice294cWarRoomOracle)

	if err := g.ActivateManaAbility(me.ID, room, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want {C}", got)
	}

	spec, ok := effects.Lookup(effectstest.Slice294cWarRoomOracle)
	if !ok {
		t.Fatal("War Room is not registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Completeness = %v, want CompletenessFull", spec.Completeness)
	}
	if len(spec.Activated) != 1 {
		t.Fatalf("War Room declares %d activated abilities, want the draw", len(spec.Activated))
	}
	if c := spec.Activated[0].Cost; c.LifeFrom != effects.LifeEqualToCommanderColors || c.Life != 0 {
		t.Errorf("the draw's life is %d + %q, want the commander-colour count alone", c.Life, c.LifeFrom.Key())
	}
}

// --- Shifting Woodland -----------------------------------------------

func TestShiftingWoodlandEntersTappedUnlessAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)

	id := effectstest.PlayLandFromHand(t, g, "Shifting Woodland", effectstest.Slice294cShiftingWoodlandOracle)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("Shifting Woodland must enter tapped with no Forest in play")
	}
}

func TestShiftingWoodlandEntersUntappedWithAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")

	id := effectstest.PlayLandFromHand(t, g, "Shifting Woodland", effectstest.Slice294cShiftingWoodlandOracle)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || card.Tapped {
		t.Fatal("Shifting Woodland must enter untapped beside a Forest")
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
	effectstest.B28TapForMana(t, g, me.ID, id, "G")
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}

	spec, ok := effects.Lookup(effectstest.Slice294cShiftingWoodlandOracle)
	if !ok {
		t.Fatal("Shifting Woodland is not registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Completeness = %v, want CompletenessFull (#1593 built the Delirium copy ability)", spec.Completeness)
	}
}
