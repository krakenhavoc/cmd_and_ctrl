package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes_not_chosen_view_test.go — ADR 0097 Decision 5: the table sees
// the modes an ability has already chosen, greyed out, and the view
// reads the same memory the gate refuses from.

// The mode_pick prompt carries the used bullets beside the offered
// ones, with the restriction that put them there.
func TestModePickProjectsTheUsedModes(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:            game.PendingChoiceModePick,
			Chooser:         me.ID,
			Count:           1,
			Reason:          "Choose one that hasn't been chosen this turn",
			ModeOptionIndex: []int{0, 2},
			ModeOptionLabel: []string{"Put a +1/+1 counter on this creature.", "You gain 2 life."},
			ModeUsedIndex:   []int{1},
			ModeUsedLabel:   []string{"Create a tapped Treasure token."},
			ModeNotChosen:   game.ModeMemoryThisTurn,
			ModeMin:         1,
			ModeMax:         1,
		})
	})
	v := ViewOfGameFor(g, me.ID.String())
	var pc *PendingChoiceView
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Kind == string(game.PendingChoiceModePick) {
			pc = &v.PendingChoices[i]
		}
	}
	if pc == nil {
		t.Fatal("the prompt reaches its chooser")
	}
	if !slices.Equal(pc.ModeIndexes, []int{0, 2}) {
		t.Errorf("mode_indexes stays the offer list: %v", pc.ModeIndexes)
	}
	if !slices.Equal(pc.ModeUsedIndexes, []int{1}) || pc.ModeUsedOptions[0] != "Create a tapped Treasure token." {
		t.Errorf("the used bullet rides beside it: %v %v", pc.ModeUsedIndexes, pc.ModeUsedOptions)
	}
	if pc.ModeNotChosen != "this_turn" {
		t.Errorf("mode_not_chosen = %q, want this_turn", pc.ModeNotChosen)
	}
}

// An activated ability's mode view marks the used bullets from the
// memory the activation gate reads, and a spell's (no identity) never
// does.
func TestModeSpecViewMarksUsedOptions(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	src := game.Card{InstanceID: uuid.New(), Name: "Fixture", Owner: me.ID, Controller: me.ID}
	ms := &game.ModeSpec{
		Prompt: "Choose one that hasn't been chosen this turn", Min: 1, Max: 1,
		Options:   []game.ModeOption{{Label: "A."}, {Label: "B."}, {Label: "C."}},
		NotChosen: game.ModeMemoryThisTurn,
	}
	ab := game.ModeAbilityOf(src, "Fixture — choose")
	g.WithWriteLock(func() {
		g.TurnTally.ModesChosen = map[string][]int{
			game.ObjectTallyKey(ab.Source, ab.Epoch, ab.Label): {2},
		}
	})
	var withID, spell *ModeSpecView
	g.ReadSnapshot(func() {
		q := game.ModeCountQuery{Chooser: me.ID}
		withID = viewOfModeSpec(g, q, game.SourceObject(me.ID, &src), ms, ab)
		spell = viewOfModeSpec(g, q, game.SourceObject(me.ID, &src), ms, game.ModeAbility{})
	})
	if withID.NotChosen != "this_turn" {
		t.Errorf("not_chosen = %q", withID.NotChosen)
	}
	for i, o := range withID.Options {
		if o.Used != (i == 2) {
			t.Errorf("option %d used=%v", i, o.Used)
		}
	}
	for i, o := range spell.Options {
		if o.Used {
			t.Errorf("a choice with no ability identity marks nothing used: option %d", i)
		}
	}
}
