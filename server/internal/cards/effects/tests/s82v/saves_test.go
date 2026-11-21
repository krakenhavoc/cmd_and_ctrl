package testss82v

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saves_test.go — #2939 gap 3: Purpose.Saves says which of its
// controller's permanents a survive or protect row covers.

func TestSavesGuard(t *testing.T) {
	creatures := &game.Saves{Matches: game.SweepCreatures}
	for _, c := range []struct {
		name string
		slot effects.PurposeSlot
		p    game.Purpose
		want string
	}{
		{"survive on an activated row", effects.PurposeOnActivated, game.Purpose{Answers: game.AnswerSurvive, Saves: creatures}, ""},
		{"protect on an activated row", effects.PurposeOnActivated, game.Purpose{Answers: game.AnswerProtect, Saves: creatures}, ""},
		{"on the card", effects.PurposeOnCard, game.Purpose{Saves: creatures}, "Saves off an activated row"},
		{"on a triggered row", effects.PurposeOnTriggered, game.Purpose{Saves: creatures}, "Saves off an activated row"},
		{"with neither answer", effects.PurposeOnActivated, game.Purpose{Answers: game.AnswerPump, Saves: creatures}, "neither survive nor protect"},
		{"an unknown class", effects.PurposeOnActivated, game.Purpose{Answers: game.AnswerSurvive, Saves: &game.Saves{Matches: "elves"}}, "unknown class"},
	} {
		t.Run(c.name, func(t *testing.T) {
			msg := effectstest.Refusal(func() { effects.CheckPurpose("Test Card", "ability 0", c.slot, c.p) })
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}

// The group-indestructible creatures declare what they cover, so the bot
// can price the sacrifice against the permanents it saves.
func TestGroupSaversDeclareWhatTheyCover(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		partial      bool
	}{
		{"Selfless Spirit", "71d785a9-ddc8-472e-a778-a551b444a4bd", false},
		{"Dauntless Escort", "", false},
		{"Boromir, Warden of the Tower", "", false},
		{"Ainok Strike Leader", "", true},
		{"Bulwark Ox", "", true},
		{"Dusk Legion Sergeant", "", true},
	} {
		var row *game.Purpose
		for _, spec := range effects.All() {
			if spec.Name != c.name {
				continue
			}
			for i := range spec.Activated {
				if p := &spec.Activated[i].Purpose; p.Answers.HasAny(game.AnswerSurvive) && spec.Activated[i].Cost.SacrificeSelf {
					row = p
				}
			}
		}
		if row == nil {
			t.Errorf("%s: no survive row that sacrifices its source", c.name)
			continue
		}
		if row.Saves == nil || row.Saves.Matches != game.SweepCreatures || row.Saves.Partial != c.partial {
			t.Errorf("%s: declares Saves %+v, want creatures (partial %v)", c.name, row.Saves, c.partial)
		}
	}
}
