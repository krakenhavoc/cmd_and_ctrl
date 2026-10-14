package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// creature_enter_payoff_test.go — #3165: a creature-enter row declares
// what it pays for each creature, and a token maker declares how many of
// its tokens are creatures.

func TestCreatureEnterPayoffDeclarations(t *testing.T) {
	cases := []struct {
		oracle, name string
		want         game.CreatureEnterPayoff
	}{
		{"4fdbbec2-e921-4b63-958d-f9ba1e417197", "Purphoros, God of the Forge", game.CreatureEnterPayoff{DamageEachOpponent: 2}},
		{"9242cd3e-1a71-4700-8182-9c1005616033", "Impact Tremors", game.CreatureEnterPayoff{DamageEachOpponent: 1}},
		{"34edbf1b-0826-49f1-a04d-15cc8c543b22", "Witty Roastmaster", game.CreatureEnterPayoff{DamageEachOpponent: 1}},
		{"90fdfba8-f29e-44f9-91d2-7bf3c458a9c1", "Molten Gatekeeper", game.CreatureEnterPayoff{DamageEachOpponent: 1}},
		{"e1a2bd31-b800-47cd-8866-1b849a601b09", "Agate Instigator", game.CreatureEnterPayoff{DamageEachOpponent: 1}},
	}
	for _, c := range cases {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		var got []*game.CreatureEnterPayoff
		for _, tr := range spec.Triggered {
			if tr.Purpose.CreatureEnterPayoff != nil {
				got = append(got, tr.Purpose.CreatureEnterPayoff)
			}
		}
		if len(got) != 1 || *got[0] != c.want {
			t.Errorf("%s declares %v, want exactly %+v", c.name, got, c.want)
		}
	}
}

func TestCreatureTokenDeclarations(t *testing.T) {
	for _, c := range []struct {
		oracle, name string
		tokens       int
	}{
		{"a6450b8e-eb18-431c-9eb7-7daf107978b2", "Hordeling Outburst", 3},
		{"d0d2c45b-b6e3-4999-bdab-976e8f0d6617", "Dragon Fodder", 2},
		{"9cfe86ae-eebe-44aa-a956-4b3e9e621105", "Krenko's Command", 2},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if spec.Purpose.Tokens != c.tokens || spec.Purpose.CreatureTokens != c.tokens {
			t.Errorf("%s declares %+v, want %d creature tokens", c.name, spec.Purpose, c.tokens)
		}
	}
	krenko, ok := Lookup("68418069-f615-40ef-ae0d-764192acae00")
	if !ok || len(krenko.Activated) != 1 || krenko.Activated[0].Purpose.TokensPerControlled != "goblin" {
		t.Errorf("Krenko, Mob Boss does not declare goblin tokens per controlled Goblin")
	}
}

func TestCreatureEnterPayoffGuard(t *testing.T) {
	ok := &game.CreatureEnterPayoff{DamageEachOpponent: 1}
	cases := []struct {
		name string
		slot purposeSlot
		p    game.Purpose
		want string // "" = accepted
	}{
		{"a triggered row", purposeOnTriggered, game.Purpose{CreatureEnterPayoff: ok}, ""},
		{"on the card", purposeOnCard, game.Purpose{CreatureEnterPayoff: ok}, "off a triggered row"},
		{"on an activated row", purposeOnActivated, game.Purpose{CreatureEnterPayoff: ok}, "off a triggered row"},
		{"pays nothing", purposeOnTriggered, game.Purpose{CreatureEnterPayoff: &game.CreatureEnterPayoff{}}, "pays nothing"},
		{"creature tokens within tokens", purposeOnCard, game.Purpose{Tokens: 2, CreatureTokens: 2}, ""},
		{"more creature tokens than tokens", purposeOnCard, game.Purpose{Tokens: 1, CreatureTokens: 2}, "more CreatureTokens"},
		{"per controlled on an activated row", purposeOnActivated, game.Purpose{TokensPerControlled: "goblin"}, ""},
		{"per controlled on the card", purposeOnCard, game.Purpose{TokensPerControlled: "goblin"}, "off a triggered or activated row"},
		{"per controlled not lowercase", purposeOnActivated, game.Purpose{TokensPerControlled: "Goblin"}, "not one lowercase word"},
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
