package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3943 — Moonsilver Key: "Search your library for an artifact card
// with a mana ability or a basic land card". Whether an artifact has a
// mana ability is what it prints (CR 605.1a/b), stamped at import, so
// these library cards carry the stamp the importer would give them.

const millikinOracle = "fa4dffda-6f04-4d0b-829d-28a1a5794dee"

// moonsilverLibraryCard is a library card as the importer would stamp
// it from its printed text.
func moonsilverLibraryCard(name, typeLine, text, oracle string, produced ...string) game.Card {
	return game.Card{
		Name: name, TypeLine: typeLine, OracleID: oracle, ProducedMana: produced,
		ManaAbilityText: game.PrintedFactOf(game.ManaAbilityInText(typeLine, text)),
	}
}

func TestMoonsilverKeyFindsWhatThePrintedTextCallsAManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	key := effectstest.B12Push(g, me.ID, "Moonsilver Key", "Artifact", effectstest.B15MoonsilverKeyOracle, 0, 0)
	effectstest.SeedSearchLibrary(me,
		// Not in the catalog: found on its printed text.
		moonsilverLibraryCard("Worn Powerstone", "Artifact", "This artifact enters tapped.\n{T}: Add {C}{C}.", "", "C"),
		// A triggered mana ability (CR 605.1b).
		moonsilverLibraryCard("Gauntlet of Power", "Artifact", "As this artifact enters, choose a color.\nCreatures of the chosen color get +1/+1.\nWhenever a basic land is tapped for mana of the chosen color, its controller adds an additional one mana of that color.", "", "B", "G", "R", "U", "W"),
		// Scryfall says each of these makes mana; none has a mana ability.
		moonsilverLibraryCard("Su-Chi", "Artifact Creature — Construct", "When this creature dies, add {C}{C}{C}{C}.", "", "C"),
		moonsilverLibraryCard("Chromatic Sphere", "Artifact", "{1}, {T}, Sacrifice this artifact: Add one mana of any color. Draw a card. (Activate only as an instant.)", "", "B", "G", "R", "U", "W"),
		moonsilverLibraryCard("Paradise Mantle", "Artifact — Equipment", "Equipped creature has \"{T}: Add one mana of any color.\"\nEquip {1}", "", "B", "G", "R", "U", "W"),
		// In the catalog, which runs its mill-cost ability as a mana
		// ability; the printed text decides, and CR 605.1a says no.
		moonsilverLibraryCard("Millikin", "Artifact Creature — Construct", "{T}, Mill a card: Add {C}. (Activate only as an instant. To mill a card, put the top card of your library into your graveyard.)", millikinOracle, "C"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
	)
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, key, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("several matches: the searcher chooses")
	}
	for _, name := range []string{"Worn Powerstone", "Gauntlet of Power", "Forest"} {
		if effectstest.SearchOptionNamed(g, c, name) == uuid.Nil {
			t.Errorf("%s is offered", name)
		}
	}
	for _, name := range []string{"Su-Chi", "Chromatic Sphere", "Paradise Mantle", "Millikin"} {
		if effectstest.SearchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s is offered, but has no mana ability of its own (CR 605.1)", name)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Worn Powerstone")
	if !effectstest.B02bHandHasNamed(me, "Worn Powerstone") {
		t.Error("the uncatalogued rock goes to hand")
	}
}

func TestMoonsilverKeyIsFull(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.B15MoonsilverKeyOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Moonsilver Key: %v, caveats %v; want Full with none", spec.Completeness, spec.Caveats)
	}
}
