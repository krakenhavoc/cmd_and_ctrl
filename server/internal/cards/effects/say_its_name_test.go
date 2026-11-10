package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// say_its_name_test.go — #5408: Say Its Name. The spell mills three and may
// return a creature or land card; the graveyard ability exiles it and two
// other copies to search the graveyard, hand, and/or library for Altanak,
// the Thrice-Called, shuffling only a library that was searched.

const (
	sayItsNameOracle = "474f0675-4088-4f77-8081-f9ed7d492837"
	sayItsNameName   = "Say Its Name"
	altanakName      = "Altanak, the Thrice-Called"
)

func TestSayItsNameIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(sayItsNameOracle)
	if !ok || spec.Name != sayItsNameName || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered %v, name %q, completeness %s, caveats %v", ok, spec.Name, spec.Completeness, spec.Caveats)
	}
}

func TestSayItsNameMillsThreeThenMayReturnACreatureOrLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	for i := 0; i < 5; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	milled := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milled, Name: "Milled Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, sayItsNameName, "Sorcery", sayItsNameOracle, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || !hasID(pick.ChooseCards, milled) || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("want a 0..1 pick offering the milled creature: %+v", pick)
	}
	answerChooseCards(t, g, me.ID, milled)
	if !me.Hand.Contains(milled) {
		t.Error("the milled creature card was not returned to hand")
	}
	if len(me.Library.Cards) != 3 {
		t.Errorf("library has %d cards, want 3 after milling three", len(me.Library.Cards))
	}
}

// sayItsNameTable is a table with Altanak in the zones named (any of
// "library", "graveyard", "hand"), filler in the library, two other copies
// of Say Its Name in the graveyard, and the ability activated from the
// third and resolved up to its search prompt. It returns the Altanaks by
// zone and the library's order before the search.
func sayItsNameTable(t *testing.T, altanakIn ...string) (g *game.Game, me *game.Player, altanak map[string]uuid.UUID, before []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	altanak = map[string]uuid.UUID{}
	mk := func() game.Card {
		return game.Card{InstanceID: uuid.New(), Name: altanakName, TypeLine: "Legendary Creature — Insect Beast",
			ManaCost: "{5}{G}{G}", Power: 10, Toughness: 10, Owner: me.ID, Controller: me.ID}
	}
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		me.Hand.Cards = nil
		for i := 0; i < 10; i++ {
			pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
		}
		for _, z := range altanakIn {
			c := mk()
			altanak[z] = c.InstanceID
			switch z {
			case game.SearchZoneLibrary:
				me.Library.Cards = slices.Insert(me.Library.Cards, 4, c)
			case game.SearchZoneGraveyard:
				me.Graveyard.PushTop(c)
			case game.SearchZoneHand:
				me.Hand.PushTop(c)
			}
		}
		for _, c := range me.Library.Cards {
			before = append(before, c.InstanceID)
		}
	})
	others := []uuid.UUID{
		pushCatalogGraveyardCard(me, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0),
		pushCatalogGraveyardCard(me, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0),
	}
	self, _ := activateFromGraveyard(t, g, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0, "",
		game.ActivateAbilityParams{ExileIDs: others})
	for _, id := range append(others, self) {
		if me.Graveyard.Contains(id) {
			t.Fatal("a Say Its Name paid to the cost is still in the graveyard")
		}
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || !c.SearchZonesPending() || c.SearchOtherZone() != game.SearchZoneGraveyardAndHand {
		t.Fatalf("want the three-zone question: %+v", c)
	}
	return g, me, altanak, before
}

func libraryOrder(me *game.Player) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range me.Library.Cards {
		ids = append(ids, c.InstanceID)
	}
	return ids
}

// Every non-empty subset of the three zones finds Altanak in a zone it
// searched, and only a search that includes the library shuffles it.
func TestSayItsNameEachZoneSubsetFindsAltanakAndOnlyTheLibraryShuffles(t *testing.T) {
	subsets := [][]string{
		{game.SearchZoneLibrary},
		{game.SearchZoneGraveyard},
		{game.SearchZoneHand},
		{game.SearchZoneLibrary, game.SearchZoneGraveyard},
		{game.SearchZoneLibrary, game.SearchZoneHand},
		{game.SearchZoneGraveyard, game.SearchZoneHand},
		{game.SearchZoneLibrary, game.SearchZoneGraveyard, game.SearchZoneHand},
	}
	for _, zones := range subsets {
		for _, from := range zones {
			g, me, altanak, before := sayItsNameTable(t, game.SearchZoneLibrary, game.SearchZoneGraveyard, game.SearchZoneHand)
			answerSearchZones(t, g, me.ID, zones...)
			c := searchChoiceFor(g, me.ID)
			for z, id := range altanak {
				if hasID(c.SearchCards, id) != slices.Contains(zones, z) {
					t.Errorf("zones %v: the %s Altanak offered = %v", zones, z, hasID(c.SearchCards, id))
				}
			}
			answerSearchByID(t, g, me.ID, altanak[from])
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(altanak[from]) {
				t.Errorf("zones %v: the %s Altanak did not enter the battlefield", zones, from)
			}
			searched := slices.Contains(zones, game.SearchZoneLibrary)
			if (searchEventCount(g) != 0) != searched {
				t.Errorf("zones %v: %d search events, library searched %v", zones, searchEventCount(g), searched)
			}
			if !searched {
				if !slices.Equal(libraryOrder(me), before) {
					t.Errorf("zones %v: the library changed though it was not searched", zones)
				}
			}
		}
	}
}

