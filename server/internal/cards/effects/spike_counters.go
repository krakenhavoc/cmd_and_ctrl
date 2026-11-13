package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spike_counters.go — the Spike cycle's shared pieces (Stronghold,
// #5627): each Spike is a 0/0 that enters with +1/+1 counters and prints
// the same counter move.

// spikeEntersWith is "This creature enters with n +1/+1 counters on it."
// for a Spike named name: a self-replacement on the entry (CR 614.1c), so
// the Spike is never a 0/0 on the battlefield, however it arrived. words
// is n as printed ("two"), for the replacement's header.
func spikeEntersWith(name, words string, n int) game.ReplacementEffect {
	r := b10EntersWithCounters(game.CounterPlusOne, n, name+" — enters with "+words+" +1/+1 counters")
	r.RowLabel = "Enters with " + strconv.Itoa(n) + " +1/+1 counters on it."
	return r
}

// spikeMovesACounter is the row every Spike prints: "{2}, Remove a +1/+1
// counter from this creature: Put a +1/+1 counter on target creature."
// The removal is a cost (CR 118.3), so a Spike with no counter left can't
// pay it, and paying its last one leaves a 0/0 that dies to the
// state-based check (CR 704.5f) with the ability still on the stack. The
// target may be the Spike itself. The counter goes on the target if it is
// still legal as the ability resolves (CR 608.2b).
func spikeMovesACounter() ActivatedAbility {
	return ActivatedAbility{
		Label:   "{2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on target creature.",
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Cost:    Plus(ManaCost("{2}"), RemoveCountersFromThis(game.CounterPlusOne, 1)),
		Targets: TargetCreature("target creature"),
		Effect:  PlusOneCounterOnTheTarget,
	}
}
