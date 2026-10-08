package heuristic

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// keyword_counter_redundant_test.go — #2538: a second exalted counter
// is a second instance of exalted (the Emissary of Soulfire ruling of
// 2024-06-07), so putting one on a creature that already has exalted is
// not the redundant activation a second indestructible counter is.
func TestAnExaltedCounterIsNeverRedundant(t *testing.T) {
	row := func(label string, abilities ...string) *protocol.CardView {
		return &protocol.CardView{
			Abilities:          abilities,
			ActivatedAbilities: []protocol.ActivatedAbilityView{{Index: 0, Ref: "own:0", Label: label}},
		}
	}
	exalted := row("Pay {E}{E}: Put an exalted counter on Emissary of Soulfire.", "exalted")
	if redundantKeywordCounter(exalted, 0) {
		t.Error("an exalted counter on a creature with exalted was priced as redundant")
	}
	indestructible := row("Put an indestructible counter on Solphim, Mayhem Dominus.", "indestructible")
	if !redundantKeywordCounter(indestructible, 0) {
		t.Error("the control case: a second indestructible counter must still be redundant")
	}
}
