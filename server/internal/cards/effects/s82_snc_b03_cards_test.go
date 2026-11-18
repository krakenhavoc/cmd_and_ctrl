package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b03_cards_test.go — Streets of New Capenna, batch 3 (#5491). Each
// test plays the real catalog entry and asserts what its text says.

const (
	sncJackhammerOracle  = "f73d89ec-05c9-4998-a24b-bc675b04ec16"
	sncPinionsOracle     = "8a7c3718-74a1-4f26-8e06-0e17db4b0077"
	sncDaggerOracle      = "47b4dd69-0b2a-4dc3-99e8-f572fe049327"
	sncParcelOracle      = "be3c138b-045b-4b08-a12f-a5679d972051"
	sncScarabOracle      = "1201c146-54af-4668-a544-cde3ece23f2e"
	sncGoldhoundOracle   = "b5eb72bb-897f-4911-a797-3153c31df3e0"
	sncCementShoesOracle = "52d04c36-9d38-42b2-9e47-905ab06cb708"
)

var sncCyclingLands = []struct{ oracle, name, a, b string }{
	{"7e2eb4d5-22a3-43c1-8cf3-e85723da1b61", "Racers' Ring", "R", "G"},
	{"e90e519d-023e-4d19-85ad-9972a76df3ba", "Tramway Station", "B", "R"},
	{"19e6ebaf-7148-428f-8c79-1e9e57a4e659", "Botanical Plaza", "G", "W"},
	{"18b0b583-d17b-4c67-bc67-c4d9aad155c7", "Waterfront District", "U", "B"},
	{"46466819-3c8b-4f7e-82c2-93cb0118a6e6", "Skybridge Towers", "W", "U"},
}

func TestSncB03EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		sncJackhammerOracle:  "Jackhammer",
		sncPinionsOracle:     "Gilded Pinions",
		sncDaggerOracle:      "Quick-Draw Dagger",
		sncParcelOracle:      "Ominous Parcel",
		sncScarabOracle:      "Halo Scarab",
		sncGoldhoundOracle:   "Goldhound",
		sncCementShoesOracle: "Cement Shoes",
	}
	for _, l := range sncCyclingLands {
		want[l.oracle] = l.name
	}
	if len(want) != 12 {
		t.Fatalf("the batch ships 12 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestSncB03CyclingLandsEnterTappedAndTapForTheirTwoColours(t *testing.T) {
	for _, l := range sncCyclingLands {
		spec, _ := Lookup(l.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{"+l.a+"|"+l.b+"}" {
			t.Errorf("%s: mana abilities %+v, want one producing {%s} or {%s}", l.name, spec.ManaAbilities, l.a, l.b)
		}
		g := newCatalogGame(t)
		id := playLandFromHand(t, g, l.name, l.oracle)
		top100AssertEnteredTapped(t, g, id, l.name)
	}
}

func TestSncB03CyclingLandSacrificesForFourToDrawACard(t *testing.T) {
	for _, l := range sncCyclingLands {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		land := pushCatalogPermanent(g, me.ID, l.name, "Land", l.oracle, false)
		avB20LibraryTop(g, me, game.Card{Name: "Library Card", TypeLine: "Sorcery"})
		// Colourless alone can't pay the two coloured pips.
		b09AddColorless(me, 4)
		if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Fatalf("%s: activated with no coloured mana", l.name)
		}
		me.ManaPool.AddMana(game.ManaToken{Color: l.a})
		me.ManaPool.AddMana(game.ManaToken{Color: l.b})
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("%s: activate: %v", l.name, err)
		}
		if !inGraveyard(me, land) {
			t.Fatalf("%s: the land is sacrificed as a cost", l.name)
		}
		passPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+1 {
			t.Errorf("%s: hand %d → %d, want one card drawn", l.name, hand, me.Hand.Size())
		}
	}
}

func TestSncB03JackhammerGivesPlusTwoPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	jack := pushCatalogPermanent(g, me.ID, "Jackhammer", "Artifact — Equipment", sncJackhammerOracle, false)
	equipTo(t, g, me.ID, jack, bear)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 4/2", p, tough)
	}
}

