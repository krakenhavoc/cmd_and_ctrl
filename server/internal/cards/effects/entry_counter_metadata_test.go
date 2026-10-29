package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEntryCounterConstructorsDeclareOnlyUnconditionalAmounts(t *testing.T) {
	for _, r := range []game.ReplacementEffect{b10EntersWithCounters(game.CounterPlusOne, 4, "irrelevant label"), SelfEntersTappedWithCounters("charge", 4)} {
		if r.SelfEntryCounters == nil || r.SelfEntryCounters.Amount != 4 || r.SelfEntryCounters.AmountIsX {
			t.Fatalf("fixed constructor metadata %+v", r.SelfEntryCounters)
		}
	}
	if !XCounters(game.CounterPlusOne).CountIsX {
		t.Fatal("unconditional X lacks metadata")
	}
	for _, r := range []game.EntryCountersFromCast{XCountersIfKicked(game.CounterPlusOne), CountersPerKick(game.CounterPlusOne, 2), CountersIfKickedWith(game.CounterPlusOne, "{G}", 2), CountersPerDelved(game.CounterPlusOne, nil)} {
		if r.CountIsX {
			t.Fatal("conditional/payment-dependent amount claims to be unconditional X")
		}
	}
	r := b19EntersWithCountersCounted(game.CounterPlusOne, func(*game.Game, *game.Card) int { return 4 }, "irrelevant label")
	if r.SelfEntryCounters != nil {
		t.Fatal("board-dependent count declared as fixed")
	}
}
