package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// may_no_gain_test.go is ADR 0197 §10 (#2731): an activated ability's
// resolution-time "you may" is priced as it resolves. When the ability
// itself says "yes" would not help on this board (no_gain: Arni
// Brokenbrow's new base power would be no higher), the bot declines;
// otherwise it says yes, as it does to any "you may" of its own.

const mayChoiceID = "00000000-0000-4000-8000-00000000a731"

func mayDecision(t *testing.T, noGain bool) string {
	t.Helper()
	view := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withChoice(protocol.PendingChoiceView{
			ID: mayChoiceID, Kind: "trigger_prompt", Chooser: seatID(0).String(),
			Reason: "Arni Brokenbrow — have its base power become 1 plus the greatest power among your other creatures?",
			NoGain: noGain,
		}))
	no, yes := false, true
	in := input(0, view,
		choiceMove(t, 0, mayChoiceID, "no", map[string]any{"apply": no}),
		choiceMove(t, 0, mayChoiceID, "yes", map[string]any{"apply": yes}),
	)
	return chose(t, in, decide(t, heuristic.New(), in))
}

func TestMayAbilityDeclinedWhenYesWouldNotHelp(t *testing.T) {
	if got := mayDecision(t, true); got != "no" {
		t.Errorf("no_gain: %q, want no", got)
	}
}

func TestMayAbilityAcceptedOtherwise(t *testing.T) {
	if got := mayDecision(t, false); got != "yes" {
		t.Errorf("no warning: %q, want yes", got)
	}
}
