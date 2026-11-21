package testsst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsst "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/st"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func studentOfElementsRow() cards.Card {
	return effectstest.FlipRow(cardsst.StudentOfElementsOracleID, "{1}{U}", []string{"U"}, nil,
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
	g, me, opp := effectstest.FlipMain(t)
	student := effectstest.FlipEntered(t, g, studentOfElementsRow(), me)
	theirStudent := effectstest.FlipEntered(t, g, studentOfElementsRow(), opp)
	bear := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("My Bear", "oracle-my-bear", 2, 2))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantUnflipped(t, g, student, "Student of Elements")

	// A second Student, flipped by hand, is a Tobita: it gives MY
	// creatures flying, which triggers the first Student's state.
	helper := effectstest.FlipEntered(t, g, studentOfElementsRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(helper); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantFlipped(t, g, student, "Tobita, Master of Winds", []string{"U"}, 2)
	effectstest.WantUnflipped(t, g, theirStudent, "Student of Elements")
	b := effectstest.FlipState(t, g, bear)
	if !game.HasKeyword(&b, "flying") {
		t.Error("Tobita does not give My Bear flying")
	}
	their := effectstest.FlipState(t, g, theirStudent)
	if game.HasKeyword(&their, "flying") {
		t.Error("Tobita gave the opponent's Student flying")
	}
}
