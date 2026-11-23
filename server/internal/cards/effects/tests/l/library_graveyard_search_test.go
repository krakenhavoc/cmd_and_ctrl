package testsl

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The seam itself: a Command Tower that is only in the graveyard is
// found, and the library, never looked at, is not shuffled and no search
// is announced.
func TestTowerWinderFindsATowerInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	tower := effectstest.PushGraveyardCardWithTypeLineID(me, "Command Tower", "Land")

	effectstest.LgCastTowerWinder(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchByID(t, g, me.ID, tower)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Hand.Contains(tower) || me.Graveyard.Contains(tower) {
		t.Fatal("the Command Tower in the graveyard was not put into hand")
	}
	if effectstest.SearchEventCount(g) != 0 {
		t.Errorf("EventSearchLibrary fired %d times: the library was never searched", effectstest.SearchEventCount(g))
	}
	if effectstest.LibraryKnownTo(me, me.ID) != 1 {
		t.Error("the library was shuffled although nothing was taken from it")
	}
}

// A Tower in the library only: ordinary stated-quality search. It may be
// declined, and taking it shuffles.
func TestTowerWinderLibraryOnlyShufflesAndMayFailToFind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	tower := effectstest.PushLibraryCardForTest(me, effectstest.LgTower())
	effectstest.LgKnowLibrary(me)

	effectstest.LgCastTowerWinder(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("a library match for a stated quality should prompt")
	}
	if c.SearchMin != 0 || len(c.SearchGraveyardCards) != 0 {
		t.Fatalf("min/graveyard = %d/%d, want 0/0", c.SearchMin, len(c.SearchGraveyardCards))
	}
	effectstest.AnswerSearchByID(t, g, me.ID, tower)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Hand.Contains(tower) {
		t.Error("the Tower should be in hand")
	}
	if effectstest.SearchEventCount(g) != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1", effectstest.SearchEventCount(g))
	}
	if effectstest.LibraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was not shuffled")
	}
}

// Neither zone has one: the search finds nothing. "If you search your
// library this way, shuffle" (#4516, #4519): a player who chose the
// library shuffles it and the event fires; one who chose only the
// graveyard does not. The question is asked, and the empty pick after it,
// either way.
func TestTowerWinderWithNoTowerAnywhere(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		effectstest.LgKnowLibrary(me)
		effectstest.PushGraveyardCardWithTypeLineID(me, "Island", "Basic Land — Island")

		effectstest.LgCastTowerWinder(t, g)

		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil || len(c.SearchCards) != 0 {
			t.Fatalf("nothing matches, so the prompt (#4140) has no candidates: %+v", c)
		}
		if !c.SearchZonesPending() {
			t.Fatal("an \"and/or graveyard\" prompt must ask which zones are searched")
		}
		return g, me
	}

	t.Run("the library was searched, so it shuffles", func(t *testing.T) {
		g, me := setup(t)
		effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
		effectstest.AnswerSearchFailToFind(t, g, me.ID)
		if effectstest.SearchEventCount(g) != 1 {
			t.Errorf("EventSearchLibrary fired %d times, want 1", effectstest.SearchEventCount(g))
		}
		if effectstest.LibraryKnownTo(me, me.ID) != 0 {
			t.Error("an empty library search must still shuffle")
		}
	})

	t.Run("only the graveyard was searched, so it does not", func(t *testing.T) {
		g, me := setup(t)
		effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
		effectstest.AnswerSearchFailToFind(t, g, me.ID)
		if effectstest.SearchEventCount(g) != 0 {
			t.Errorf("EventSearchLibrary fired %d times, want 0", effectstest.SearchEventCount(g))
		}
		if effectstest.LibraryKnownTo(me, me.ID) != 1 {
			t.Error("a graveyard-only search must not shuffle the library")
		}
	})
}

