package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_graveyard_search_test.go — #3217, "search your library and/or
// graveyard", through the real catalog cards (Tower Winder, Finale of
// Devastation, Ajani's Aid, and the five cards that shipped with the
// seam). The engine rule is pinned beside the code in
// game/search_graveyard_test.go; this file pins that the cards inherit
// it.
//
//   - a card in the graveyard is found, and a graveyard-only take does
//     not shuffle the library or fire EventSearchLibrary;
//   - a library take shuffles, and fires the event;
//   - while a matching card is in the graveyard the search cannot fail
//     to find (a public zone has no fail-to-find, CR 701.23b);
//   - the prompt offers both zones, and says which is which.

const (
	lgTowerWinderOracle = "c1aecc8a-db3a-4158-b1f1-bfe06d782482"
	lgFinaleOracle      = "69872a9a-fe54-4e58-940c-89395af71acd"
	lgAjanisAidOracle   = "b3779fbe-7701-433b-afad-bebd6096a3c8"
	lgMoogleOracle      = "60043b87-6008-4cfa-9d8f-6ee6e3f70766"
)

// lgKnowLibrary marks every library card known to the seat, so a
// shuffle (which un-knows the zone) is observable as a change.
func lgKnowLibrary(p *game.Player) {
	for i := range p.Library.Cards {
		p.Library.Cards[i].AddKnower(p.ID)
	}
}

// lgCastTowerWinder casts Tower Winder and lets the enters trigger
// resolve to the point where the search either prompted or finished.
func lgCastTowerWinder(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Tower Winder", "Creature — Snake", lgTowerWinderOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
}

func lgTower() game.Card { return game.Card{Name: "Command Tower", TypeLine: "Land"} }

// The seam itself: a Command Tower that is only in the graveyard is
// found. One public match and nothing else on offer leaves nothing to
// decide, so it is taken unasked — and the library, never looked at, is
// not shuffled and no search is announced.
func TestTowerWinderFindsATowerInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	lgKnowLibrary(me)
	tower := pushGraveyardCardWithTypeLineID(me, "Command Tower", "Land")

	lgCastTowerWinder(t, g)

	if !me.Hand.Contains(tower) || me.Graveyard.Contains(tower) {
		t.Fatal("the Command Tower in the graveyard was not put into hand")
	}
	if searchEventCount(g) != 0 {
		t.Errorf("EventSearchLibrary fired %d times: the library was never searched", searchEventCount(g))
	}
	if libraryKnownTo(me, me.ID) != 1 {
		t.Error("the library was shuffled although nothing was taken from it")
	}
}

// A Tower in the library only: ordinary stated-quality search. It may be
// declined, and taking it shuffles.
func TestTowerWinderLibraryOnlyShufflesAndMayFailToFind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	tower := pushLibraryCardForTest(me, lgTower())
	lgKnowLibrary(me)

	lgCastTowerWinder(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("a library match for a stated quality should prompt")
	}
	if c.SearchMin != 0 || len(c.SearchGraveyardCards) != 0 {
		t.Fatalf("min/graveyard = %d/%d, want 0/0", c.SearchMin, len(c.SearchGraveyardCards))
	}
	answerSearchByID(t, g, me.ID, tower)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(tower) {
		t.Error("the Tower should be in hand")
	}
	if searchEventCount(g) != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1", searchEventCount(g))
	}
	if libraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was not shuffled")
	}
}

// Neither zone has one: the search finds nothing, and since the library
// was searched it is shuffled and the event fires.
func TestTowerWinderWithNoTowerAnywhereStillShuffles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	lgKnowLibrary(me)
	pushGraveyardCardWithTypeLineID(me, "Island", "Basic Land — Island")

	lgCastTowerWinder(t, g)

	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("nothing matches, so there is nothing to prompt")
	}
	if searchEventCount(g) != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1", searchEventCount(g))
	}
	if libraryKnownTo(me, me.ID) != 0 {
		t.Error("an empty library search must still shuffle")
	}
}

