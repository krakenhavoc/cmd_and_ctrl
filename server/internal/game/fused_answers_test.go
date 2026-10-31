package game

import "testing"

// fused_answers_test.go — #2877: a fused split spell answers what both
// halves declare, and a half that declares nothing leaves the fused
// spell undeclared, never `value`, so a reader stops for it (ADR 0142
// owner answer 3).
func TestFusedAnswersNeedBothHalves(t *testing.T) {
	for _, c := range []struct {
		name        string
		left, right Answers
		want        Answers
	}{
		{"both value", AnswerValue, AnswerValue, AnswerValue},
		{"value and an answer", AnswerValue, AnswerRemove, AnswerValue | AnswerRemove},
		{"one half undeclared", AnswerValue, 0, 0},
		{"neither declared", 0, 0, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := Purpose{Answers: c.left}.plus(Purpose{Answers: c.right}, 0).Answers
			if got != c.want {
				t.Fatalf("fused answers = %s, want %s", got, c.want)
			}
		})
	}
}
