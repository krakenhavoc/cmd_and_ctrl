package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_multicolored_test.go — "protection from multicolored"
// (Stonecoil Serpent, #4834): a source with two or more colours (CR 105.2b)
// is stopped by it, a monocoloured or colourless one is not.

func TestProtectionFromMulticoloredGrammar(t *testing.T) {
	for _, token := range []string{"protection from multicolored", "Protection from Multicolored"} {
		q, ok := ParseProtectionQuality(token)
		if !ok || q.Kind != ProtectionQualityMulticolored || q.Kind.String() != "multicolored" {
			t.Errorf("%q = (%+v, %v)", token, q, ok)
		}
		if _, ok := CanonicalKeywords(token); !ok {
			t.Errorf("CanonicalKeywords(%q) refused it", token)
		}
	}
	if toks, ok := ProtectionTokens("protection from multicolored"); !ok || len(toks) != 1 {
		t.Errorf("ProtectionTokens = %v, %v", toks, ok)
	}
}

func TestProtectionFromMulticoloredStopsOnlyTwoOrMoreColours(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cost  string
		legal bool
	}{
		{"colourless", "{3}", true},
		{"one colour", "{1}{G}", true},
		{"two colours", "{G}{W}", false},
		{"three colours", "{G}{W}{U}", false},
		{"hybrid of two colours", "{G/W}", false},
	} {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		attacker := pushMVCreature(g, me, "Shielded", "{7}", "protection from multicolored")
		blocker := pushMVCreature(g, opp, tc.name, tc.cost)
		card := func(id uuid.UUID) *Card { return &g.Battlefield.Cards[findCardOnBattlefield(g, id)] }
		var r BlockRefusal
		g.ReadSnapshot(func() { r = g.BlockPairRefusalLocked(card(attacker), card(blocker)) })
		if tc.legal && !r.Legal() {
			t.Errorf("%s: refused (%q), want legal", tc.name, r.Reason)
		}
		if !tc.legal && r.Reason != BlockReasonProtection {
			t.Errorf("%s: reason %q, want protection", tc.name, r.Reason)
		}
	}
}
