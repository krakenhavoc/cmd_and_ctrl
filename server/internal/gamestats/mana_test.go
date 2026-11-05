package gamestats

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mana_test.go is ADR 0217 §4's cost reading and colour matching.

func TestParseCost(t *testing.T) {
	for in, want := range map[string]struct {
		mv   int
		pips int
	}{
		"":                    {0, 0},
		"{2}{G}{G}":           {4, 2},
		"{X}{R}":              {1, 1},
		"{W/U}{W/U}":          {2, 2},
		"{G/P}{1}":            {2, 0}, // Phyrexian: life can pay
		"{2/W}{2/W}":          {4, 0}, // two generic can pay each
		"{C}{C}":              {2, 2},
		"{1}{U} // {3}{B}{B}": {2, 1}, // the face you cast from the hand
		"{10}":                {10, 0},
		"{S}{G}":              {2, 1},
	} {
		c := parseCost(in)
		if c.mv != want.mv || len(c.pips) != want.pips {
			t.Errorf("parseCost(%q) = mv %d, %d pips; want %d, %d", in, c.mv, len(c.pips), want.mv, want.pips)
		}
	}
}

func TestPayableNeedsAMatchingNotAGreedyPick(t *testing.T) {
	// {W}{U} against a Plains and a Tundra: a greedy pick that spends the
	// Tundra on {W} strands {U}; the matching puts the Plains on {W}.
	pips := parseCost("{W}{U}").pips
	if !payable(pips, []colours{colW | colU, colW}) {
		t.Error("{W}{U} with a Tundra and a Plains is payable")
	}
	if payable(pips, []colours{colW, colW}) {
		t.Error("{W}{U} with two Plains is not payable")
	}
	if !payable(parseCost("{W/U}{B}").pips, []colours{colU, colB}) {
		t.Error("a hybrid symbol takes either colour")
	}
	if payable(parseCost("{C}").pips, []colours{colG}) {
		t.Error("{C} needs a colourless source")
	}
}

func TestCastableByAmountAndColour(t *testing.T) {
	forest := []colours{colG, colG, colG, colG}
	for _, c := range []struct {
		cost             string
		amount, byColour bool
	}{
		{"{2}{G}{G}", true, true},
		{"{3}{U}", true, false}, // colour screw: enough mana, wrong colour
		{"{4}{G}", false, false},
		{"{X}{G}", true, true},
		{"{B/P}{B/P}{2}", true, true},
	} {
		a, b := castable(c.cost, forest)
		if a != c.amount || b != c.byColour {
			t.Errorf("castable(%q) = %v, %v; want %v, %v", c.cost, a, b, c.amount, c.byColour)
		}
	}
}

type fakeIndex map[uuid.UUID]cards.Card

func (f fakeIndex) Get(id uuid.UUID) (cards.Card, bool) { c, ok := f[id]; return c, ok }

func TestProducedByPrefersTheIndex(t *testing.T) {
	id := uuid.New()
	idx := fakeIndex{id: {ID: id, ProducedMana: []string{"W", "U"}}}
	tundra := protocol.CardView{ScryfallID: id.String(), ManaAbilities: []protocol.ManaAbilityView{{Produced: "{G}"}}}
	if got := producedBy(&tundra, idx); got != colW|colU {
		t.Errorf("from the index: %b, want W and U", got)
	}
	if got := producedBy(&tundra, nil); got != colG {
		t.Errorf("from the view: %b, want G", got)
	}
	choice := protocol.CardView{ManaAbilities: []protocol.ManaAbilityView{{ColorOptions: [][]string{{"R"}, {"G"}}}}}
	if got := producedBy(&choice, nil); got != colR|colG {
		t.Errorf("a choice of colours: %b, want R and G", got)
	}
	if got := producedBy(&protocol.CardView{}, nil); got != 0 {
		t.Errorf("no mana ability: %b, want none", got)
	}
}

func TestLandKind(t *testing.T) {
	for _, c := range []struct {
		card protocol.CardView
		want landKind
	}{
		{protocol.CardView{TypeLine: "Basic Land — Forest"}, landCard},
		{protocol.CardView{TypeLine: "Artifact Land"}, landCard},
		{protocol.CardView{TypeLine: "Creature — Bear"}, spellCard},
		{protocol.CardView{TypeLine: "Sorcery // Land"}, landBackCard},
		{protocol.CardView{TypeLine: "Instant", Faces: []protocol.CardFaceView{{TypeLine: "Instant"}, {TypeLine: "Land"}}}, landBackCard},
	} {
		if got := landKindOf(&c.card); got != c.want {
			t.Errorf("%q: %d, want %d", c.card.TypeLine, got, c.want)
		}
	}
}
