package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4977: a harm's "can't block this turn" rides the entry's pump as
// `cant_block`, and an entry without it sends no key.
func TestCantBlockRidesTheTargetEntry(t *testing.T) {
	p := game.Purpose{Targets: game.ForTargets(
		game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{CantBlock: true}},
	)}
	raw, err := json.Marshal(viewOfPurpose(p))
	if err != nil {
		t.Fatal(err)
	}
	const wantJSON = `{"targets":[{"slot":0,"until_end_of_turn":"harm","pump":{"cant_block":true}}]}`
	if string(raw) != wantJSON {
		t.Errorf("wire = %s, want %s", raw, wantJSON)
	}
	p = game.Purpose{Targets: game.ForTargets(
		game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{Power: -2, Toughness: -2}},
	)}
	raw, _ = json.Marshal(viewOfPurpose(p))
	if strings.Contains(string(raw), "cant_block") {
		t.Errorf("a shrink sends %s", raw)
	}
}
