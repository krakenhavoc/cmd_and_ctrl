package effects

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice294c_test.go — slice 294-c (Lands): Access Tunnel, Canyon
// Slough, Sheltered Thicket, Commercial District, Hushwood Verge,
// Mortuary Mire, Opal Palace, Viridescent Bog, War Room, Shifting
// Woodland.

const (
	slice294cAccessTunnelOracle       = "ed9cc560-f30b-4b60-a094-ccf93ed656a7"
	slice294cCanyonSloughOracle       = "2031b17c-0536-446f-a9aa-b46fe79b7ea7"
	slice294cShelteredThicketOracle   = "db8d8643-3d0b-4f20-bf53-f4cd26a0e8df"
	slice294cCommercialDistrictOracle = "b33656ae-3473-4223-845f-f9147f87678b"
	slice294cHushwoodVergeOracle      = "cce328b9-6100-417e-9ddf-808bbe3e3bc5"
	slice294cMortuaryMireOracle       = "1b3fb20a-e090-4286-9c03-6b71c27c45be"
	slice294cOpalPalaceOracle         = "aa6723a2-75da-49f5-a1ba-cbfa82c55301"
	slice294cViridescentBogOracle     = "6bd6d259-1af7-4dff-a79c-48a616d2a36e"
	slice294cWarRoomOracle            = "71c52bf5-2a5d-488e-8b15-7ef290e4b77d"
	slice294cShiftingWoodlandOracle   = "7c2a4fe5-43e8-4e20-bef2-0278d18afc4b"
)

func TestSlice294cCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		slice294cAccessTunnelOracle:       "Access Tunnel",
		slice294cCanyonSloughOracle:       "Canyon Slough",
		slice294cShelteredThicketOracle:   "Sheltered Thicket",
		slice294cCommercialDistrictOracle: "Commercial District",
		slice294cHushwoodVergeOracle:      "Hushwood Verge",
		slice294cMortuaryMireOracle:       "Mortuary Mire",
		slice294cOpalPalaceOracle:         "Opal Palace",
		slice294cViridescentBogOracle:     "Viridescent Bog",
		slice294cWarRoomOracle:            "War Room",
		slice294cShiftingWoodlandOracle:   "Shifting Woodland",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tunnel := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Access Tunnel",
		TypeLine:   "Land",
		OracleID:   slice294cAccessTunnelOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	small := pushRestrictionBear(g, me, "Small Beast") // 2/2
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, tunnel, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	assertRestrictions(t, g, small, game.CantBeBlocked)
}

// A creature with power over 3 is not a legal target at all.
func TestAccessTunnelRefusesACreatureOverPowerThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tunnel := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Access Tunnel",
		TypeLine:   "Land",
		OracleID:   slice294cAccessTunnelOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	big := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Big Beast",
		TypeLine:   "Creature — Beast",
		Power:      4,
		Toughness:  4,
		Owner:      me.ID,
		Controller: me.ID,
	})
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
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
		{"Canyon Slough", slice294cCanyonSloughOracle, "B", "R"},
		{"Sheltered Thicket", slice294cShelteredThicketOracle, "R", "G"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]

			id := playLandFromHand(t, g, tc.name, tc.oracle)
			card, ok := battlefieldCard(g, id)
			if !ok || !card.Tapped {
				t.Fatalf("%s must enter tapped", tc.name)
			}
			if n := tapEventsFor(g, id); n != 0 {
				t.Errorf("%d tap events; the land should have ENTERED tapped", n)
			}

			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
			b28TapForMana(t, g, me.ID, id, tc.a)
			if got := poolColors(me); !reflect.DeepEqual(got, []string{tc.a}) {
				t.Errorf("pool = %v, want {%s}", got, tc.a)
			}

			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
			me.ManaPool = nil
			b28TapForMana(t, g, me.ID, id, tc.b)
			if got := poolColors(me); !reflect.DeepEqual(got, []string{tc.b}) {
				t.Errorf("pool = %v, want {%s}", got, tc.b)
			}
		})
	}
}

func TestCanyonSloughAndShelteredThicketCycle(t *testing.T) {
	for _, tc := range []struct{ name, oracle, typeLine string }{
		{"Canyon Slough", slice294cCanyonSloughOracle, "Land — Swamp Mountain"},
		{"Sheltered Thicket", slice294cShelteredThicketOracle, "Land — Mountain Forest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			handBefore := me.Hand.Size()
			_, _ = cycleFromHand(t, g, tc.name, tc.typeLine, tc.oracle, "{C}{C}")
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Next Draw")

	id := playLandFromHand(t, g, "Commercial District", slice294cCommercialDistrictOracle)
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("Commercial District must enter tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
	if surveilChoiceFor(g, me.ID) == nil {
		t.Fatal("playing Commercial District did not queue a surveil")
	}
}

// --- Mortuary Mire -----------------------------------------------------

func TestMortuaryMireEntersTappedAndOffersToRebuyACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zombie := batch01GraveyardCard(me, "Rotting Zombie", "Creature")

	id := playLandFromHand(t, g, "Mortuary Mire", slice294cMortuaryMireOracle)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("Mortuary Mire must enter tapped")
	}

	answerLatestTriggerPrompt(t, g, me.ID, true)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no pick_target prompt after the \"you may\"")
	}
	pickCard(t, g, me.ID, zombie)
	passPriorityAroundTable(t, g)

	if me.Graveyard.Contains(zombie) {
		t.Error("the creature is still in the graveyard")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != zombie {
		t.Error("the creature did not go on top of the library")
	}
}

