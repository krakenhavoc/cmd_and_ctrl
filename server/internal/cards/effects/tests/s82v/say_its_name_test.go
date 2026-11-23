package testss82v

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	spec, ok := effects.Lookup(sayItsNameOracle)
	if !ok || spec.Name != sayItsNameName || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered %v, name %q, completeness %s, caveats %v", ok, spec.Name, spec.Completeness, spec.Caveats)
	}
}

func TestSayItsNameMillsThreeThenMayReturnACreatureOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	milled := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milled, Name: "Milled Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, sayItsNameName, "Sorcery", sayItsNameOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, milled) || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("want a 0..1 pick offering the milled creature: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, milled)
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
	g = effectstest.NewCatalogGame(t)
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
			effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
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
		effectstest.PushCatalogGraveyardCard(me, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0),
		effectstest.PushCatalogGraveyardCard(me, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0),
	}
	self, _ := effectstest.ActivateFromGraveyard(t, g, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0, "",
		game.ActivateAbilityParams{ExileIDs: others})
	for _, id := range append(others, self) {
		if me.Graveyard.Contains(id) {
			t.Fatal("a Say Its Name paid to the cost is still in the graveyard")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || !c.SearchZonesPending() || c.SearchOtherZone() != game.SearchZoneGraveyardAndHand {
		t.Fatalf("want the three-zone question: %+v", c)
	}
	return g, me, altanak, before
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
			effectstest.AnswerSearchZones(t, g, me.ID, zones...)
			c := effectstest.SearchChoiceFor(g, me.ID)
			for z, id := range altanak {
				if effectstest.HasID(c.SearchCards, id) != slices.Contains(zones, z) {
					t.Errorf("zones %v: the %s Altanak offered = %v", zones, z, effectstest.HasID(c.SearchCards, id))
				}
			}
			effectstest.AnswerSearchByID(t, g, me.ID, altanak[from])
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(altanak[from]) {
				t.Errorf("zones %v: the %s Altanak did not enter the battlefield", zones, from)
			}
			searched := slices.Contains(zones, game.SearchZoneLibrary)
			if (effectstest.SearchEventCount(g) != 0) != searched {
				t.Errorf("zones %v: %d search events, library searched %v", zones, effectstest.SearchEventCount(g), searched)
			}
			if !searched {
				if !slices.Equal(effectstest.LibraryOrder(me), before) {
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
		effectstest.AnswerSearchZones(t, g, me.ID, z)
		effectstest.AnswerSearchByID(t, g, me.ID, altanak[z])
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.BattlefieldCountNamed(g, me.ID, altanakName) != 1 {
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
		effectstest.AnswerSearchZones(t, g, me.ID, zones...)
		effectstest.AnswerSearchFailToFind(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.BattlefieldCountNamed(g, me.ID, altanakName) != 0 {
			t.Errorf("zones %v: something named Altanak entered", zones)
		}
		searched := slices.Contains(zones, game.SearchZoneLibrary)
		if (effectstest.SearchEventCount(g) != 0) != searched {
			t.Errorf("zones %v: %d search events, library searched %v", zones, effectstest.SearchEventCount(g), searched)
		}
		if !searched && !slices.Equal(effectstest.LibraryOrder(me), before) {
			t.Errorf("zones %v: the library changed though it was not searched", zones)
		}
	}
}

// The cost needs two OTHER cards named Say Its Name: one other is not
// enough, and a card with another name is not one.
func TestSayItsNameNeedsTwoOtherCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	other := effectstest.PushCatalogGraveyardCard(me, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0)
	stranger := effectstest.PushCatalogGraveyardCard(me, "Grapple with the Past", "Instant", effectstest.B20GrappleWithThePastOracle, 0, 0)
	self := effectstest.PushCatalogGraveyardCard(me, sayItsNameName, "Sorcery", sayItsNameOracle, 0, 0)
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
	c := effectstest.SearchChoiceFor(g, me.ID)
	effectstest.BeginSearchIfAsked(t, g, c)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if !effectstest.HasID(c.SearchCards, altanak[game.SearchZoneLibrary]) {
		t.Fatalf("searching every zone does not offer the library's Altanak: %v", c.SearchCards)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, altanak[game.SearchZoneLibrary])
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldCountNamed(g, me.ID, altanakName) != 1 || effectstest.SearchEventCount(g) != 1 {
		t.Errorf("Altanak entered %d, search events %d", effectstest.BattlefieldCountNamed(g, me.ID, altanakName), effectstest.SearchEventCount(g))
	}
}
