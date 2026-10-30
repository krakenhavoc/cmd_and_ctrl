package legal

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_answers_internal_test.go — #2877: castAnswersOf reads a cast's
// declaration from the card, the chosen modes and the alternative cost,
// and an undeclared part leaves the whole undeclared.
func TestCastAnswersOf(t *testing.T) {
	value, remove := game.AnswerValue, game.AnswerRemove
	modes := &game.ModeSpec{Options: []game.ModeOption{
		{Purpose: game.Purpose{Answers: remove}},
		{},
		{Purpose: game.Purpose{Answers: value}},
	}}
	overload := &game.AlternativeCost{Key: "overload", ClearsTargets: true}
	flashback := &game.AlternativeCost{Key: "flashback"}
	declared := &game.AlternativeCost{Key: "overload", ClearsTargets: true, Purpose: game.Purpose{Answers: remove}}
	for _, c := range []struct {
		name  string
		own   game.Answers
		modes []int
		offer *game.AlternativeCost
		want  game.Answers
	}{
		{"the card's own", value, nil, nil, value},
		{"undeclared", 0, nil, nil, 0},
		{"a declared mode", 0, []int{2, 2}, nil, value},
		{"a declared mode beside an undeclared one", 0, []int{2, 1}, nil, 0},
		{"the card covers an undeclared mode", value, []int{1, 2}, nil, value},
		{"a mode's own answer replaces the card's", value, []int{0}, nil, remove},
		{"modes of two kinds", 0, []int{0, 2}, nil, remove | value},
		{"a cost that leaves the effect alone", value, nil, flashback, value},
		{"a cost that changes the spell and declares nothing", value, nil, overload, 0},
		{"a cost that declares its own", value, nil, declared, remove},
	} {
		t.Run(c.name, func(t *testing.T) {
			ms := modes
			if c.modes == nil {
				ms = nil
			}
			if got := castAnswersOf(c.own, ms, c.modes, c.offer); got != c.want {
				t.Fatalf("castAnswersOf = %s, want %s", got, c.want)
			}
		})
	}
}

// castValueOnly: exactly `value`, nothing on the stack targeted, and no
// creature given up as a cost.
func TestCastValueOnly(t *testing.T) {
	g := game.NewGame()
	bear := game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear"}
	land := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"}
	g.Battlefield.PushTop(bear)
	g.Battlefield.PushTop(land)
	for _, c := range []struct {
		name    string
		answers game.Answers
		stack   bool
		ids     []uuid.UUID
		want    bool
	}{
		{"value", game.AnswerValue, false, nil, true},
		{"value giving up a land", game.AnswerValue, false, []uuid.UUID{land.InstanceID}, true},
		{"value giving up a creature", game.AnswerValue, false, []uuid.UUID{bear.InstanceID}, false},
		{"value aimed at the stack", game.AnswerValue, true, nil, false},
		{"undeclared", 0, false, nil, false},
		{"a sacrifice outlet", game.AnswerSacOutlet, false, nil, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := castValueOnly(g, c.answers, c.stack, c.ids); got != c.want {
				t.Fatalf("castValueOnly = %v, want %v", got, c.want)
			}
		})
	}
}

// The printed costs the guard reads: a creature sacrifice or exile is an
// outlet, a land is a price, and an either/or cost answers what any
// branch does.
func TestSpellCostAnswers(t *testing.T) {
	creature := &game.TargetSpec{Label: "a creature"}
	land := &game.TargetSpec{Label: "a land"}
	if a := AdditionalCostAnswers(&game.AdditionalCost{Sacrifice: creature}); a != game.AnswerSacOutlet {
		t.Errorf("sacrifice a creature: %s", a)
	}
	if a := AdditionalCostAnswers(&game.AdditionalCost{Sacrifice: land}); a != 0 {
		t.Errorf("sacrifice a land: %s", a)
	}
	if a := AdditionalCostAnswers(&game.AdditionalCost{Either: []game.AdditionalCost{{DiscardCards: 1}, {Sacrifice: creature}}}); a != game.AnswerSacOutlet {
		t.Errorf("discard or sacrifice a creature: %s", a)
	}
	if a := AdditionalCostAnswers(nil); a != 0 {
		t.Errorf("no cost: %s", a)
	}
	if a := AlternativeCostAnswers(&game.AlternativeCost{ReturnToHand: creature}); a != game.AnswerSacOutlet {
		t.Errorf("return a creature instead of paying: %s", a)
	}
	if a := AlternativeCostAnswers(&game.AlternativeCost{ReturnToHand: &game.TargetSpec{Label: "two Islands"}}); a != 0 {
		t.Errorf("return two Islands instead of paying: %s", a)
	}
}
