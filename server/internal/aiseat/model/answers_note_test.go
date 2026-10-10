package model

import "testing"

// ADR 0142 decision 6: the prompt prints "(answers: …)" after a label,
// and a model that copies it along still names the move.
func TestNormLabelDropsTheAnswersNote(t *testing.T) {
	shown := []Choice{{0, "Pass priority"}, {1, "Albino Troll: Regenerate"}}
	a := ResolveAnswer(`{"index": 1, "move": "Albino Troll: Regenerate (answers: protect)"}`, shown, 2)
	if a.Index != 1 || a.Pick != PickIndex {
		t.Errorf("got index %d pick %q, want 1 %q", a.Index, a.Pick, PickIndex)
	}
}