// Both zones hold a Tower. The prompt offers both and marks the
// graveyard one; the searcher cannot fail to find, because one of the
// matches is public. Taking the library one shuffles; taking the
// graveyard one does not.
func TestTowerWinderBothZonesOfferedAndMustFind(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		libTower := pushLibraryCardForTest(me, lgTower())
		lgKnowLibrary(me)
		gyTower := pushGraveyardCardWithTypeLineID(me, "Command Tower", "Land")
		lgCastTowerWinder(t, g)
		return g, me, libTower, gyTower
	}

	t.Run("both are offered, marked by zone, with a floor of one", func(t *testing.T) {
		g, me, libTower, gyTower := setup(t)
		c := searchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("two matches should prompt")
		}
		if len(c.SearchCards) != 2 || c.SearchMax != 1 || c.SearchMin != 1 {
			t.Fatalf("candidates/max/min = %d/%d/%d, want 2/1/1", len(c.SearchCards), c.SearchMax, c.SearchMin)
		}
		if len(c.SearchGraveyardCards) != 1 || c.SearchGraveyardCards[0] != gyTower {
			t.Fatalf("graveyard candidates = %v, want just %s", c.SearchGraveyardCards, gyTower)
		}
		offered := map[uuid.UUID]bool{}
		for _, id := range c.SearchCards {
			offered[id] = true
		}
		if !offered[libTower] || !offered[gyTower] {
			t.Errorf("offered %v, want both Towers", c.SearchCards)
		}
	})

	t.Run("failing to find is refused while a graveyard match exists", func(t *testing.T) {
		g, me, _, _ := setup(t)
		c := searchChoiceFor(g, me.ID)
		err := g.ResolveSearchLibrary(c.ID, me.ID, nil)
		if !errors.Is(err, game.ErrInvalidParam) {
			t.Fatalf("an empty answer = %v, want ErrInvalidParam (CR 701.23b is a hidden-zone rule)", err)
		}
		if searchChoiceFor(g, me.ID) == nil {
			t.Error("a refused answer must leave the prompt open")
		}
	})

	t.Run("taking the library Tower shuffles", func(t *testing.T) {
		g, me, libTower, _ := setup(t)
		answerSearchByID(t, g, me.ID, libTower)
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(libTower) {
			t.Fatal("library Tower not in hand")
		}
		if searchEventCount(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
			t.Errorf("events=%d known=%d: a library take must announce the search and shuffle",
				searchEventCount(g), libraryKnownTo(me, me.ID))
		}
	})

	t.Run("taking the graveyard Tower does not", func(t *testing.T) {
		g, me, _, gyTower := setup(t)
		known := libraryKnownTo(me, me.ID)
		answerSearchByID(t, g, me.ID, gyTower)
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(gyTower) {
			t.Fatal("graveyard Tower not in hand")
		}
		if searchEventCount(g) != 0 {
			t.Errorf("EventSearchLibrary fired %d times: the library was not searched", searchEventCount(g))
		}
		if libraryKnownTo(me, me.ID) != known {
			t.Error("the library was shuffled although the take came from the graveyard")
		}
	})
}

// A candidate is checked against the zone it was offered from: a card
// that was offered from the library cannot be claimed by turning up in
// the graveyard mid-prompt.
func TestTowerWinderLibraryCandidateMilledMidPromptIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	libTower := pushLibraryCardForTest(me, lgTower())
	pushGraveyardCardWithTypeLineID(me, "Command Tower", "Land")
	lgCastTowerWinder(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("expected a prompt")
	}
	g.WithWriteLock(func() {
		card, err := me.Library.Remove(libTower)
		if err != nil {
			t.Fatalf("remove: %v", err)
		}
		me.Graveyard.PushTop(card)
	})
	err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{libTower})
	if !errors.Is(err, game.ErrCardNotFound) {
		t.Fatalf("a library candidate now in the graveyard = %v, want ErrCardNotFound", err)
	}
}