// Both zones hold a Tower. Once both zones are chosen the prompt offers
// both and marks the graveyard one; the searcher may still find nothing
// ("and/or" lets them search only the library). The library is shuffled
// if and only if it was chosen (#4516, #4519), whatever is taken.
func TestTowerWinderBothZonesOfferedMayFindNothing(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		libTower := effectstest.PushLibraryCardForTest(me, effectstest.LgTower())
		effectstest.LgKnowLibrary(me)
		gyTower := effectstest.PushGraveyardCardWithTypeLineID(me, "Command Tower", "Land")
		effectstest.LgCastTowerWinder(t, g)
		return g, me, libTower, gyTower
	}

	t.Run("both are offered, marked by zone, with a floor of zero", func(t *testing.T) {
		g, me, libTower, gyTower := setup(t)
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("two matches should prompt")
		}
		effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary, game.SearchZoneGraveyard)
		if len(c.SearchCards) != 2 || c.SearchMax != 1 || c.SearchMin != 0 {
			t.Fatalf("candidates/max/min = %d/%d/%d, want 2/1/0", len(c.SearchCards), c.SearchMax, c.SearchMin)
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

	t.Run("finding nothing is allowed with a graveyard match present, and shuffles when the library was searched", func(t *testing.T) {
		g, me, libTower, gyTower := setup(t)
		effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
		effectstest.AnswerSearchFailToFind(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Contains(libTower) || me.Hand.Contains(gyTower) {
			t.Fatal("nothing should have been taken")
		}
		if !me.Graveyard.Contains(gyTower) {
			t.Error("the graveyard Tower should be untouched")
		}
		if effectstest.SearchEventCount(g) != 1 || effectstest.LibraryKnownTo(me, me.ID) != 0 {
			t.Errorf("events=%d known=%d: a library search that found nothing shuffles",
				effectstest.SearchEventCount(g), effectstest.LibraryKnownTo(me, me.ID))
		}
	})

	t.Run("finding nothing in the graveyard alone does not shuffle", func(t *testing.T) {
		g, me, libTower, gyTower := setup(t)
		effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
		effectstest.AnswerSearchFailToFind(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Contains(libTower) || me.Hand.Contains(gyTower) {
			t.Fatal("nothing should have been taken")
		}
		if effectstest.SearchEventCount(g) != 0 || effectstest.LibraryKnownTo(me, me.ID) != 2 {
			t.Errorf("events=%d known=%d: \"if you search your library this way\"",
				effectstest.SearchEventCount(g), effectstest.LibraryKnownTo(me, me.ID))
		}
	})

	t.Run("taking the graveyard Tower after searching both zones shuffles", func(t *testing.T) {
		g, me, _, gyTower := setup(t)
		effectstest.AnswerSearchByID(t, g, me.ID, gyTower)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(gyTower) {
			t.Fatal("graveyard Tower not in hand")
		}
		if effectstest.SearchEventCount(g) != 1 || effectstest.LibraryKnownTo(me, me.ID) != 0 {
			t.Errorf("events=%d known=%d: the library was searched, so it shuffles",
				effectstest.SearchEventCount(g), effectstest.LibraryKnownTo(me, me.ID))
		}
	})

	t.Run("taking the library Tower shuffles", func(t *testing.T) {
		g, me, libTower, _ := setup(t)
		effectstest.AnswerSearchByID(t, g, me.ID, libTower)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(libTower) {
			t.Fatal("library Tower not in hand")
		}
		if effectstest.SearchEventCount(g) != 1 || effectstest.LibraryKnownTo(me, me.ID) != 0 {
			t.Errorf("events=%d known=%d: a library take must announce the search and shuffle",
				effectstest.SearchEventCount(g), effectstest.LibraryKnownTo(me, me.ID))
		}
	})

	t.Run("taking the graveyard Tower from the graveyard alone does not", func(t *testing.T) {
		g, me, _, gyTower := setup(t)
		known := effectstest.LibraryKnownTo(me, me.ID)
		effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
		effectstest.AnswerSearchByID(t, g, me.ID, gyTower)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(gyTower) {
			t.Fatal("graveyard Tower not in hand")
		}
		if effectstest.SearchEventCount(g) != 0 {
			t.Errorf("EventSearchLibrary fired %d times: the library was not searched", effectstest.SearchEventCount(g))
		}
		if effectstest.LibraryKnownTo(me, me.ID) != known {
			t.Error("the library was shuffled although the take came from the graveyard")
		}
	})
}

// A candidate is checked against the zone it was offered from: a card
// that was offered from the library cannot be claimed by turning up in
// the graveyard mid-prompt.
func TestTowerWinderLibraryCandidateMilledMidPromptIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	libTower := effectstest.PushLibraryCardForTest(me, effectstest.LgTower())
	effectstest.PushGraveyardCardWithTypeLineID(me, "Command Tower", "Land")
	effectstest.LgCastTowerWinder(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("expected a prompt")
	}
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary, game.SearchZoneGraveyard)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := effectstest.PushVanillaCreature(g, me.ID, "Already Here", 2, 2)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	gyCreature := effectstest.PushGraveyardPermanent(me, "Buried Beast", "Creature — Beast", "{9}")
	tooBig := effectstest.PushGraveyardPermanent(me, "Buried Titan", "Creature — Giant", "{11}")

	effectstest.CastXSpell(t, g, "Finale of Devastation", "Sorcery", effectstest.LgFinaleOracle, "{X}{G}{G}", 10, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchByID(t, g, me.ID, gyCreature)

	if !g.Battlefield.Contains(gyCreature) {
		t.Error("the graveyard creature with mana value <= X should be on the battlefield")
	}
	if g.Battlefield.Contains(tooBig) || !me.Graveyard.Contains(tooBig) {
		t.Error("a mana value 11 creature is not a legal find for X=10")
	}
	if effectstest.SearchEventCount(g) != 0 || effectstest.LibraryKnownTo(me, me.ID) != 1 {
		t.Errorf("events=%d known=%d: a graveyard take must not search or shuffle the library",
			effectstest.SearchEventCount(g), effectstest.LibraryKnownTo(me, me.ID))
	}
	// "If X is 10 or more" is not conditioned on the search.
	if p := effectstest.EffectivePower(t, g, other); p != 12 {
		t.Errorf("Already Here power = %d, want 12", p)
	}
}

// With a graveyard creature AND a library creature on offer the prompt
// shows both; the searcher may take either or nothing. Taking the
// library one shuffles.
func TestFinaleOfDevastationOffersBothZones(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me)
	libCreature := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Library Beast", TypeLine: "Creature — Beast", ManaCost: "{4}"})
	effectstest.LgKnowLibrary(me)
	gyCreature := effectstest.PushGraveyardPermanent(me, "Buried Beast", "Creature — Beast", "{3}")

	effectstest.CastXSpell(t, g, "Finale of Devastation", "Sorcery", effectstest.LgFinaleOracle, "{X}{G}{G}", 5, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two creatures on offer should prompt")
	}
	if len(c.SearchCards) != 2 || c.SearchMin != 0 || len(c.SearchGraveyardCards) != 1 {
		t.Fatalf("candidates/min/graveyard = %d/%d/%d, want 2/0/1",
			len(c.SearchCards), c.SearchMin, len(c.SearchGraveyardCards))
	}
	effectstest.AnswerSearchByID(t, g, me.ID, libCreature)
	if !g.Battlefield.Contains(libCreature) || !me.Graveyard.Contains(gyCreature) {
		t.Error("only the library creature should have moved")
	}
	if effectstest.LibraryKnownTo(me, me.ID) != 0 || effectstest.SearchEventCount(g) != 1 {
		t.Error("taking the library creature must shuffle and announce the search")
	}
}

