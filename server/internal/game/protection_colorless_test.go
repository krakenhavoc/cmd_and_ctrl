package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_colorless_test.go — "protection from colorless" (Angelic
// Intervention, #5650): a source with no colour (CR 105.2c) is stopped
// by it, a coloured one is not; and the one prompt that offers
// colorless beside the five colours.

func TestProtectionFromColorlessGrammar(t *testing.T) {
	for _, token := range []string{ProtectionFromColorless, "Protection from Colorless"} {
		q, ok := ParseProtectionQuality(token)
		if !ok || q.Kind != ProtectionQualityColorless || q.Kind.String() != "colorless" || q.Value != "" {
			t.Errorf("%q = (%+v, %v)", token, q, ok)
		}
		if _, ok := CanonicalKeywords(token); !ok {
			t.Errorf("CanonicalKeywords(%q) refused it", token)
		}
	}
	if got := ProtectionFromColorOrColorless(ColorlessChoice); got != ProtectionFromColorless {
		t.Errorf("the colourless answer = %q", got)
	}
	if got := ProtectionFromColorOrColorless("R"); got != "protection from red" {
		t.Errorf("the red answer = %q", got)
	}
	if got := ProtectionFromColorOrColorless("X"); got != "" {
		t.Errorf("junk minted %q", got)
	}
	// The five-colour door stays five colours.
	if got := ProtectionFromColor(ColorlessChoice); got != "" {
		t.Errorf("ProtectionFromColor(C) = %q, want nothing", got)
	}
	// Hexproof's grammar does not gain it: no card prints it.
	if _, ok := ParseHexproofFromQuality("hexproof from colorless"); ok {
		t.Error("hexproof from colorless parsed")
	}
}

func TestProtectionFromColorlessStopsOnlySourcesWithNoColour(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cost  string
		legal bool
	}{
		{"colourless", "{3}", false},
		{"no mana cost", "", false},
		{"one colour", "{1}{G}", true},
		{"two colours", "{G}{W}", true},
		{"hybrid with generic", "{2/W}", true},
	} {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		attacker := pushMVCreature(g, me, "Shielded", "{1}{W}", ProtectionFromColorless)
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
	// An unknown source matches nothing, the weaker direction.
	q, _ := ParseProtectionQuality(ProtectionFromColorless)
	if q.Matches(nil) {
		t.Error("an unknown source matched protection from colorless")
	}
}

func TestOnlyOrColorlessOffersTheColourlessAnswer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var plain, wide uuid.UUID
	g.WithWriteLock(func() {
		plain = g.QueueColorChoiceThenForEffect(ColorPrompt{Chooser: me.ID, Options: []string{"C", "R"}})
		wide = g.QueueColorChoiceThenForEffect(ColorPrompt{Chooser: me.ID, OrColorless: true})
	})
	options := func(id uuid.UUID) []string {
		var out []string
		g.ReadSnapshot(func() {
			if _, c := g.findChoiceLocked(id); c != nil {
				out = append(out, c.ColorOptions...)
			}
		})
		return out
	}
	if got := options(plain); len(got) != 1 || got[0] != "R" {
		t.Errorf("a plain prompt's options = %v, want [R]: colorless is not a color (CR 105.4)", got)
	}
	if got := options(wide); len(got) != 6 || got[5] != ColorlessChoice {
		t.Errorf("an OrColorless prompt's options = %v, want the five colours then C", got)
	}
	if err := g.ResolveColorChoice(plain, me.ID, "C"); err == nil {
		t.Error("a plain prompt accepted the colourless answer")
	}
	if err := g.ResolveColorChoice(plain, me.ID, "R"); err != nil {
		t.Fatalf("answering the plain prompt: %v", err)
	}
	if err := g.ResolveColorChoice(wide, me.ID, "c"); err != nil {
		t.Errorf("an OrColorless prompt refused the colourless answer: %v", err)
	}
	if ColorName(ColorlessChoice) != "colorless" {
		t.Errorf("ColorName(C) = %q", ColorName(ColorlessChoice))
	}
}