// Altanak in one zone only: the matching single-zone search finds it.
func TestSayItsNameFindsAltanakWhereverItIs(t *testing.T) {
	for _, z := range []string{game.SearchZoneLibrary, game.SearchZoneGraveyard, game.SearchZoneHand} {
		g, me, altanak, _ := sayItsNameTable(t, z)
		answerSearchZones(t, g, me.ID, z)
		answerSearchByID(t, g, me.ID, altanak[z])
		passPriorityAroundTable(t, g)
		if battlefieldCountNamed(g, me.ID, altanakName) != 1 {
			t.Errorf("Altanak from the %s did not enter", z)
		}
	}
}

// No Altanak anywhere: the search finds nothing, which is fine, and the
// library is shuffled only if it was searched.
func TestSayItsNameFindsNothingAndShufflesOnlyASearchedLibrary(t *testing.T) {
	for _, zones := range [][]string{
		{game.SearchZoneGraveyard, game.SearchZoneHand},
		{game.SearchZoneLibrary, game.SearchZoneGraveyard, game.SearchZoneHand},
	} {
		g, me, _, before := sayItsNameTable(t)
		answerSearchZones(t, g, me.ID, zones...)
		answerSearchFailToFind(t, g, me.ID)
		passPriorityAroundTable(t, g)
		if battlefieldCountNamed(g, me.ID, altanakName) != 0 {
			t.Errorf("zones %v: something named Altanak entered", zones)
		}
		searched := slices.Contains(zones, game.SearchZoneLibrary)
		if (searchEventCount(g) != 0) != searched {
			t.Errorf("zones %v: %d search events, library searched %v", zones, searchEventCount(g), searched)
		}
		if !searched && !slices.Equal(libraryOrder(me), before) {
			t.Errorf("zones %v: the library changed though it was not searched", zones)
		}
	}
}

// The cost needs two OTHER cards named Say Its Name: one other is not
// enough, and a card with another name is not one.
func TestSayItsNameNeedsTwoOtherCopies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	other := pushCatalogGraveyardCard(me, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0)
	stranger := pushCatalogGraveyardCard(me, "Grapple with the Past", "Instant", b20GrappleWithThePastOracle, 0, 0)
	self := pushCatalogGraveyardCard(me, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0)
	for _, ids := range [][]uuid.UUID{{other}, {other, stranger}, {other, self}} {
		if err := g.ActivateCatalogAbility(me.ID, self, 0, game.ActivateAbilityParams{ExileIDs: ids}); err == nil {
			t.Fatalf("activated with %v exiled", ids)
		}
	}
	if !me.Graveyard.Contains(self) || !me.Graveyard.Contains(other) {
		t.Error("a refused activation moved a card")
	}
}

// The bot seat answers the three-zone question with every zone and takes
// Altanak (the seat's heuristic is covered in aiseat/heuristic; this is the
// engine side of that answer).
func TestSayItsNameSearchingEveryZoneFindsTheLibraryAltanak(t *testing.T) {
	g, me, altanak, _ := sayItsNameTable(t, game.SearchZoneLibrary)
	c := searchChoiceFor(g, me.ID)
	beginSearchIfAsked(t, g, c)
	c = searchChoiceFor(g, me.ID)
	if !hasID(c.SearchCards, altanak[game.SearchZoneLibrary]) {
		t.Fatalf("searching every zone does not offer the library's Altanak: %v", c.SearchCards)
	}
	answerSearchByID(t, g, me.ID, altanak[game.SearchZoneLibrary])
	passPriorityAroundTable(t, g)
	if battlefieldCountNamed(g, me.ID, altanakName) != 1 || searchEventCount(g) != 1 {
		t.Errorf("Altanak entered %d, search events %d", battlefieldCountNamed(g, me.ID, altanakName), searchEventCount(g))
	}
}
