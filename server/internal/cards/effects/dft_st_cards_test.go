package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_st_cards_test.go — Aetherdrift (DFT), cards from S to T (#4464).
// Each test plays the real catalog entry and asserts what its text says.
// The fixtures are dft_ab_cards_test.go's.

const (
	dftSTSundialOracle      = "9ca808c5-9338-4199-a45b-df1a1a20625e"
	dftSTTerrianOracle      = "856245a6-c697-4f03-90be-77d8eb2d274e"
	dftSTTyroxOracle        = "b24d95fa-3cf7-495e-9082-11f8db151906"
	dftSTTuneUpOracle       = "aaa14260-2537-4668-8517-c825118777cc"
	dftSTTripUpOracle       = "bac52ae6-24ee-4c4f-a32c-3eb9859dc025"
	dftSTTransitMageOracle  = "991efbf1-299d-4892-aeab-8ff9827f7168"
	dftSTTradeTheHelmOracle = "c478fb5c-9fde-400d-a2cd-f8c4fa6b957a"
	dftSTTortoiseOracle     = "952083c5-9ffa-4cd9-915f-b7de7f3e5326"
	dftSTBroodwagonOracle   = "817464d8-1e2d-4e12-96f8-5c7ec9b3e681"
	dftSTGunnerOracle       = "d9b6656d-d71f-4359-bf6f-b76089741e29"
	dftSTSyphonFuelOracle   = "8696eae4-3dc9-4572-aab9-aa4f7e8f53af"
	dftSTStallOutOracle     = "ce55c0a9-3334-440e-ae1b-c04a9e1c3d87"
	dftSTScurryfootOracle   = "7b860806-0118-40f8-861b-225a7c1cf151"
)

func TestDftSTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftSTSundialOracle:      "Sundial, Dawn Tyrant",
		dftSTTerrianOracle:      "Terrian, World Tyrant",
		dftSTTyroxOracle:        "Tyrox, Saurid Tyrant",
		dftSTTuneUpOracle:       "Tune Up",
		dftSTTripUpOracle:       "Trip Up",
		dftSTTransitMageOracle:  "Transit Mage",
		dftSTTradeTheHelmOracle: "Trade the Helm",
		dftSTTortoiseOracle:     "Ticket Tortoise",
		dftSTBroodwagonOracle:   "Thundering Broodwagon",
		dftSTGunnerOracle:       "Thunderhead Gunner",
		dftSTSyphonFuelOracle:   "Syphon Fuel",
		dftSTStallOutOracle:     "Stall Out",
		dftSTScurryfootOracle:   "Stampeding Scurryfoot",
	} {
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

// --- fixtures ------------------------------------------------------------------

// dftSTCostly puts a permanent with a printed mana cost onto the
// battlefield under owner's control.
func dftSTCostly(g *game.Game, owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// dftSTCycles puts the card into the active player's hand, cycles it
// with its cycling row at idx, and checks that it was discarded for a
// card.
func dftSTCycles(t *testing.T, name, typeLine, oracle string, idx int) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedLibrary(me, "Drawn")
	id := pushCatalogHandCard(me, name, typeLine, oracle)
	hand := me.Hand.Size()
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: cycle: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Errorf("%s was not discarded to cycle", name)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("%s: hand is %d after cycling, want %d (one discarded, one drawn)", name, got, hand)
	}
}

// --- the legendary vanillas -------------------------------------------------------

func TestDftSTTheTyrantsResolveAsPlainCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Sundial, Dawn Tyrant", "Legendary Artifact Creature — Construct", dftSTSundialOracle, 3, 3},
		{"Terrian, World Tyrant", "Legendary Creature — Dinosaur Ooze", dftSTTerrianOracle, 9, 7},
		{"Tyrox, Saurid Tyrant", "Legendary Creature — Dinosaur Warrior", dftSTTyroxOracle, 4, 1},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		id := dftABCast(t, g, me, game.Card{Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
			Power: tc.power, Toughness: tc.toughness})
		if !g.Battlefield.Contains(id) {
			t.Fatalf("%s did not enter the battlefield", tc.name)
		}
		if p, tough := effectivePTOf(t, g, id); p != tc.power || tough != tc.toughness {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.toughness)
		}
		if len(g.PendingChoices) != 0 || g.Stack.Size() != 0 {
			t.Errorf("%s left something behind: choices %v, stack %d", tc.name, g.PendingChoices, g.Stack.Size())
		}
	}
}

