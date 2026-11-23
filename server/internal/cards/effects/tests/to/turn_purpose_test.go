package teststo

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestExtraTurnCardsDeclareTheirTurns pins the shipped extra-turn cards
// the slice declares, spell by spell and row by row.
func TestExtraTurnCardsDeclareTheirTurns(t *testing.T) {
	// "Target player takes N extra turns": a target entry, because whoever
	// the move aims it at takes them.
	for name, n := range map[string]int{"Time Warp": 1, "Time Stretch": 2} {
		p := effectstest.SpecNamed(t, name).Purpose
		if p.ExtraTurns != 0 {
			t.Errorf("%s declares ExtraTurns %d on the card: the target takes the turns, not the caster", name, p.ExtraTurns)
		}
		list := p.Targets.List()
		if len(list) != 1 || list[0].Slot != 0 || list[0].ExtraTurns != n {
			t.Errorf("%s target entries = %+v, want slot 0 with ExtraTurns %d", name, list, n)
		}
	}
	// "Take an extra turn after this one": the caster's own.
	for _, name := range []string{"Temporal Manipulation", "Capture of Jingzhou", "Temporal Mastery", "Temporal Trespass"} {
		if p := effectstest.SpecNamed(t, name).Purpose; p.ExtraTurns != 1 || p.SkipsTurns != 0 {
			t.Errorf("%s purpose = %+v, want ExtraTurns 1", name, p)
		}
	}
	// And the activated rows.
	rows := []struct {
		card  string
		index int
		extra int
		skips int
	}{
		{"Magosi, the Waterveil", 0, 0, 1},
		{"Magosi, the Waterveil", 1, 1, 0},
		{"Magistrate's Scepter", 1, 1, 0},
		{"Avatar Kuruk", 0, 1, 0},
		{"Timestream Navigator", 0, 1, 0},
	}
	for _, r := range rows {
		s := effectstest.SpecNamed(t, r.card)
		if r.index >= len(s.Activated) {
			t.Fatalf("%s has no activated row %d", r.card, r.index)
		}
		if p := s.Activated[r.index].Purpose; p.ExtraTurns != r.extra || p.SkipsTurns != r.skips {
			t.Errorf("%s row %d purpose = %+v, want ExtraTurns %d SkipsTurns %d", r.card, r.index, p, r.extra, r.skips)
		}
	}
}

// TestAnExtraTurnThatLosesTheGameIsNotAGain: Final Fortune, Last Chance
// and Warrior's Oath take an extra turn and lose the game at its end
// step. Declared as a gain the bot would cast them, so they declare none.
func TestAnExtraTurnThatLosesTheGameIsNotAGain(t *testing.T) {
	for _, name := range []string{"Final Fortune", "Last Chance", "Warrior's Oath"} {
		if p := effectstest.SpecNamed(t, name).Purpose; p.ExtraTurns != 0 {
			t.Errorf("%s declares ExtraTurns %d, but its turn ends in a loss", name, p.ExtraTurns)
		}
	}
}

// TestLethalVapors declares no turn cost: its skip falls on whoever
// activates the any-player row, and a bot must never activate another
// player's copy (ADR 0106 decision 2).
func TestLethalVaporsDeclaresNoSkip(t *testing.T) {
	s := effectstest.SpecNamed(t, "Lethal Vapors")
	if s.Purpose.SkipsTurns != 0 {
		t.Errorf("Lethal Vapors declares SkipsTurns on the card: %+v", s.Purpose)
	}
	for i, a := range s.Activated {
		if a.Purpose.SkipsTurns != 0 || a.Purpose.ExtraTurns != 0 {
			t.Errorf("Lethal Vapors row %d declares a turn amount: %+v", i, a.Purpose)
		}
	}
}

func TestTurnPurposeGuard(t *testing.T) {
	for _, c := range []struct {
		name string
		slot effects.PurposeSlot
		p    game.Purpose
		want string
	}{
		{"extra turn on the card", effects.PurposeOnCard, game.Purpose{ExtraTurns: 1}, ""},
		{"skip on an activated row", effects.PurposeOnActivated, game.Purpose{SkipsTurns: 1}, ""},
		{"extra turn on a triggered row", effects.PurposeOnTriggered, game.Purpose{ExtraTurns: 1}, ""},
		{"extra turn on an any-player row", effects.PurposeOnAnyPlayerActivated, game.Purpose{ExtraTurns: 1}, "any-player"},
		{"skip on an any-player row", effects.PurposeOnAnyPlayerActivated, game.Purpose{SkipsTurns: 1}, "any-player"},
		{"negative extra turns", effects.PurposeOnCard, game.Purpose{ExtraTurns: -1}, "negative amount"},
		{"negative skips", effects.PurposeOnCard, game.Purpose{SkipsTurns: -1}, "negative amount"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				effects.CheckPurpose("Test Card", "card", c.slot, c.p)
			}()
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}
