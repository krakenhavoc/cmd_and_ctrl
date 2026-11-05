package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lifegain_payoff_test.go — #4979: a "whenever you gain life" row
// declares what it pays for each life-gain event.

func TestLifeGainPayoffDeclarations(t *testing.T) {
	cases := []struct {
		oracle, name string
		want         game.LifeGainPayoff
	}{
		{"63e596a2-9126-4af6-8782-c38687d664ad", "Heliod, Sun-Crowned", game.LifeGainPayoff{Counters: 1}},
		{"95e94dea-5ac0-4d6f-adec-ca147aee861f", "Ajani's Pridemate", game.LifeGainPayoff{Counters: 1}},
	}
	for _, c := range cases {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		var got []*game.LifeGainPayoff
		for _, tr := range spec.Triggered {
			if tr.Purpose.LifeGainPayoff != nil {
				got = append(got, tr.Purpose.LifeGainPayoff)
			}
		}
		if len(got) != 1 || *got[0] != c.want {
			t.Errorf("%s declares %v, want exactly %+v", c.name, got, c.want)
		}
	}
}

func TestLifeGainPayoffGuard(t *testing.T) {
	ok := &game.LifeGainPayoff{Counters: 1}
	cases := []struct {
		name string
		slot purposeSlot
		p    game.Purpose
		want string // "" = accepted
	}{
		{"a triggered row", purposeOnTriggered, game.Purpose{LifeGainPayoff: ok}, ""},
		{"on the card", purposeOnCard, game.Purpose{LifeGainPayoff: ok}, "off a triggered row"},
		{"on an activated row", purposeOnActivated, game.Purpose{LifeGainPayoff: ok}, "off a triggered row"},
		{"pays nothing", purposeOnTriggered, game.Purpose{LifeGainPayoff: &game.LifeGainPayoff{}}, "pays nothing"},
		{"a negative amount", purposeOnTriggered, game.Purpose{LifeGainPayoff: &game.LifeGainPayoff{Counters: -1}}, "negative"},
		{"gains life", purposeOnTriggered, game.Purpose{LifeGainPayoff: &game.LifeGainPayoff{LifeGain: 1}}, "trigger itself"},
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
				checkPurpose("Test Card", "ability 0", c.slot, c.p)
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
