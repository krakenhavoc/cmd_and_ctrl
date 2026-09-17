package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// color_choice_test.go — #742's choose_color kind has answers (#499:
// a kind with no enumerator case is a seat with an empty move list),
// and the shared `color` payload reaches the right resolver for both
// colour prompts.

func TestChooseColorOffersEveryOptionBoardColoursFirst(t *testing.T) {
	g := newTable(t)
	me, them := g.Seats[0], g.Seats[1]
	battlefieldCard(g, me, creature("Llanowar Elves", "{G}", 1, 1))
	battlefieldCard(g, me, creature("Elvish Mystic", "{G}", 1, 1))
	battlefieldCard(g, me, creature("Goblin Guide", "{R}", 2, 2))
	// An opponent's colours are not this seat's preference.
	battlefieldCard(g, them, creature("Savannah Lions", "{W}", 2, 1))
	battlefieldCard(g, them, creature("Savannah Lions", "{W}", 2, 1))
	battlefieldCard(g, them, creature("Savannah Lions", "{W}", 2, 1))
	g.WithWriteLock(func() {
		g.QueueColorChoiceForEffect(me.ID, uuid.New(), "Thriving Isle — choose a color other than blue", game.ColorsOtherThan("U"))
	})

	moves := legal.EnumerateFor(g, me.ID)
	got := make([]string, 0, len(moves))
	for _, m := range moves {
		got = append(got, m.Label[strings.LastIndex(m.Label, ": ")+2:])
	}
	want := []string{"green", "red", "white", "black"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("answers = %v, want %v (the seat's own colours first, blue never)", got, want)
	}
	if !allAlwaysLegal(moves) {
		t.Error("the resolver accepts every listed colour, so every answer is always legal")
	}
	dispatchAll(t, g, me.ID, moves)
}

func TestChooseColorResolutionFormIsAnswerable(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueColorChoiceThenForEffect(game.ColorPrompt{
			Chooser: me.ID, Question: "Wash Out — choose a color",
			Then: func(*game.Game, string) error { return nil },
		})
	})
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 5 {
		t.Fatalf("enumerated %d answers, want the five colours: %v", len(moves), labels(moves))
	}
	dispatchAll(t, g, me.ID, moves)
}

// The mana pick and the colour prompt share `color`; routing by kind
// is what keeps each answer on its own resolver.
func TestColorPayloadRoutesByKind(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	src := battlefieldCard(g, me, game.Card{Name: "Coldsteel Heart", TypeLine: "Artifact"})
	var colorChoice, manaChoice uuid.UUID
	g.WithWriteLock(func() {
		colorChoice = g.QueueColorChoiceForEffect(me.ID, src, "Coldsteel Heart", nil)
		manaChoice = g.QueueChoiceForEffect(game.PendingChoice{
			Kind: game.PendingChoiceMana, Chooser: me.ID, FromPlayer: me.ID, Count: 1,
			Reason: "Gilded Lotus", ColorOptions: []string{"W", "U", "B", "R", "G"},
			ManaAmounts: map[string]int{"W": 3, "U": 3, "B": 3, "R": 3, "G": 3},
		})
	})

	moves := legal.EnumerateFor(g, me.ID)
	if !hasLabel(moves, "Gilded Lotus: add {G}{G}{G}") {
		t.Errorf("the mana pick's label does not name all three tokens: %v", labels(moves))
	}

	answer := func(id uuid.UUID, color string) {
		t.Helper()
		err := actions.Dispatch(g, actions.Action{
			Type: actions.TypeResolveChoice, Player: me.ID, Caller: me.ID,
			Params: []byte(`{"choice_id":"` + id.String() + `","color":"` + color + `"}`),
		})
		if err != nil {
			t.Fatalf("resolve %s with %s: %v", id, color, err)
		}
	}
	answer(colorChoice, "U")
	answer(manaChoice, "G")

	var chosen string
	g.ReadSnapshot(func() { chosen = g.ChosenColorOf(src) })
	if chosen != "U" {
		t.Errorf("ChosenColor = %q, want U", chosen)
	}
	if n := len(me.ManaPool); n != 3 {
		t.Errorf("pool = %v, want three {G}", me.ManaPool)
	}
}