// Finale of Devastation: a creature in the graveyard at mana value <= X
// is found and put onto the battlefield, the pump still runs, and a
// graveyard creature above X is not offered.
func TestFinaleOfDevastationFindsACreatureInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := pushVanillaCreature(g, me.ID, "Already Here", 2, 2)
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	lgKnowLibrary(me)
	gyCreature := pushGraveyardPermanent(me, "Buried Beast", "Creature — Beast", "{9}")
	tooBig := pushGraveyardPermanent(me, "Buried Titan", "Creature — Giant", "{11}")

	castXSpell(t, g, "Finale of Devastation", "Sorcery", lgFinaleOracle, "{X}{G}{G}", 10, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(gyCreature) {
		t.Error("the graveyard creature with mana value <= X should be on the battlefield")
	}
	if g.Battlefield.Contains(tooBig) || !me.Graveyard.Contains(tooBig) {
		t.Error("a mana value 11 creature is not a legal find for X=10")
	}
	if searchEventCount(g) != 0 || libraryKnownTo(me, me.ID) != 1 {
		t.Errorf("events=%d known=%d: a graveyard take must not search or shuffle the library",
			searchEventCount(g), libraryKnownTo(me, me.ID))
	}
	// "If X is 10 or more" is not conditioned on the search.
	if p := effectivePower(t, g, other); p != 12 {
		t.Errorf("Already Here power = %d, want 12", p)
	}
}

// With a graveyard creature AND a library creature on offer the prompt
// shows both and the searcher must take one; taking the library one
// shuffles.
func TestFinaleOfDevastationOffersBothZones(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me)
	libCreature := pushLibraryCardForTest(me, game.Card{Name: "Library Beast", TypeLine: "Creature — Beast", ManaCost: "{4}"})
	lgKnowLibrary(me)
	gyCreature := pushGraveyardPermanent(me, "Buried Beast", "Creature — Beast", "{3}")

	castXSpell(t, g, "Finale of Devastation", "Sorcery", lgFinaleOracle, "{X}{G}{G}", 5, nil)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two creatures on offer should prompt")
	}
	if len(c.SearchCards) != 2 || c.SearchMin != 1 || len(c.SearchGraveyardCards) != 1 {
		t.Fatalf("candidates/min/graveyard = %d/%d/%d, want 2/1/1",
			len(c.SearchCards), c.SearchMin, len(c.SearchGraveyardCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err == nil {
		t.Fatal("failing to find with a graveyard match on offer must be refused")
	}
	answerSearchByID(t, g, me.ID, libCreature)
	if !g.Battlefield.Contains(libCreature) || !me.Graveyard.Contains(gyCreature) {
		t.Error("only the library creature should have moved")
	}
	if libraryKnownTo(me, me.ID) != 0 || searchEventCount(g) != 1 {
		t.Error("taking the library creature must shuffle and announce the search")
	}
}

// A Finale whose only graveyard creature is too big for X is just a
// library search: failing to find stays legal.
func TestFinaleOfDevastationGraveyardTooBigLeavesFailToFindLegal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me)
	pushLibraryCardForTest(me, game.Card{Name: "Library Beast", TypeLine: "Creature — Beast", ManaCost: "{2}"})
	pushGraveyardPermanent(me, "Buried Titan", "Creature — Giant", "{11}")

	castXSpell(t, g, "Finale of Devastation", "Sorcery", lgFinaleOracle, "{X}{G}{G}", 3, nil)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil || c.SearchMin != 0 || len(c.SearchGraveyardCards) != 0 {
		t.Fatalf("want a prompt with no graveyard half and a floor of 0, got %+v", c)
	}
	answerSearchFailToFind(t, g, me.ID)
	if libraryKnownTo(me, me.ID) != 0 {
		t.Error("failing to find in the library still shuffles it")
	}
}

