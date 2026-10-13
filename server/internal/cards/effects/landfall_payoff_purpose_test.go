package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// landfall_payoff_purpose_test.go — #3072: a landfall row declares what
// it pays for each land that enters under its controller's control.

func TestLandfallPayoffDeclarations(t *testing.T) {
	cases := []struct {
		oracle, name string
		want         game.LandfallPayoff
	}{
		{"0715e860-3b3b-4331-9718-207973e94fee", "Tatyova, Benthic Druid", game.LandfallPayoff{Draws: 1, LifeGain: 1}},
		{"8ad91f64-ccab-4edc-bd54-b2ee9267d614", "Lotus Cobra", game.LandfallPayoff{Mana: 1}},
		{"e6555cca-e608-4c87-b835-9cd47fd1f543", "Tireless Tracker", game.LandfallPayoff{Tokens: 1}},
		{"ab8d5f5c-1976-4f77-8ed2-8d28ee666741", "Tireless Provisioner", game.LandfallPayoff{Tokens: 1}},
		{"1be56a3d-a6c0-4b65-ae71-3d90ceefc6c0", "Sporemound", game.LandfallPayoff{Tokens: 1}},
		{"2d3e6549-6cc6-434f-a189-ba3b55e64c34", "Rampaging Baloths", game.LandfallPayoff{Tokens: 1}},
		{"46779609-4fa7-4fd2-b5b4-7d4d749339e6", "Courser of Kruphix", game.LandfallPayoff{LifeGain: 1}},
		{"579225b5-50e2-4891-8871-0bf6dbc07e33", "Omnath, Locus of the Void", game.LandfallPayoff{Mana: 2}},
		{"cb0eb84d-b41b-4147-aa7c-089b7fabd835", "Undergrowth Champion", game.LandfallPayoff{Counters: 1}},
		{"be6cab52-b3ae-46a2-a581-2468a49bff25", "Architect of the Untamed", game.LandfallPayoff{Energy: 1}},
		{"65385311-1158-4e94-892a-683997706ca8", "Sabotender", game.LandfallPayoff{DamageEachOpponent: 1}},
	}
	for _, c := range cases {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		var got []*game.LandfallPayoff
		for _, tr := range spec.Triggered {
			if tr.Purpose.LandfallPayoff != nil {
				got = append(got, tr.Purpose.LandfallPayoff)
			}
		}
		if len(got) != 1 {
			t.Errorf("%s declares %d landfall payoffs, want exactly one, on its landfall trigger", c.name, len(got))
			continue
		}
		if *got[0] != c.want {
			t.Errorf("%s declares %+v, want %+v", c.name, *got[0], c.want)
		}
	}
}

func TestLandfallPayoffGuard(t *testing.T) {
	ok := &game.LandfallPayoff{Draws: 1, LifeGain: 1}
	cases := []struct {
		name string
		slot purposeSlot
		p    game.Purpose
		want string // "" = accepted
	}{
		{"a triggered row", purposeOnTriggered, game.Purpose{LandfallPayoff: ok}, ""},
		{"on the card", purposeOnCard, game.Purpose{LandfallPayoff: ok}, "off a triggered row"},
		{"on an activated row", purposeOnActivated, game.Purpose{LandfallPayoff: ok}, "off a triggered row"},
		{"on a mode", purposeOnMode, game.Purpose{LandfallPayoff: ok}, "off a triggered row"},
		{"pays nothing", purposeOnTriggered, game.Purpose{LandfallPayoff: &game.LandfallPayoff{}}, "pays nothing"},
		{"negative", purposeOnTriggered, game.Purpose{LandfallPayoff: &game.LandfallPayoff{Tokens: 1, Mana: -1}}, "negative landfall amount"},
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

// LandfallPaying is Landfall with the payoff on the row.
func TestLandfallPayingDeclaresTheRow(t *testing.T) {
	row := LandfallPaying("x", game.LandfallPayoff{Tokens: 1}, Do())
	plain := Landfall("x", Do())
	if row.Purpose.LandfallPayoff == nil || *row.Purpose.LandfallPayoff != (game.LandfallPayoff{Tokens: 1}) {
		t.Errorf("LandfallPaying's purpose = %+v, want one token per land", row.Purpose)
	}
	if len(row.Watches) != len(plain.Watches) || row.Watches[0] != plain.Watches[0] {
		t.Errorf("LandfallPaying watches %v, want Landfall's %v", row.Watches, plain.Watches)
	}
}
