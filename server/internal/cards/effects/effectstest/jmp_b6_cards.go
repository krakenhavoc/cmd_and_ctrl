package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Skittering Surveyor and Sylvan Ranger ---

func JmpB6FetchesABasicToHand(t *testing.T, name, typeLine, oracle string, power, toughness int) {
	g := NewCatalogGame(t)
	me := g.Seats[0]
	SeedSearchLibrary(me,
		SearchTestLand("Forest", "Basic Land — Forest"),
		SearchTestLand("Temple Garden", "Land — Forest Plains"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	CastHoldCreature(t, g, name, typeLine, oracle, power, toughness)
	c := SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, bad := range []string{"Temple Garden", "Bear"} {
		if SearchOptionNamed(g, c, bad) != uuid.Nil {
			t.Errorf("the search offered %s, which is not a basic land card", bad)
		}
	}
	handBefore := me.Hand.Size()
	AnswerSearchNamed(t, g, me.ID, "Forest")
	PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d -> %d, want +1", handBefore, me.Hand.Size())
	}
	if me.Library.Size() != 2 {
		t.Errorf("library has %d cards, want 2", me.Library.Size())
	}
}

func JmpB6DeclinesTheSearch(t *testing.T, name, typeLine, oracle string, power, toughness int) {
	g := NewCatalogGame(t)
	me := g.Seats[0]
	SeedSearchLibrary(me,
		SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	before := LibraryIDs(me)
	CastHoldCreature(t, g, name, typeLine, oracle, power, toughness)
	c := SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := AnswerSearchApply(g, c, me.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	PassPriorityAroundTable(t, g)
	if !SameIDList(before, LibraryIDs(me)) {
		t.Error("declining the search still moved or shuffled the library")
	}
}
