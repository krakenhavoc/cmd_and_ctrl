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

// ADR 0188 Decision 12: the prompt prints "(recommended)" after the
// recommended answer, and a model that copies it along still names it
// (the label match already tolerates a trailing note; this pins it).
// The index is deliberately wrong, so only the label can name the move.
func TestNormLabelDropsTheRecommendedNote(t *testing.T) {
	shown := []Choice{{0, "which skip takes your extra turn: Use my skip"}, {1, "which skip takes your extra turn: Let Trouble in Pairs skip it"}}
	a := ResolveAnswer(`{"index": 7, "move": "which skip takes your extra turn: Use my skip (recommended)"}`, shown, 2)
	if a.Index != 0 {
		t.Errorf("got index %d pick %q, want 0", a.Index, a.Pick)
	}
}
