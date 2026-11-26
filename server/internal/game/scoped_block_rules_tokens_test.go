package game

import "testing"

// scoped_block_rules_tokens_test.go pins the cantBeBlockedByTokens kind
// (Legion Loyalist, CR 509.1b): the registration check, and that the layer
// adapter leaves it alone as it does every block-rule kind. The card's own
// behaviour is pinned in cards/effects.

func TestCantBeBlockedByTokensModNeedsText(t *testing.T) {
	if blockRuleModProblem(Mod{Kind: ModCantBeBlockedByTokens}) == "" {
		t.Error("a cantBeBlockedByTokens rule with no Text passed the registration check")
	}
	if p := blockRuleModProblem(CantBeBlockedByTokensMod("creature tokens")); p != "" {
		t.Errorf("unexpected problem %q", p)
	}
}