// --- Tune Up ----------------------------------------------------------------------

func TestDftSTTuneUpReturnsAnArtifactAndAnimatesAVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	vehicle := dftABInGraveyard(me, "Old Vehicle", dftABVehicle)
	bear := dftABInGraveyard(me, "Bear", "Creature — Bear")
	if err := castCatalogSpellErr(t, g, "Tune Up", "Sorcery", dftSTTuneUpOracle, dftABCardRef(bear)); err == nil {
		t.Fatal("a creature card that is not an artifact was a legal target")
	}
	castCatalogSpell(t, g, "Tune Up", "Sorcery", dftSTTuneUpOracle, dftABCardRef(vehicle))
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, vehicle)
	if !ok {
		t.Fatal("the Vehicle did not return to the battlefield")
	}
	if !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("the returned Vehicle is not an artifact creature: %q", c.TypeLine)
	}
	if p, tough := effectivePTOf(t, g, vehicle); p != 2 || tough != 2 {
		t.Errorf("the returned Vehicle is %d/%d, want its printed 2/2", p, tough)
	}

	// It stays a creature past the end of the turn: the animation has
	// no duration.
	for g.Turn.Step != game.StepCleanup {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if c, _ := battlefieldCard(g, vehicle); !c.IsCreature() {
		t.Error("the Vehicle stopped being a creature after the turn ended")
	}
}

func TestDftSTTuneUpLeavesANonVehicleArtifactAsItIs(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	rock := dftABInGraveyard(me, "Mind Stone", "Artifact")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := dftABInGraveyard(opp, "Their Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Tune Up", "Sorcery", dftSTTuneUpOracle, dftABCardRef(theirs)); err == nil {
		t.Fatal("an artifact card in an opponent's graveyard was a legal target")
	}
	castCatalogSpell(t, g, "Tune Up", "Sorcery", dftSTTuneUpOracle, dftABCardRef(rock))
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, rock)
	if !ok {
		t.Fatal("the artifact did not return to the battlefield")
	}
	if c.IsCreature() {
		t.Error("an artifact that is not a Vehicle became a creature")
	}
}

// --- Trip Up ----------------------------------------------------------------------

func TestDftSTTripUpTucksANonlandPermanentByItsOwnersChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Trip Up", "Instant", dftSTTripUpOracle, dftABCardRef(land)); err == nil {
		t.Fatal("a land was a legal target")
	}
	castCatalogSpell(t, g, "Trip Up", "Instant", dftSTTripUpOracle, dftABCardRef(bear))
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Fatal("the caster was asked; the owner chooses")
	}
	c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil || c.LibraryPlacement != game.LibraryPlaceTopOrBottom {
		t.Fatalf("the owner was not asked top or bottom: %+v", c)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{bear}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(bear) || !opp.Library.Contains(bear) {
		t.Error("the creature is not in its owner's library")
	}
}

func TestDftSTTripUpCycles(t *testing.T) {
	dftSTCycles(t, "Trip Up", "Instant", dftSTTripUpOracle, 0)
}

// --- Transit Mage -----------------------------------------------------------------

func TestDftSTTransitMageMayFindAnArtifactOfManaValueFourOrFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	put := func(name, typeLine, cost string) uuid.UUID {
		return pushLibraryCardForTest(me, game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	}
	four := put("Four Rock", "Artifact", "{4}")
	five := put("Five Vehicle", dftABVehicle, "{3}{W}{W}")
	three := put("Three Rock", "Artifact", "{3}")
	six := put("Six Rock", "Artifact", "{6}")
	beast := put("Four Beast", "Creature — Beast", "{2}{G}{G}")

	dftABCast(t, g, me, game.Card{Name: "Transit Mage", TypeLine: "Creature — Human Wizard",
		OracleID: dftSTTransitMageOracle, Power: 2, Toughness: 2})
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search when the Mage entered")
	}
	for _, id := range []uuid.UUID{four, five} {
		if !slices.Contains(c.SearchCards, id) {
			t.Errorf("an artifact card of mana value 4 or 5 (%s) was not offered", id)
		}
	}
	for _, id := range []uuid.UUID{three, six, beast} {
		if slices.Contains(c.SearchCards, id) {
			t.Errorf("%s was offered; it is the wrong mana value or not an artifact", id)
		}
	}
	answerSearchByID(t, g, me.ID, five)
	if !me.Hand.Contains(five) {
		t.Error("the found card is not in hand")
	}
}

func TestDftSTTransitMageMayFindNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	four := pushLibraryCardForTest(me, game.Card{Name: "Four Rock", TypeLine: "Artifact", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	dftABCast(t, g, me, game.Card{Name: "Transit Mage", TypeLine: "Creature — Human Wizard",
		OracleID: dftSTTransitMageOracle, Power: 2, Toughness: 2})
	answerSearchFailToFind(t, g, me.ID)
	if me.Hand.Contains(four) || !me.Library.Contains(four) {
		t.Error("declining the search still took the card")
	}
}

// --- Trade the Helm ---------------------------------------------------------------

func TestDftSTTradeTheHelmExchangesAnArtifactOrCreatureWithAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	theirs := b12Permanent(g, opp.ID, "Their Vehicle", dftABVehicle)
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")

	if err := castCatalogSpellErr(t, g, "Trade the Helm", "Sorcery", dftSTTradeTheHelmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: other}}); err == nil {
		t.Fatal("the second target was mine; the text says an opponent controls it")
	}
	if err := castCatalogSpellErr(t, g, "Trade the Helm", "Sorcery", dftSTTradeTheHelmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: land}}); err == nil {
		t.Fatal("a land was a legal second target")
	}
	castCatalogSpell(t, g, "Trade the Helm", "Sorcery", dftSTTradeTheHelmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, mine); got != opp.ID {
		t.Error("the opponent did not gain control of my creature")
	}
	if got := controllerOf(t, g, theirs); got != me.ID {
		t.Error("I did not gain control of their Vehicle")
	}
}

func TestDftSTTradeTheHelmDoesNothingWhenATargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Trade the Helm", "Sorcery", dftSTTradeTheHelmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, mine); got != me.ID {
		t.Error("my creature changed hands although the exchange could not be completed")
	}
}

func TestDftSTTradeTheHelmCycles(t *testing.T) {
	dftSTCycles(t, "Trade the Helm", "Sorcery", dftSTTradeTheHelmOracle, 0)
}

// --- Ticket Tortoise --------------------------------------------------------------

func TestDftSTTicketTortoiseMakesATreasureWhenBehindOnLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	tortoise := dftABCast(t, g, me, game.Card{Name: "Ticket Tortoise", TypeLine: "Artifact Creature — Turtle",
		OracleID: dftSTTortoiseOracle, Power: 3, Toughness: 1})
	if !hasEffectiveKeyword(t, g, tortoise, "defender") {
		t.Error("Ticket Tortoise has no defender")
	}
	if got := countTreasures(g, me.ID); got != 1 {
		t.Errorf("%d Treasures, want 1: an opponent controls more lands", got)
	}
}

func TestDftSTTicketTortoiseMakesNothingWhenNotBehind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	dftABCast(t, g, me, game.Card{Name: "Ticket Tortoise", TypeLine: "Artifact Creature — Turtle",
		OracleID: dftSTTortoiseOracle, Power: 3, Toughness: 1})
	if got := countTreasures(g, me.ID); got != 0 {
		t.Errorf("%d Treasures, want 0: nobody controls more lands than I do", got)
	}
}

// --- Thundering Broodwagon --------------------------------------------------------

func TestDftSTThunderingBroodwagonDestroysACheapNonlandPermanentOfAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	cheap := dftSTCostly(g, opp.ID, "Their Four-Drop", "Creature — Beast", "{2}{G}{G}")
	pricey := dftSTCostly(g, opp.ID, "Their Five-Drop", "Creature — Beast", "{3}{G}{G}")
	mine := dftSTCostly(g, me.ID, "My Two-Drop", "Creature — Bear", "{1}{G}")

	wagon := dftABCast(t, g, me, game.Card{Name: "Thundering Broodwagon", TypeLine: dftABVehicle,
		OracleID: dftSTBroodwagonOracle, Power: 6, Toughness: 5})
	for _, kw := range []string{"reach", "menace"} {
		if !hasEffectiveKeyword(t, g, wagon, kw) {
			t.Errorf("Thundering Broodwagon has no %s", kw)
		}
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target pick when the Broodwagon entered")
	}
	for _, bad := range []uuid.UUID{pricey, mine} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was a legal target", bad)
		}
	}
	pickTriggerTarget(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the mana value 4 permanent was not destroyed")
	}
	if !g.Battlefield.Contains(pricey) || !g.Battlefield.Contains(mine) {
		t.Error("a permanent that was not the target was destroyed")
	}
}