// Declining the "you may" leaves the graveyard alone.
func TestMortuaryMireDeclinedLeavesTheGraveyardAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zombie := batch01GraveyardCard(me, "Rotting Zombie", "Creature")

	playLandFromHand(t, g, "Mortuary Mire", slice294cMortuaryMireOracle)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(zombie) {
		t.Error("declining the trigger still moved the creature")
	}
}

// --- Hushwood Verge ------------------------------------------------------

func TestHushwoodVergeSecondColourNeedsAForestOrAPlains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	verge := seedPermanentWithOracle(g, me.ID, "Hushwood Verge", "Land", slice294cHushwoodVergeOracle)
	// An Island is a land, but not one the clause names.
	seedManaLand(g, me.ID, "Island", "Basic Land — Island", "U")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("gated {W} with no Forest or Plains: err = %v, want ErrConditionNotMet", err)
	}
	if c, _ := battlefieldCard(g, verge); c.Tapped {
		t.Fatal("a refused activation tapped the verge")
	}
	if err := g.ActivateManaAbility(me.ID, verge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("unconditional {G}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}
}

func TestHushwoodVergeGatedColourWithAForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	verge := seedPermanentWithOracle(g, me.ID, "Hushwood Verge", "Land", slice294cHushwoodVergeOracle)
	seedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("gated ability: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}
}

// --- Opal Palace ---------------------------------------------------------

func TestOpalPalaceTapsForColourlessAndFilteredCommanderIdentity(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	riderGiveImportedCommander(g, me, "Test Commander", "{W}{U}", []string{"W", "U"})
	palace := seedPermanentWithOracle(g, me.ID, "Opal Palace", "Land", slice294cOpalPalaceOracle)

	if err := g.ActivateManaAbility(me.ID, palace, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("colourless ability: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
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
	if n := b10ResolveAllManaPicks(t, g, me.ID, "W"); n != 1 {
		t.Fatalf("resolved %d mana picks, want 1", n)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}
}

// --- Viridescent Bog -------------------------------------------------

func TestViridescentBogAddsBothColoursForOneGenericAndATap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bog := seedPermanentWithOracle(g, me.ID, "Viridescent Bog", "Land", slice294cViridescentBogOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateManaAbility(me.ID, bog, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := poolColors(me)
	if len(got) != 2 || got[0] != "B" || got[1] != "G" {
		t.Errorf("pool = %v, want [B G]", got)
	}
}

// Without the {1} in the pool, the land stays untapped.
func TestViridescentBogRefusesWithoutTheGenericMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bog := seedPermanentWithOracle(g, me.ID, "Viridescent Bog", "Land", slice294cViridescentBogOracle)

	err := g.ActivateManaAbility(me.ID, bog, 0, game.ManaAbilityParams{})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("err = %v, want ErrInsufficientMana", err)
	}
	if c, _ := battlefieldCard(g, bog); c.Tapped {
		t.Error("a refused activation tapped the land")
	}
}

// --- War Room ------------------------------------------------------------

func TestWarRoomTapsForColourlessAndDeclaresOnlyThat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	room := seedPermanentWithOracle(g, me.ID, "War Room", "Land", slice294cWarRoomOracle)

	if err := g.ActivateManaAbility(me.ID, room, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want {C}", got)
	}

	spec, ok := Lookup(slice294cWarRoomOracle)
	if !ok {
		t.Fatal("War Room is not registered")
	}
	if spec.Completeness != CompletenessCaveats {
		t.Errorf("Completeness = %v, want CompletenessCaveats (the draw ability is not implemented)", spec.Completeness)
	}
	if len(spec.Activated) != 0 {
		t.Error("War Room declares an activated ability it cannot cost correctly; it should declare none")
	}
}

// --- Shifting Woodland -----------------------------------------------

func TestShiftingWoodlandEntersTappedUnlessAForest(t *testing.T) {
	g := newCatalogGame(t)

	id := playLandFromHand(t, g, "Shifting Woodland", slice294cShiftingWoodlandOracle)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("Shifting Woodland must enter tapped with no Forest in play")
	}
}

func TestShiftingWoodlandEntersUntappedWithAForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")

	id := playLandFromHand(t, g, "Shifting Woodland", slice294cShiftingWoodlandOracle)
	card, ok := battlefieldCard(g, id)
	if !ok || card.Tapped {
		t.Fatal("Shifting Woodland must enter untapped beside a Forest")
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
	b28TapForMana(t, g, me.ID, id, "G")
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}

	spec, ok := Lookup(slice294cShiftingWoodlandOracle)
	if !ok {
		t.Fatal("Shifting Woodland is not registered")
	}
	if spec.Completeness != CompletenessCaveats {
		t.Errorf("Completeness = %v, want CompletenessCaveats (the Delirium copy ability is not implemented)", spec.Completeness)
	}
}