func TestSncB03GildedPinionsMakesATreasureAndGivesFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Gilded Pinions", "Artifact — Equipment", sncPinionsOracle, nil)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Fatal("no Treasure token on entry")
	}
	if hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Fatal("an unattached Pinions gave flying")
	}
	pinions := findBattlefieldByName(g, "Gilded Pinions")
	equipTo(t, g, me.ID, pinions, bear)
	if !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the equipped Bear lacks flying")
	}
}

func TestSncB03QuickDrawDaggerAttachesOnEntryAndGivesFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dagger := castCatalogSpell(t, g, "Quick-Draw Dagger", "Artifact — Equipment", sncDaggerOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if c := e2Card(t, g, dagger); c.AttachedTo.ID != bear {
		t.Fatalf("the Dagger is attached to %v, want the Bear", c.AttachedTo.ID)
	}
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the Bear did not gain first strike")
	}
	spec, _ := Lookup(sncDaggerOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestSncB03OminousParcelTutorsABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	parcel := pushCatalogPermanent(g, me.ID, "Ominous Parcel", "Artifact", sncParcelOracle, false)
	stapleLibraryCard(me, "Urborg, Tomb of Yawgmoth", "Legendary Land")
	island := stapleLibraryCard(me, "Island", "Basic Land — Island")
	stapleLibraryCard(me, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateCatalogAbility(me.ID, parcel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, island)
	if !me.Hand.Contains(island) {
		t.Error("the basic land did not reach the hand")
	}
	if g.Battlefield.Contains(parcel) {
		t.Error("the Parcel is sacrificed as a cost")
	}
}

func TestSncB03OminousParcelDealsFourDamageToACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	parcel := pushCatalogPermanent(g, me.ID, "Ominous Parcel", "Artifact", sncParcelOracle, false)
	big := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 4, 4)
	if err := g.ActivateCatalogAbility(me.ID, parcel, 1, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	}); err == nil {
		t.Fatal("activated the damage ability with no mana")
	}
	apaMana(me, "C")
	b09AddColorless(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, parcel, 1, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(parcel) {
		t.Error("the Parcel was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("4 damage did not kill the 4/4")
	}
}

func TestSncB03HaloScarabExilesItselfForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	scarab, me := activateFromGraveyard(t, g, "Halo Scarab", "Artifact Creature — Insect", sncScarabOracle, 2, 1, "{C}{C}",
		game.ActivateAbilityParams{})
	if me.Graveyard.Contains(scarab) {
		t.Fatal("the Scarab was not exiled as part of the cost")
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("no Treasure token")
	}
}

func TestSncB03GoldhoundSacrificesForOneManaOfAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	dog := pushCatalogPermanent(g, me.ID, "Goldhound", "Artifact Creature — Treasure Dog", sncGoldhoundOracle, false)
	spec, _ := Lookup(sncGoldhoundOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" || !spec.ManaAbilities[0].Cost.Sacrifice || !spec.ManaAbilities[0].Cost.Tap {
		t.Fatalf("mana ability %+v, want tap-and-sacrifice for any colour", spec.ManaAbilities)
	}
	assertKeywords(t, g, dog, "first strike", "menace")
	if err := g.ActivateManaAbility(me.ID, dog, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if pick := pendingOfKind(g, game.PendingChoiceMana); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
			t.Fatalf("choose a colour: %v", err)
		}
	}
	if g.Battlefield.Contains(dog) {
		t.Error("the Goldhound was not sacrificed")
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("the pool holds %d mana, want 1", me.ManaPool.Total())
	}
}

func TestSncB03CementShoesPumpsTapsAtEndStepAndLocksUntap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shoes := pushPermanentForTest(g, me.ID, "Cement Shoes", sncCementShoesOracle, "Artifact — Equipment")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(shoes, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
	})
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the equipped Bear is %d/%d, want 5/5", p, tough)
	}
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !apaLive(g, bear).Tapped {
		t.Fatal("the equipped creature was not tapped at the beginning of its controller's end step")
	}
	// The creature stays tapped through its controller's next untap step.
	startTurn := g.Turn.Seq
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepUpkeep || g.Turn.Seq <= startTurn {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !apaLive(g, bear).Tapped {
		t.Error("the equipped creature untapped in its untap step")
	}
}
