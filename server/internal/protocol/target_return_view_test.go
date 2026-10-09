package protocol

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// target_return_view_test.go — #2679: what a removal hands its
// target's controller rides the target entry, and is absent when
// nothing comes back.
func TestViewOfPurposeProjectsTargetReturns(t *testing.T) {
	p := game.Purpose{Targets: game.ForTargets(
		game.TargetPurpose{Slot: 0, Returns: game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3}},
		game.TargetPurpose{Slot: 1, Returns: game.TargetReturn{LifeEqualToPower: true, Lands: 1, LandsUntapped: 1}},
		game.TargetPurpose{Slot: 2, Damage: 2},
	)}
	raw, err := json.Marshal(viewOfPurpose(p))
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"targets":[` +
		`{"slot":0,"returns":{"creature_tokens":1,"token_power":3,"token_toughness":3}},` +
		`{"slot":1,"returns":{"life_equal_to_power":true,"lands":1,"lands_untapped":1}},` +
		`{"slot":2,"damage":2}]}`
	if string(raw) != want {
		t.Errorf("wire = %s, want %s", raw, want)
	}
}
