package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrounging Bandar — Creature — Cat Monkey {1}{G}, 0/0 (Jumpstart 2022,
// #5948):
//
//	"This creature enters with two +1/+1 counters on it.
//	 At the beginning of your upkeep, you may move any number of +1/+1
//	 counters from this creature onto another target creature."
//
// The counters are a replacement on its entry (CR 614.1c), so it never sits
// at 0/0. The upkeep trigger picks "another target creature" (any player's) as
// it goes on the stack; how many counters move is asked as it resolves, zero
// declining (CR 608.2c). Moving takes them off this creature and puts them on
// the target (CR 122.5), and only while both are still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e1eff5a-77e8-4ea2-96b8-547504f83b45",
		Name:         "Scrounging Bandar",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Scrounging Bandar: enters with two +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Scrounging Bandar — you may move any number of +1/+1 counters onto another target creature", gapB153BandarMove),
				Another(TargetCreature("another target creature"))),
		},
	})
}
