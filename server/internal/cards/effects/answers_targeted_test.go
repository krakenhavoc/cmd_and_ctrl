package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answers_targeted_test.go — ADR 0009's amendment of 2026-10-30 (#4734),
// owner decision Q2 (a), slice 7: a targeted activated row may declare
// what it answers, optionally and with no ratchet, and smart auto-pass
// matches it by its kind (answers_now) instead of stopping at every key
// window through has_targets. The protective rows declare first.

// TestProtectiveTargetedRowsDeclare pins slice 7's first declarations:
// the six protective rows the amendment names, each with a target it
// must name, each declaring the union of what it gives its target (ADR
// 0142 sweep ruling 4).
func TestProtectiveTargetedRowsDeclare(t *testing.T) {
	for _, r := range []struct {
		card string
		want game.Answers
		why  string
	}{
		{"Mother of Runes", game.AnswerProtect, "protection from a color"},
		{"Liliana the Faultless", game.AnswerProtect, "hexproof"},
		{"Plaza of Heroes", game.AnswerProtect | game.AnswerSurvive, "hexproof and indestructible"},
		{"Skrelv, Defector Mite", game.AnswerProtect | game.AnswerCombatGrant, "hexproof from a color, toxic 1 and evasion"},
		{"Arwen, Mortal Queen", game.AnswerSurvive | game.AnswerPump | game.AnswerCombatGrant, "indestructible, a +1/+1 counter and a lifelink counter"},
		{"Zack Fair", game.AnswerSurvive | game.AnswerPump, "indestructible and Zack's +1/+1 counters"},
	} {
		t.Run(r.card, func(t *testing.T) {
			spec := specNamed(t, r.card)
			n := 0
			for _, ab := range spec.Activated {
				if ab.Targets == nil {
					continue
				}
				n++
				if hasUntargetedAnnouncement(activatedShapeOf(ab)) {
					t.Errorf("%q can be announced with no target; slice 7 is about rows that must name one", ab.Label)
				}
				if got := ab.Purpose.Answers; got != r.want {
					t.Errorf("%q declares %s, want %s (%s)", ab.Label, got, r.want, r.why)
				}
			}
			if n != 1 {
				t.Errorf("%s has %d targeted activated rows, want 1", r.card, n)
			}
		})
	}
}
