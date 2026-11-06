package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pump_x_purpose_test.go — #4976: a benefit until end of turn whose power
// or toughness is the announced X declares which side is X
// (TargetPurpose.PumpPowerIsX, PumpToughnessIsX). The registry refuses
// one that cannot be; coverage.TestPumpXDeclarationsMatchThePrintedCard
// holds the catalog's to the printed words.

func TestPumpXGuard(t *testing.T) {
	creature := func() *game.TargetSpec { return TargetCreature("target creature") }
	cases := []struct {
		name string
		e    game.TargetPurpose
		want string // "" = accepted
	}{
		{"+X/+X", PumpXUntilEOT(0), ""},
		{"+X/+X with keywords", PumpXUntilEOT(0, "hexproof", "indestructible"), ""},
		{"+X/+0", PowerXUntilEOT(0), ""},
		{"+X/+0 with haste", PowerXUntilEOT(0, "haste"), ""},
		{"+X/+1", game.TargetPurpose{UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{Toughness: 1}, PumpPowerIsX: true}, ""},
		{"on a harm", game.TargetPurpose{UntilEndOfTurn: game.TargetHarm, PumpPowerIsX: true}, "not a benefit"},
		{"with no direction", game.TargetPurpose{PumpToughnessIsX: true}, "not a benefit"},
		{"X beside a printed power", game.TargetPurpose{UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{Power: 2}, PumpPowerIsX: true}, "printed amount and X"},
		{"on a player", game.TargetPurpose{UntilEndOfTurn: game.TargetBenefit, PumpPowerIsX: true}, "cannot target a permanent"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			targets := creature()
			if c.name == "on a player" {
				targets = TargetPlayer("target player")
			}
			spec := Spec{Name: "Test Card", Targets: targets, Purpose: ForTargets(c.e)}
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				checkSpecPurposes(spec)
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

// TestTheIssuesCardsDeclareTheirX: the three cards #4976 names.
func TestTheIssuesCardsDeclareTheirX(t *testing.T) {
	entry := func(oracle string, mode int) game.TargetPurpose {
		t.Helper()
		s, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		p := s.Purpose
		if mode >= 0 {
			p = s.Modes.Options[mode].Purpose
		}
		got := p.Targets.List()
		if len(got) != 1 {
			t.Fatalf("%s declares %+v, want one entry", s.Name, got)
		}
		return got[0]
	}
	if e := entry("4247b667-a0a8-4995-97c7-622d20132f7d", -1); !e.PumpPowerIsX || !e.PumpToughnessIsX || e.Pump == nil || len(e.Pump.Keywords) != 2 {
		t.Errorf("Tyvar's Stand declares %+v, want +X/+X with hexproof and indestructible", e)
	}
	if e := entry("40601061-1b25-4db8-8b99-33324ce945cf", -1); !e.PumpPowerIsX || e.PumpToughnessIsX || e.Pump != nil {
		t.Errorf("Hatred declares %+v, want +X/+0", e)
	}
	if e := entry("25434ea3-bcfa-4dae-a16e-10ab87ce32af", 3); !e.PumpPowerIsX || e.PumpToughnessIsX || e.Pump == nil || e.Pump.Keywords[0] != "haste" {
		t.Errorf("Mishra's Command's last bullet declares %+v, want +X/+0 with haste", e)
	}
}