func TestDftSTThunderingBroodwagonCrewsThreeAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wagon := dftABPush(g, me.ID, "Thundering Broodwagon", dftABVehicle, dftSTBroodwagonOracle, 6, 5)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, wagon, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{two}}); err == nil {
		t.Fatal("crewed with total power 2; crew 3 wants 3")
	}
	dftABCrew(t, g, me, wagon, 0, two, one)

	dftSTCycles(t, "Thundering Broodwagon", dftABVehicle, dftSTBroodwagonOracle, 1)
}

// --- Thunderhead Gunner -----------------------------------------------------------

func TestDftSTThunderheadGunnerRummagesOnceEachTurnAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gunner := dftABPush(g, me.ID, "Thunderhead Gunner", "Creature — Shark Pirate", dftSTGunnerOracle, 4, 5)
	if !hasEffectiveKeyword(t, g, gunner, "reach") {
		t.Error("Thunderhead Gunner has no reach")
	}
	seedLibrary(me, "Drawn A", "Drawn B")
	first := handCard(me, "First", "Sorcery")
	second := handCard(me, "Second", "Sorcery")
	// The table starts in the draw step, not a main phase.
	if err := g.ActivateCatalogAbility(me.ID, gunner, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{first}}); err == nil {
		t.Fatal("activated in the draw step; the text says only as a sorcery")
	}
	advanceToMain(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gunner, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{first}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(first) {
		t.Error("the discarded card is not in the graveyard")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d after the rummage, want %d", got, hand)
	}
	if err := g.ActivateCatalogAbility(me.ID, gunner, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{second}}); err == nil {
		t.Error("activated a second time this turn; the text says only once each turn")
	}
}

// --- Syphon Fuel ------------------------------------------------------------------

func TestDftSTSyphonFuelShrinksTheTargetAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 7, 7)
	life := me.Life
	castCatalogSpell(t, g, "Syphon Fuel", "Instant", dftSTSyphonFuelOracle, dftABCardRef(big))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestDftSTSyphonFuelDoesNothingWhenItsTargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Syphon Fuel", "Instant", dftSTSyphonFuelOracle, dftABCardRef(bear))
	killCreature(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d: a spell whose only target is gone does nothing", me.Life, life)
	}
}

// --- Stall Out --------------------------------------------------------------------

func TestDftSTStallOutTapsAndPutsThreeStunCounters(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := b12Permanent(g, opp.ID, "Their Vehicle", dftABVehicle)
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Stall Out", "Sorcery", dftSTStallOutOracle, dftABCardRef(land)); err == nil {
		t.Fatal("a land was a legal target")
	}
	castCatalogSpell(t, g, "Stall Out", "Sorcery", dftSTStallOutOracle, dftABCardRef(vehicle))
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, vehicle)
	if !c.Tapped {
		t.Error("the Vehicle is not tapped")
	}
	if got := countersOn(g, vehicle, game.CounterStun); got != 3 {
		t.Errorf("%d stun counters, want 3", got)
	}
}

func TestDftSTStallOutCycles(t *testing.T) {
	dftSTCycles(t, "Stall Out", "Sorcery", dftSTStallOutOracle, 0)
}

// --- Stampeding Scurryfoot --------------------------------------------------------

func TestDftSTStampedingScurryfootExhaustsForACounterAndAnElephant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mouse := dftABPush(g, me.ID, "Stampeding Scurryfoot", "Creature — Mouse", dftSTScurryfootOracle, 1, 1)
	advanceToMain(t, g)
	fillPool(me, 3)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, mouse, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, mouse, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	var elephant uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elephant" && c.Controller == me.ID {
			elephant = c.InstanceID
		}
	}
	if elephant == uuid.Nil {
		t.Fatal("no Elephant token")
	}
	if p, tough := effectivePTOf(t, g, elephant); p != 3 || tough != 3 {
		t.Errorf("the Elephant is %d/%d, want 3/3", p, tough)
	}
	fillPool(me, 3)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, mouse, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

func TestDftSTStampedingScurryfootStillMakesTheElephantIfItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mouse := dftABPush(g, me.ID, "Stampeding Scurryfoot", "Creature — Mouse", dftSTScurryfootOracle, 1, 1)
	advanceToMain(t, g)
	fillPool(me, 3)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, mouse, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	killCreature(t, g, me.ID, mouse)
	passPriorityAroundTable(t, g)
	if got := countNamed(g, "Elephant"); got != 1 {
		t.Errorf("%d Elephants, want 1: the token does not need the Mouse", got)
	}
}
