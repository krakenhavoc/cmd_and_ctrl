package heuristic

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// protection_exchange_test.go — #662, the policy half.
//
// Block LEGALITY is the enumerator's business and stays there
// (game.BlockPairRefusalLocked, ADR 0045 §3). What the policy has to
// know is the DAMAGE rule: an exchange with a creature protected from
// your creature's quality is not an exchange, it is a free block.
//
// The parse comes from the server on CardView.Protection. A policy
// package may not import internal/game (ADR 0033 §3), so a prefix
// match on the raw token is the only alternative and it throws away
// exactly the half that matters — which quality.

func TestKillsRespectsProtection(t *testing.T) {
	creature := func(colors []string, typeLine string, power, toughness int, pro ...protocol.ProtectionView) *protocol.CardView {
		return &protocol.CardView{
			TypeLine:   typeLine,
			Colors:     colors,
			Power:      power,
			Toughness:  toughness,
			Protection: pro,
		}
	}
	proRed := protocol.ProtectionView{Printed: "red", Kind: "color", Value: "R"}
	proDemons := protocol.ProtectionView{Printed: "Demons", Kind: "subtype", Value: "Demon"}
	proEverything := protocol.ProtectionView{Printed: "everything", Kind: "everything"}

	for _, tc := range []struct {
		name string
		a    *protocol.CardView
		b    *protocol.CardView
		want bool
	}{
		{
			name: "a red 3/3 does not kill a pro-red 2/2",
			a:    creature([]string{"R"}, "Creature — Goblin", 3, 3),
			b:    creature([]string{"W"}, "Creature — Kor", 2, 2, proRed),
			want: false,
		},
		{
			name: "a white 3/3 kills the same pro-red 2/2",
			a:    creature([]string{"W"}, "Creature — Human", 3, 3),
			b:    creature([]string{"W"}, "Creature — Kor", 2, 2, proRed),
			want: true,
		},
		{
			name: "a Demon does not kill a creature with protection from Demons",
			a:    creature([]string{"B"}, "Creature — Demon", 6, 6),
			b:    creature([]string{"W"}, "Creature — Angel", 5, 5, proDemons),
			want: false,
		},
		{
			name: "a Demonlord is not a Demon",
			a:    creature([]string{"B"}, "Creature — Demonlord", 6, 6),
			b:    creature([]string{"W"}, "Creature — Angel", 5, 5, proDemons),
			want: true,
		},
		{
			name: "protection from everything is not killed by anything",
			a:    creature(nil, "Artifact Creature — Golem", 9, 9),
			b:    creature([]string{"W"}, "Creature — Hydra", 1, 1, proEverything),
			want: false,
		},
		{
			// CR 615.1: prevented damage is never dealt, so deathtouch
			// has nothing to make lethal.
			name: "deathtouch does not get round protection",
			a: &protocol.CardView{
				TypeLine: "Creature — Snake", Colors: []string{"R"},
				Power: 1, Toughness: 1, Abilities: []string{"deathtouch"},
			},
			b:    creature([]string{"W"}, "Creature — Kor", 2, 2, proRed),
			want: false,
		},
		{
			name: "no protection at all is the ordinary answer",
			a:    creature([]string{"R"}, "Creature — Goblin", 3, 3),
			b:    creature([]string{"G"}, "Creature — Bear", 2, 2),
			want: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := kills(tc.a, tc.b); got != tc.want {
				t.Errorf("kills = %v, want %v", got, tc.want)
			}
		})
	}
}
