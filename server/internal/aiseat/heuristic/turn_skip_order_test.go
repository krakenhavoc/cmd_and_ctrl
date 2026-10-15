package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// turn_skip_order_test.go is ADR 0188 Decision 12 (#3561): the bot uses
// its own "skip your next turn" on an extra turn another effect would
// skip anyway, the recommended answer, whichever order the answers are
// listed in.
func TestTheBotUsesItsOwnSkipOnTheExtraTurn(t *testing.T) {
	const id = "00000000-0000-4000-8000-000000003561"
	use := choiceMove(t, 0, id, "use my skip", map[string]any{"apply": true})
	let := choiceMove(t, 0, id, "let it skip", map[string]any{"apply": false})
	for name, moves := range map[string][]legal.Move{"use first": {use, let}, "let first": {let, use}} {
		t.Run(name, func(t *testing.T) {
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withTurn(3, 0, "cleanup"),
				withChoice(protocol.PendingChoiceView{
					ID: id, Kind: "turn_skip_order", Chooser: seatID(0).String(), FromPlayer: seatID(0).String(), Count: 1,
					TurnSkipOrder: &protocol.TurnSkipOrderView{Others: []string{"Trouble in Pairs"}},
				}))
			in := input(0, v, moves...)
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != "use my skip" {
				t.Errorf("chose %q, want use my skip", got)
			}
		})
	}
}
