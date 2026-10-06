package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_payoff_purpose_test.go — ADR 0126's amendment of 2026-10-06:
// a triggered row that pays out whenever its controller discards a card
// declares which cards it pays on and what it pays for each.

func TestDiscardPayoffDeclarations(t *testing.T) {
	cases := []struct {
		oracle, name string
		want         game.DiscardPayoff
	}{
		{"5182de2d-aceb-450e-bd20-8bc7db124334", "Mary Read and Anne Bonny",
			game.DiscardPayoff{Types: []string{"island", "pirate", "vehicle"}, Tokens: 1}},
		{"e349be42-5f14-44a9-9608-281985c10e2d", "Marauding Mako", game.DiscardPayoff{Any: true, Counters: 1}},
		{"3a46d85b-ce1a-4842-a342-92a5bddb1053", "Scrounging Skyray", game.DiscardPayoff{Any: true, Counters: 1}},
		{"900b9409-9c16-414d-8674-2ea42c2415a1", "Magmakin Artillerist", game.DiscardPayoff{Any: true, DamageEachOpponent: 1}},
		{"64ad5657-78e9-4f34-8877-18c4f51fff9a", "Glint-Horn Buccaneer", game.DiscardPayoff{Any: true, DamageEachOpponent: 1}},
		{"db266661-f783-4907-9e52-6963eec05431", "Hashaton, Scarab's Fist", game.DiscardPayoff{Types: []string{"creature"}, Tokens: 1}},
	}
	for _, c := range cases {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		var got []*game.DiscardPayoff
		for _, tr := range spec.Triggered {
			if tr.Purpose.DiscardPayoff != nil {
				got = append(got, tr.Purpose.DiscardPayoff)
			}
		}
		if len(got) != 1 {
			t.Errorf("%s declares %d discard payoffs, want exactly one, on its discard trigger", c.name, len(got))
			continue
		}
		d := got[0]
		if d.Any != c.want.Any || strings.Join(d.Types, ",") != strings.Join(c.want.Types, ",") ||
			d.Tokens != c.want.Tokens || d.Counters != c.want.Counters || d.DamageEachOpponent != c.want.DamageEachOpponent {
			t.Errorf("%s declares %+v, want %+v", c.name, *d, c.want)
		}
	}
}

func TestDiscardPayoffGuard(t *testing.T) {
	ok := &game.DiscardPayoff{Any: true, Counters: 1}
	cases := []struct {
		name string
		slot purposeSlot
		p    game.Purpose
		want string // "" = accepted
	}{
		{"a triggered row", purposeOnTriggered, game.Purpose{DiscardPayoff: ok}, ""},
		{"by types", purposeOnTriggered, game.Purpose{DiscardPayoff: &game.DiscardPayoff{Types: []string{"island"}, Tokens: 1}}, ""},
		{"on the card", purposeOnCard, game.Purpose{DiscardPayoff: ok}, "off a triggered row"},
		{"on an activated row", purposeOnActivated, game.Purpose{DiscardPayoff: ok}, "off a triggered row"},
		{"on a mode", purposeOnMode, game.Purpose{DiscardPayoff: ok}, "off a triggered row"},
		{"no cards", purposeOnTriggered, game.Purpose{DiscardPayoff: &game.DiscardPayoff{Tokens: 1}}, "pays on no discarded card"},
		{"both ways", purposeOnTriggered, game.Purpose{DiscardPayoff: &game.DiscardPayoff{Any: true, Types: []string{"island"}, Tokens: 1}}, "both as Any and by Types"},
		{"a capital", purposeOnTriggered, game.Purpose{DiscardPayoff: &game.DiscardPayoff{Types: []string{"Island"}, Tokens: 1}}, "not one lowercase word"},
		{"two words", purposeOnTriggered, game.Purpose{DiscardPayoff: &game.DiscardPayoff{Types: []string{"basic land"}, Tokens: 1}}, "not one lowercase word"},
		{"pays nothing", purposeOnTriggered, game.Purpose{DiscardPayoff: &game.DiscardPayoff{Any: true}}, "pays nothing"},
		{"negative", purposeOnTriggered, game.Purpose{DiscardPayoff: &game.DiscardPayoff{Any: true, Tokens: 2, Counters: -1}}, "negative amount"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				checkPurpose("Test Card", "triggered ability 0", c.slot, c.p)
			}()
			switch {
			case c.want == "" && msg != "":
				t.Errorf("refused: %s", msg)
			case c.want != "" && !strings.Contains(msg, c.want):
				t.Errorf("got %q, want a refusal containing %q", msg, c.want)
			}
		})
	}
}
