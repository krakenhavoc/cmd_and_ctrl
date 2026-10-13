package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func studentOfElementsRow() cards.Card {
	return flipRow(studentOfElementsOracleID, "{1}{U}", []string{"U"}, nil,
		cards.CardFace{Name: "Student of Elements", TypeLine: "Creature — Human Wizard",
			OracleText: "When this creature has flying, flip it.",
			Power:      "1", Toughness: "1"},
		cards.CardFace{Name: "Tobita, Master of Winds", TypeLine: "Legendary Creature — Human Wizard",
			OracleText: "Creatures you control have flying.",
			Power:      "3", Toughness: "3"})
}

// Student of Elements flips the moment it has flying, whatever gives it
// flying; Tobita then gives every creature its controller has flying,
// and only that controller's.
func TestStudentOfElementsFlipsWhenItHasFlying(t *testing.T) {
	g, me, opp := flipMain(t)
	student := flipEntered(t, g, studentOfElementsRow(), me)
	theirStudent := flipEntered(t, g, studentOfElementsRow(), opp)
	bear := apaPush(g, me.ID, me.ID, apaCreature("My Bear", "oracle-my-bear", 2, 2))
	passPriorityAroundTable(t, g)
	wantUnflipped(t, g, student, "Student of Elements")

	// A second Student, flipped by hand, is a Tobita: it gives MY
	// creatures flying, which triggers the first Student's state.
	helper := flipEntered(t, g, studentOfElementsRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(helper); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	wantFlipped(t, g, student, "Tobita, Master of Winds", []string{"U"}, 2)
	wantUnflipped(t, g, theirStudent, "Student of Elements")
	b := flipState(t, g, bear)
	if !game.HasKeyword(&b, "flying") {
		t.Error("Tobita does not give My Bear flying")
	}
	their := flipState(t, g, theirStudent)
	if game.HasKeyword(&their, "flying") {
		t.Error("Tobita gave the opponent's Student flying")
	}
}
