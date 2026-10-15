package boardtext_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0188 Decision 12: the which-skip question's "use my skip" is the
// recommended answer, and no other move is marked.
func TestRecommendedNoteMarksUseMySkip(t *testing.T) {
	v := &protocol.GameView{PendingChoices: []protocol.PendingChoiceView{
		{ID: "q", Kind: "turn_skip_order", TurnSkipOrder: &protocol.TurnSkipOrderView{}},
		{ID: "riot", Kind: "entry_riot"},
	}}
	move := func(id string, apply bool) legal.Move {
		b, _ := json.Marshal(map[string]any{"choice_id": id, "apply": apply})
		return legal.Move{Kind: legal.KindChoice, Params: b}
	}
	if got := boardtext.RecommendedNote(v, move("q", true)); got != boardtext.RecommendedText {
		t.Errorf("use my skip: got %q", got)
	}
	if got := boardtext.RecommendedNote(v, move("q", false)); got != "" {
		t.Errorf("let it skip: got %q", got)
	}
	if got := boardtext.RecommendedNote(v, move("riot", true)); got != "" {
		t.Errorf("another prompt: got %q", got)
	}
	if got := boardtext.RecommendedNote(v, legal.Move{Kind: legal.KindPass}); got != "" {
		t.Errorf("a pass: got %q", got)
	}
}

// ADR 0142 decision 6: a move line names what its row declares it
// answers, read from the view.
func TestAnswersNoteReadsTheRow(t *testing.T) {
	a := []string{"protect", "pump"}
	v := &protocol.GameView{}
	v.Battlefield.Cards = []protocol.CardView{{InstanceID: "troll", Name: "Albino Troll",
		ActivatedAbilities: []protocol.ActivatedAbilityView{{Index: 0}, {Index: 1, Purpose: &protocol.PurposeView{Answers: &a}}}}}
	move := func(i int) legal.Move {
		b, _ := json.Marshal(map[string]any{"source_card_id": "troll", "ability_index": i})
		return legal.Move{Kind: legal.KindActivate, Params: b}
	}
	if got := boardtext.AnswersNote(v, move(1)); got != "answers: protect, pump" {
		t.Errorf("row 1: got %q", got)
	}
	if got := boardtext.AnswersNote(v, move(0)); got != "" {
		t.Errorf("row 0 declares nothing: got %q", got)
	}
	if got := boardtext.AnswersNote(v, legal.Move{Kind: legal.KindPass}); got != "" {
		t.Errorf("a pass: got %q", got)
	}
}