// A Finale whose only graveyard creature is too big for X is just a
// library search: failing to find stays legal.
func TestFinaleOfDevastationGraveyardTooBigLeavesFailToFindLegal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Library Beast", TypeLine: "Creature — Beast", ManaCost: "{2}"})
	effectstest.PushGraveyardPermanent(me, "Buried Titan", "Creature — Giant", "{11}")

	effectstest.CastXSpell(t, g, "Finale of Devastation", "Sorcery", effectstest.LgFinaleOracle, "{X}{G}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || c.SearchMin != 0 || len(c.SearchGraveyardCards) != 0 {
		t.Fatalf("want a prompt with no graveyard half and a floor of 0, got %+v", c)
	}
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if effectstest.LibraryKnownTo(me, me.ID) != 0 {
		t.Error("failing to find in the library still shuffles it")
	}
}

// Ajani's Aid: "you may search" — declining neither searches nor
// shuffles; saying yes with Ajani in the graveyard finds it.
func TestAjanisAidFindsAjaniInTheGraveyard(t *testing.T) {
	cast := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		effectstest.LgKnowLibrary(me)
		ajani := effectstest.PushGraveyardCardWithTypeLineID(me, "Ajani, Valiant Protector", "Legendary Planeswalker — Ajani")
		effectstest.CastCatalogSpell(t, g, "Ajani's Aid", "Enchantment", effectstest.LgAjanisAidOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		return g, me, ajani
	}

	t.Run("yes", func(t *testing.T) {
		g, me, ajani := cast(t)
		effectstest.AnswerMayChoice(t, g, me.ID, true)
		effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
		effectstest.AnswerSearchByID(t, g, me.ID, ajani)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(ajani) {
			t.Fatal("Ajani in the graveyard was not put into hand")
		}
		if effectstest.SearchEventCount(g) != 0 || effectstest.LibraryKnownTo(me, me.ID) != 1 {
			t.Error("a graveyard take must not shuffle the library")
		}
	})

	t.Run("no", func(t *testing.T) {
		g, me, ajani := cast(t)
		effectstest.AnswerMayChoice(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Contains(ajani) || !me.Graveyard.Contains(ajani) {
			t.Error("declining must leave Ajani in the graveyard")
		}
		if effectstest.SearchEventCount(g) != 0 || effectstest.LibraryKnownTo(me, me.ID) != 1 {
			t.Error("declining neither searches nor shuffles")
		}
	})
}

// The five cards that shipped with the seam, one case each. The four
// "named" creatures share one builder; Delivery Moogle is the only
// compulsory one.
func TestDeliveryMoogleTakesAnArtifactFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	cheap := effectstest.PushGraveyardCardWithTypeLineID(me, "Cheap Rock", "Artifact")
	// pushGraveyardCardWithTypeLineID leaves the mana cost empty (value 0);
	// a dear artifact needs one.
	dear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear Rock", TypeLine: "Artifact", ManaCost: "{5}", Owner: me.ID, Controller: me.ID})

	effectstest.CastCatalogSpell(t, g, "Delivery Moogle", "Creature — Moogle", effectstest.LgMoogleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != cheap {
		t.Fatalf("only the cheap artifact is a candidate, got %+v", c)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Hand.Contains(cheap) {
		t.Error("the artifact card with mana value <= 2 should be in hand")
	}
	if me.Hand.Contains(dear) {
		t.Error("a mana value 5 artifact is not a legal find")
	}
	if effectstest.SearchEventCount(g) != 0 {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			effectstest.LgKnowLibrary(me)
			pw := effectstest.PushGraveyardCardWithTypeLineID(me, tc.fetches, "Legendary Planeswalker")
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
			effectstest.AnswerSearchByID(t, g, me.ID, pw)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(pw) {
				t.Fatalf("%s was not fetched from the graveyard", tc.fetches)
			}
			if effectstest.LibraryKnownTo(me, me.ID) != 1 {
				t.Error("a graveyard take must not shuffle the library")
			}
		})
		t.Run(tc.name+" library", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			pw := effectstest.PushLibraryCardForTest(me, game.Card{Name: tc.fetches, TypeLine: "Legendary Planeswalker"})
			effectstest.LgKnowLibrary(me)
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			effectstest.AnswerSearchByID(t, g, me.ID, pw)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(pw) {
				t.Fatalf("%s was not fetched from the library", tc.fetches)
			}
			if effectstest.LibraryKnownTo(me, me.ID) != 0 {
				t.Error("a library take must shuffle")
			}
		})
	}
}