// Ajani's Aid: "you may search" — declining neither searches nor
// shuffles; saying yes with Ajani in the graveyard finds it.
func TestAjanisAidFindsAjaniInTheGraveyard(t *testing.T) {
	cast := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		lgKnowLibrary(me)
		ajani := pushGraveyardCardWithTypeLineID(me, "Ajani, Valiant Protector", "Legendary Planeswalker — Ajani")
		castCatalogSpell(t, g, "Ajani's Aid", "Enchantment", lgAjanisAidOracle, nil)
		passPriorityAroundTable(t, g)
		passPriorityAroundTable(t, g)
		return g, me, ajani
	}

	t.Run("yes", func(t *testing.T) {
		g, me, ajani := cast(t)
		answerMayChoice(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(ajani) {
			t.Fatal("Ajani in the graveyard was not put into hand")
		}
		if searchEventCount(g) != 0 || libraryKnownTo(me, me.ID) != 1 {
			t.Error("a graveyard take must not shuffle the library")
		}
	})

	t.Run("no", func(t *testing.T) {
		g, me, ajani := cast(t)
		answerMayChoice(t, g, me.ID, false)
		passPriorityAroundTable(t, g)
		if me.Hand.Contains(ajani) || !me.Graveyard.Contains(ajani) {
			t.Error("declining must leave Ajani in the graveyard")
		}
		if searchEventCount(g) != 0 || libraryKnownTo(me, me.ID) != 1 {
			t.Error("declining neither searches nor shuffles")
		}
	})
}

// The five cards that shipped with the seam, one case each. The four
// "named" creatures share one builder; Delivery Moogle is the only
// compulsory one.
func TestDeliveryMoogleTakesAnArtifactFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	lgKnowLibrary(me)
	cheap := pushGraveyardCardWithTypeLineID(me, "Cheap Rock", "Artifact")
	// pushGraveyardCardWithTypeLineID leaves the mana cost empty (value 0);
	// a dear artifact needs one.
	dear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear Rock", TypeLine: "Artifact", ManaCost: "{5}", Owner: me.ID, Controller: me.ID})

	castCatalogSpell(t, g, "Delivery Moogle", "Creature — Moogle", lgMoogleOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(cheap) {
		t.Error("the artifact card with mana value <= 2 should be in hand")
	}
	if me.Hand.Contains(dear) {
		t.Error("a mana value 5 artifact is not a legal find")
	}
	if searchEventCount(g) != 0 {
		t.Error("a graveyard take must not search the library")
	}
}

func TestNamedPlaneswalkerCreaturesFetchFromEitherZone(t *testing.T) {
	cases := []struct {
		name, oracle, typeLine, fetches string
	}{
		{"Goldmane Griffin", "e32573e8-91a8-438b-966f-1c6d098d54ae", "Creature — Griffin", "Ajani, Inspiring Leader"},
		{"Chandra's Firemaw", "53b17c51-a216-484f-80df-3904284df0ea", "Creature — Hellion", "Chandra, Flame's Catalyst"},
		{"Teferi's Wavecaster", "b3f6e68e-9490-4719-8d2e-d1ec22a96132", "Creature — Merfolk Wizard", "Teferi, Timeless Voyager"},
		{"Yanling's Harbinger", "c4bf4397-22e5-458d-b9d1-3f3a0a27e9e5", "Creature — Bird", "Mu Yanling, Celestial Wind"},
	}
	for _, tc := range cases {
		t.Run(tc.name+" graveyard", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			lgKnowLibrary(me)
			pw := pushGraveyardCardWithTypeLineID(me, tc.fetches, "Legendary Planeswalker")
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			passPriorityAroundTable(t, g)
			answerMayChoice(t, g, me.ID, true)
			passPriorityAroundTable(t, g)
			if !me.Hand.Contains(pw) {
				t.Fatalf("%s was not fetched from the graveyard", tc.fetches)
			}
			if libraryKnownTo(me, me.ID) != 1 {
				t.Error("a graveyard take must not shuffle the library")
			}
		})
		t.Run(tc.name+" library", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			pw := pushLibraryCardForTest(me, game.Card{Name: tc.fetches, TypeLine: "Legendary Planeswalker"})
			lgKnowLibrary(me)
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			passPriorityAroundTable(t, g)
			answerMayChoice(t, g, me.ID, true)
			answerSearchByID(t, g, me.ID, pw)
			passPriorityAroundTable(t, g)
			if !me.Hand.Contains(pw) {
				t.Fatalf("%s was not fetched from the library", tc.fetches)
			}
			if libraryKnownTo(me, me.ID) != 0 {
				t.Error("a library take must shuffle")
			}
		})
	}
}
