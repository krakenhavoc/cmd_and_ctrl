package testsds

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEntryCounterConstructorsDeclareOnlyUnconditionalAmounts(t *testing.T) {
	for _, r := range []game.ReplacementEffect{effects.B10EntersWithCounters(game.CounterPlusOne, 4, "irrelevant label"), effects.SelfEntersTappedWithCounters("charge", 4)} {
		if r.SelfEntryCounters == nil || r.SelfEntryCounters.Amount != 4 || r.SelfEntryCounters.AmountIsX {
			t.Fatalf("fixed constructor metadata %+v", r.SelfEntryCounters)
		}
	}
	if !effects.XCounters(game.CounterPlusOne).CountIsX {
		t.Fatal("unconditional X lacks metadata")
	}
	for _, r := range []game.EntryCountersFromCast{effects.XCountersIfKicked(game.CounterPlusOne), effects.CountersPerKick(game.CounterPlusOne, 2), effects.CountersIfKickedWith(game.CounterPlusOne, "{G}", 2), effects.CountersPerDelved(game.CounterPlusOne, nil)} {
		if r.CountIsX {
			t.Fatal("conditional/payment-dependent amount claims to be unconditional X")
		}
	}
	r := effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(*game.Game, *game.Card) int { return 4 }, "irrelevant label")
	if r.SelfEntryCounters != nil {
		t.Fatal("board-dependent count declared as fixed")
	}
}
