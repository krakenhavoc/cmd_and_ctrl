package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// death_payoff_purpose_test.go — #4990: a death payoff declares its
// purpose on its dies row (ADR 0126 §6), so the bot keeps it in play
// and sacrifices around it instead of playing it as a plain body.
//
// internal/decks' TestCuratedDeckPurposes holds a curated deck's
// payoffs to a declaration, but only the cards a curated deck carries.
// This pins the payoffs #4990 found, whatever deck carries them.

func TestDeathPayoffDeclarations(t *testing.T) {
	for _, c := range []struct {
		oracle, name string
		want         game.Purpose // the dies row's whole purpose
	}{
		{"3ee78cfc-0e9e-4737-a7e2-b42f94228040", "Cruel Celebrant", game.Purpose{DeathPayoff: true}},
		{"391978f6-0bbc-41e8-9246-f7d0e21c7900", "Nadier's Nightblade", game.Purpose{DeathPayoff: true}},
		{"a4d33e21-d3fe-4d5b-903c-b75e859d6f7f", "Elas il-Kor, Sadistic Pilgrim", game.Purpose{DeathPayoff: true}},
		{"322f44f0-e6da-4ee0-b474-e7d5e9a461c5", "Morbid Opportunist", game.Purpose{Draws: 1, DeathPayoff: true}},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		var got []game.Purpose
		for _, tr := range spec.Triggered {
			if tr.Purpose.DeathPayoff {
				got = append(got, tr.Purpose)
			}
		}
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%s declares %+v on its death-payoff rows, want exactly one: %+v", c.name, got, c.want)
		}
	}
}
