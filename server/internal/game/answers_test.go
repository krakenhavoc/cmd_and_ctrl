package game

import (
	"slices"
	"testing"
)

// ADR 0142 decision 1: ten values, in the table's order, in two tiers.
func TestAnswersVocabulary(t *testing.T) {
	want := []string{"protect", "pump", "prevent", "remove", "sac_outlet", "restrict", "combat_grant", "animate", "makes_blocker", "value"}
	if got := AnswerWireNames(); !slices.Equal(got, want) {
		t.Fatalf("AnswerWireNames() = %v, want %v", got, want)
	}
	all := answersKnown
	if got := all.Wire(); !slices.Equal(got, want) {
		t.Fatalf("every bit's wire = %v, want %v", got, want)
	}
	for _, c := range []struct {
		a             Answers
		stack, combat bool
	}{
		{AnswerProtect | AnswerPump | AnswerPrevent | AnswerRemove | AnswerSacOutlet | AnswerRestrict, true, false},
		{AnswerCombatGrant | AnswerAnimate | AnswerMakesBlocker, false, true},
		{AnswerProtect | AnswerCombatGrant, true, true},
		{AnswerValue, false, false},
		{0, false, false},
	} {
		if got := c.a.HasTier(TierStack); got != c.stack {
			t.Errorf("%s: stack tier = %v, want %v", c.a, got, c.stack)
		}
		if got := c.a.HasTier(TierCombat); got != c.combat {
			t.Errorf("%s: combat tier = %v, want %v", c.a, got, c.combat)
		}
	}
	if u := (AnswerValue << 1).Unknown(); u == 0 {
		t.Error("a bit past the vocabulary is not reported unknown")
	}
	if Answers(0).Wire() != nil {
		t.Error("the empty set has wire names")
	}
}
