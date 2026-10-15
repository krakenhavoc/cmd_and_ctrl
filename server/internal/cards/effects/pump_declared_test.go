package effects

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pump_declared_test.go — #2939 gap 2: a plain self pump that adds
// toughness declares the amounts it adds, so the bot can work out how
// many activations keep the creature alive (ADR 0142 decision 6).

// manaSelfPump is "{R}: This creature gets +0/+1 until end of turn.": a
// mana-only cost and a flat +N/+M, which is all the declaration covers.
var manaSelfPump = regexp.MustCompile(`^(?:\{[^}]+\})+: This creature gets \+(\d+)/\+(\d+) until end of turn\.?$`)

func TestSelfPumpRowsDeclareTheirPump(t *testing.T) {
	found := 0
	for _, spec := range All() {
		for _, row := range spec.Activated {
			m := manaSelfPump.FindStringSubmatch(row.Label)
			if m == nil || !row.Purpose.Answers.Has(game.AnswerPump) || row.AnyPlayer {
				continue
			}
			power, _ := strconv.Atoi(m[1])
			toughness, _ := strconv.Atoi(m[2])
			if toughness == 0 {
				continue
			}
			found++
			p := row.Purpose.Pump
			if p == nil || p.Power != power || p.Toughness != toughness {
				t.Errorf("%s: %q declares Pump %+v, want +%d/+%d", spec.Name, row.Label, p, power, toughness)
			}
		}
	}
	if found < 7 {
		t.Errorf("found %d plain self-pump rows, want at least the seven declared in #2939 gap 2", found)
	}
}
