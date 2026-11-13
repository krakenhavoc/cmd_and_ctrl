package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spike Soldier — Creature — Spike Soldier {2}{G}{G}, 0/0 (Stronghold,
// #5627):
//
//	"This creature enters with three +1/+1 counters on it.
//	 {2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter
//	 on target creature.
//	 Remove a +1/+1 counter from this creature: This creature gets +2/+2
//	 until end of turn."
//
// The cycle's shared entry replacement (CR 614.1c) and counter move
// (spike_counters.go). The pump trades a permanent +1/+1 for +2/+2 until
// the cleanup step (CR 514.2) through the shared thisGetsUntilEndOfTurn
// body; its only cost is the counter (CR 118.3), so it can be activated
// once per counter at instant speed.
//
// No simplification.
func init() {
	const name = "Spike Soldier"
	Register(Spec{
		OracleID:     "a34d1845-e02f-4795-b445-d73366c2baea",
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{spikeEntersWith(name, "three", 3)},
		Activated: []ActivatedAbility{
			spikeMovesACounter(),
			{
				Label:   "Remove a +1/+1 counter from this creature: This creature gets +2/+2 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
				Cost:    RemoveCountersFromThis(game.CounterPlusOne, 1),
				Effect:  thisGetsUntilEndOfTurn(2, 2, name+" — +2/+2 until end of turn"),
			},
		},
	})
}
